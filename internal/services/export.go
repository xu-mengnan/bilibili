package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"bilibili/pkg/bilibili"
	"bilibili/pkg/file"
	"github.com/google/uuid"
)

// ExportService 导出服务
type ExportService struct {
	exportDir string
	files     map[string]*ExportFile
	mu        sync.RWMutex
}

// ExportFile 导出文件信息
type ExportFile struct {
	FileID    string
	Filename  string
	FilePath  string
	Format    string
	CreatedAt time.Time
}

// NewExportService 创建导出服务
func NewExportService(exportDir string) *ExportService {
	// 确保导出目录存在
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		fmt.Printf("Failed to create export directory: %v\n", err)
	}

	es := &ExportService{
		exportDir: exportDir,
		files:     make(map[string]*ExportFile),
	}

	// 启动清理goroutine
	go es.cleanupWorker()

	return es
}

// ExportComments 导出评论
func (es *ExportService) ExportComments(comments []bilibili.CommentData, format, customFilename string) (*ExportFile, error) {
	fileID := uuid.New().String()

	// 生成文件名
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	var filename string
	if customFilename != "" {
		filename = fmt.Sprintf("%s_%s.%s", customFilename, timestamp, format)
	} else {
		filename = fmt.Sprintf("comments_%s.%s", timestamp, format)
	}

	filePath := filepath.Join(es.exportDir, filename)

	// 准备数据
	rows := es.PrepareCommentRows(comments)

	// 根据格式导出
	var err error
	switch format {
	case "excel", "xlsx":
		err = file.WriteExcel(rows, filePath)
		format = "xlsx"
	case "csv":
		err = file.WriteCSV(rows, filePath)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to export: %v", err)
	}

	// 保存文件信息
	exportFile := &ExportFile{
		FileID:    fileID,
		Filename:  filename,
		FilePath:  filePath,
		Format:    format,
		CreatedAt: time.Now(),
	}

	es.mu.Lock()
	es.files[fileID] = exportFile
	es.mu.Unlock()

	return exportFile, nil
}

// GetExportFile 获取导出文件信息
func (es *ExportService) GetExportFile(fileID string) (*ExportFile, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	exportFile, exists := es.files[fileID]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", fileID)
	}

	return exportFile, nil
}

// PrepareCommentRows 准备评论数据行（包含子评论）
func (es *ExportService) PrepareCommentRows(comments []bilibili.CommentData) [][]string {
	// 表头
	rows := [][]string{
		{"层级", "评论ID", "用户ID", "用户名", "等级", "评论内容", "点赞数", "评论时间"},
	}

	// 递归添加评论数据
	for _, comment := range comments {
		es.addCommentRow(&rows, comment, 0)
	}

	return rows
}

// addCommentRow 递归添加评论行
func (es *ExportService) addCommentRow(rows *[][]string, comment bilibili.CommentData, level int) {
	// 格式化时间
	timeStr := time.Unix(int64(comment.Ctime), 0).Format("2006-01-02 15:04:05")

	// 层级标识
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

	// 递归处理子评论
	if len(comment.Replies) > 0 {
		for _, reply := range comment.Replies {
			es.addCommentRow(rows, reply, level+1)
		}
	}
}

// cleanupWorker 定期清理旧文件（2小时前）
func (es *ExportService) cleanupWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		es.CleanOldFiles()
	}
}

// CleanOldFiles 清理旧文件
func (es *ExportService) CleanOldFiles() {
	es.mu.Lock()
	defer es.mu.Unlock()

	cutoff := time.Now().Add(-2 * time.Hour)
	for fileID, exportFile := range es.files {
		if exportFile.CreatedAt.Before(cutoff) {
			// 删除文件
			os.Remove(exportFile.FilePath)
			// 删除记录
			delete(es.files, fileID)
		}
	}
}
