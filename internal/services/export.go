package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"bilibili/pkg/bilibili"
	"bilibili/pkg/file"
	"bilibili/pkg/utils"
	"github.com/google/uuid"
)

type ExportService struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	exportDir string
	files     map[string]*ExportFile
	mu        sync.RWMutex
}

type ExportFile struct {
	FileID    string
	Filename  string
	FilePath  string
	Format    string
	CreatedAt time.Time
}

func NewExportService(ctx context.Context, exportDir string) *ExportService {
	serviceCtx, cancel := context.WithCancel(ctx)

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		utils.LogError("Failed to create export directory: " + err.Error())
	}

	es := &ExportService{
		ctx:       serviceCtx,
		cancel:    cancel,
		exportDir: exportDir,
		files:     make(map[string]*ExportFile),
	}

	es.wg.Add(1)
	go func() {
		defer es.wg.Done()
		es.cleanupWorker()
	}()

	return es
}

var exportBasenamePattern = regexp.MustCompile(`^[\p{L}\p{N}._ -]+$`)

func normalizeExportFormat(format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "excel", "xlsx":
		return "xlsx", nil
	case "csv":
		return "csv", nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func validateExportBasename(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "comments", nil
	}
	if filepath.IsAbs(name) || filepath.Base(name) != name ||
		strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return "", fmt.Errorf("invalid export filename")
	}
	if len([]rune(name)) > 80 {
		return "", fmt.Errorf("export filename is too long")
	}
	if !exportBasenamePattern.MatchString(name) {
		return "", fmt.Errorf("export filename contains unsupported characters")
	}
	return name, nil
}

func (es *ExportService) ExportComments(comments []bilibili.CommentData, format, customFilename string) (*ExportFile, error) {
	fileID := uuid.New().String()

	normalizedFormat, err := normalizeExportFormat(format)
	if err != nil {
		return nil, err
	}
	baseName, err := validateExportBasename(customFilename)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.%s", baseName, timestamp, normalizedFormat)

	absDir, err := filepath.Abs(es.exportDir)
	if err != nil {
		return nil, fmt.Errorf("resolve export directory: %w", err)
	}
	filePath := filepath.Join(absDir, filename)
	relPath, err := filepath.Rel(absDir, filePath)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("export path escapes export directory")
	}

	rows := es.PrepareCommentRows(comments)
	switch normalizedFormat {
	case "xlsx":
		err = file.WriteExcel(rows, filePath)
	case "csv":
		err = file.WriteCSV(rows, filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to export: %w", err)
	}

	exportFile := &ExportFile{
		FileID:    fileID,
		Filename:  filename,
		FilePath:  filePath,
		Format:    normalizedFormat,
		CreatedAt: time.Now(),
	}

	es.mu.Lock()
	es.files[fileID] = exportFile
	es.mu.Unlock()

	return exportFile, nil
}

func (es *ExportService) GetExportFile(fileID string) (*ExportFile, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	exportFile, exists := es.files[fileID]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", fileID)
	}

	return exportFile, nil
}

func (es *ExportService) PrepareCommentRows(comments []bilibili.CommentData) [][]string {
	rows := [][]string{
		{"层级", "评论ID", "用户ID", "用户名", "等级", "评论内容", "点赞数", "评论时间"},
	}

	for _, comment := range comments {
		es.addCommentRow(&rows, comment, 0)
	}
	return rows
}

func (es *ExportService) addCommentRow(rows *[][]string, comment bilibili.CommentData, level int) {
	timeStr := time.Unix(int64(comment.Ctime), 0).Format("2006-01-02 15:04:05")

	levelStr := "主评论"
	if level > 0 {
		levelStr = fmt.Sprintf("└ 回复 (L%d)", level)
	}

	row := []string{
		levelStr,
		strconv.FormatInt(comment.RPID, 10),
		strconv.FormatInt(comment.Mid, 10),
		comment.Member.Uname,
		strconv.Itoa(comment.Member.LevelInfo.CurrentLevel),
		comment.Content.Message,
		strconv.Itoa(comment.Like),
		timeStr,
	}
	*rows = append(*rows, row)

	for _, reply := range comment.Replies {
		es.addCommentRow(rows, reply, level+1)
	}
}

func (es *ExportService) cleanupWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-es.ctx.Done():
			utils.LogInfo("cleanupWorker stopped in ExportService")
			return
		case <-ticker.C:
			es.CleanOldFiles()
		}
	}
}

func (es *ExportService) CleanOldFiles() {
	es.mu.Lock()
	defer es.mu.Unlock()

	cutoff := time.Now().Add(-2 * time.Hour)
	for fileID, exportFile := range es.files {
		if exportFile.CreatedAt.Before(cutoff) {
			if err := os.Remove(exportFile.FilePath); err != nil && !os.IsNotExist(err) {
				utils.LogError("Failed to remove old export file: " + err.Error())
			}
			delete(es.files, fileID)
		}
	}
}

func (es *ExportService) Shutdown(ctx context.Context) error {
	utils.LogInfo("Shutting down ExportService...")
	es.cancel()

	done := make(chan struct{})
	go func() {
		es.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		utils.LogInfo("ExportService shutdown complete")
		return nil
	case <-ctx.Done():
		utils.LogError("ExportService shutdown timeout")
		return ctx.Err()
	}
}
