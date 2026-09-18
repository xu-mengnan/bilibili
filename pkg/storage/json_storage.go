package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// JSONStorage stores task state in JSON files.
// Public methods own the mutex; private helpers never re-acquire it.
type JSONStorage struct {
	dataDir  string
	tasksDir string
	mu       sync.RWMutex
}

func NewJSONStorage(dataDir string) *JSONStorage {
	return &JSONStorage{
		dataDir:  dataDir,
		tasksDir: filepath.Join(dataDir, "tasks"),
	}
}

func (js *JSONStorage) Initialize() error {
	js.mu.Lock()
	defer js.mu.Unlock()
	return js.initializeLocked()
}

func (js *JSONStorage) initializeLocked() error {
	for _, dir := range []string{js.tasksDir, filepath.Join(js.tasksDir, ".backup")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}
	return nil
}

func (js *JSONStorage) SaveTask(task *TaskData) error {
	if task == nil {
		return fmt.Errorf("任务数据不能为空")
	}

	js.mu.Lock()
	defer js.mu.Unlock()

	if err := js.initializeLocked(); err != nil {
		return err
	}
	return js.writeJSONAtomicallyLocked(js.getTaskFilePath(task.TaskID), task)
}

func (js *JSONStorage) LoadTask(taskID string) (*TaskData, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	data, err := os.ReadFile(js.getTaskFilePath(taskID))
	if err != nil {
		return nil, fmt.Errorf("读取任务文件失败: %w", err)
	}

	var task TaskData
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("解析任务数据失败: %w", err)
	}
	return &task, nil
}

func (js *JSONStorage) DeleteTask(taskID string) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	taskFile := js.getTaskFilePath(taskID)
	if _, err := os.Stat(taskFile); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("检查任务文件失败: %w", err)
	}

	if err := js.backupFileLocked(taskFile); err != nil {
		return fmt.Errorf("备份任务文件失败: %w", err)
	}
	if err := os.Remove(taskFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除任务文件失败: %w", err)
	}
	return nil
}

func (js *JSONStorage) SaveIndex(index *TaskIndex) error {
	if index == nil {
		return fmt.Errorf("索引数据不能为空")
	}

	js.mu.Lock()
	defer js.mu.Unlock()

	if err := js.initializeLocked(); err != nil {
		return err
	}

	snapshot := *index
	snapshot.Tasks = append([]TaskMeta(nil), index.Tasks...)
	snapshot.Version = "1.1"
	snapshot.LastUpdated = time.Now()

	return js.writeJSONAtomicallyLocked(js.getIndexFilePath(), &snapshot)
}

func (js *JSONStorage) LoadIndex() (*TaskIndex, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()
	return js.loadIndexLocked()
}

func (js *JSONStorage) loadIndexLocked() (*TaskIndex, error) {
	indexFile := js.getIndexFilePath()
	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &TaskIndex{
				Version:     "1.1",
				LastUpdated: time.Now(),
				Tasks:       []TaskMeta{},
			}, nil
		}
		return nil, fmt.Errorf("读取索引文件失败: %w", err)
	}

	var index TaskIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("解析索引文件失败: %w", err)
	}
	if index.Tasks == nil {
		index.Tasks = []TaskMeta{}
	}
	return &index, nil
}

func (js *JSONStorage) ListTasks() ([]TaskMeta, error) {
	index, err := js.LoadIndex()
	if err != nil {
		return nil, err
	}
	return append([]TaskMeta(nil), index.Tasks...), nil
}

func (js *JSONStorage) CleanOldTasks(beforeTime time.Time) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	index, err := js.loadIndexLocked()
	if err != nil {
		return err
	}

	kept := make([]TaskMeta, 0, len(index.Tasks))
	changed := false

	for _, task := range index.Tasks {
		expired := !task.EndTime.IsZero() && task.EndTime.Before(beforeTime)
		if !expired && task.Status != "failed" {
			kept = append(kept, task)
			continue
		}

		taskFile := js.getTaskFilePath(task.TaskID)
		if _, statErr := os.Stat(taskFile); statErr == nil {
			if err := js.backupFileLocked(taskFile); err != nil {
				return fmt.Errorf("备份旧任务 %s 失败: %w", task.TaskID, err)
			}
			if err := os.Remove(taskFile); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("删除旧任务 %s 失败: %w", task.TaskID, err)
			}
		} else if !os.IsNotExist(statErr) {
			return fmt.Errorf("检查旧任务 %s 失败: %w", task.TaskID, statErr)
		}
		changed = true
	}

	if !changed {
		return nil
	}

	index.Tasks = kept
	index.Version = "1.1"
	index.LastUpdated = time.Now()
	return js.writeJSONAtomicallyLocked(js.getIndexFilePath(), index)
}

func (js *JSONStorage) writeJSONAtomicallyLocked(target string, value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpName := tmp.Name()
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			_ = os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(0644); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("设置临时文件权限失败: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("同步临时文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}

	if _, err := os.Stat(target); err == nil {
		if err := js.backupFileLocked(target); err != nil {
			return fmt.Errorf("备份旧文件失败: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查目标文件失败: %w", err)
	}

	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("原子替换失败: %w", err)
	}
	cleanupTemp = false
	return nil
}

func (js *JSONStorage) backupFileLocked(filePath string) error {
	src, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer src.Close()

	backupDir := filepath.Join(js.tasksDir, ".backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102-150405.000000000")
	backupName := filepath.Base(filePath) + ".bak." + timestamp
	backupPath := filepath.Join(backupDir, backupName)

	dst, err := os.OpenFile(backupPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		_ = os.Remove(backupPath)
		return err
	}
	if err := dst.Sync(); err != nil {
		_ = dst.Close()
		_ = os.Remove(backupPath)
		return err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(backupPath)
		return err
	}

	return js.cleanOldBackupsLocked(backupDir, filepath.Base(filePath)+".bak.")
}

func (js *JSONStorage) cleanOldBackupsLocked(backupDir, prefix string) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}

	type backupFile struct {
		path    string
		modTime time.Time
	}
	backups := make([]backupFile, 0)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		backups = append(backups, backupFile{
			path:    filepath.Join(backupDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.After(backups[j].modTime)
	})

	for i := 10; i < len(backups); i++ {
		if err := os.Remove(backups[i].path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (js *JSONStorage) getTaskFilePath(taskID string) string {
	return filepath.Join(js.tasksDir, taskID+".json")
}

func (js *JSONStorage) getIndexFilePath() string {
	return filepath.Join(js.tasksDir, "tasks.json")
}
