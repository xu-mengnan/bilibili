package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONStorage JSON 文件存储实现
type JSONStorage struct {
	dataDir  string       // 数据根目录
	tasksDir string       // 任务数据目录
	mu       sync.RWMutex // 读写锁
}

// NewJSONStorage 创建 JSON 存储实例
func NewJSONStorage(dataDir string) *JSONStorage {
	return &JSONStorage{
		dataDir:  dataDir,
		tasksDir: filepath.Join(dataDir, "tasks"),
	}
}

// Initialize 初始化存储，创建必要的目录结构
func (js *JSONStorage) Initialize() error {
	js.mu.Lock()
	defer js.mu.Unlock()

	// 创建目录结构
	dirs := []string{
		js.tasksDir,
		filepath.Join(js.tasksDir, ".backup"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}

	return nil
}

// SaveTask 保存单个任务的完整数据
func (js *JSONStorage) SaveTask(task *TaskData) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if task == nil {
		return fmt.Errorf("任务数据不能为空")
	}

	taskFile := js.getTaskFilePath(task.TaskID)
	tmpFile := taskFile + ".tmp"

	// 如果文件已存在，先备份
	if _, err := os.Stat(taskFile); err == nil {
		js.createBackup(taskFile)
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化任务数据失败: %w", err)
	}

	// 写入临时文件
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	// 原子重命名
	if err := os.Rename(tmpFile, taskFile); err != nil {
		os.Remove(tmpFile) // 清理临时文件
		return fmt.Errorf("原子重命名失败: %w", err)
	}

	return nil
}

// LoadTask 根据任务ID加载任务的完整数据
func (js *JSONStorage) LoadTask(taskID string) (*TaskData, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	taskFile := js.getTaskFilePath(taskID)

	data, err := os.ReadFile(taskFile)
	if err != nil {
		return nil, fmt.Errorf("读取任务文件失败: %w", err)
	}

	var task TaskData
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("解析任务数据失败: %w", err)
	}

	return &task, nil
}

// DeleteTask 删除单个任务的数据文件
func (js *JSONStorage) DeleteTask(taskID string) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	taskFile := js.getTaskFilePath(taskID)

	// 先备份
	if _, err := os.Stat(taskFile); err == nil {
		js.createBackup(taskFile)
	}

	// 删除文件
	if err := os.Remove(taskFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除任务文件失败: %w", err)
	}

	return nil
}

// SaveIndex 保存任务索引文件
func (js *JSONStorage) SaveIndex(index *TaskIndex) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if index == nil {
		return fmt.Errorf("索引数据不能为空")
	}

	index.Version = "1.0"
	index.LastUpdated = time.Now()

	indexFile := js.getIndexFilePath()
	tmpFile := indexFile + ".tmp"

	// 备份旧索引
	if _, err := os.Stat(indexFile); err == nil {
		js.createBackup(indexFile)
	}

	// 序列化
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化索引失败: %w", err)
	}

	// 写入临时文件
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("写入索引临时文件失败: %w", err)
	}

	// 原子重命名
	if err := os.Rename(tmpFile, indexFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("原子重命名索引文件失败: %w", err)
	}

	return nil
}

// LoadIndex 加载任务索引文件
func (js *JSONStorage) LoadIndex() (*TaskIndex, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	indexFile := js.getIndexFilePath()

	// 文件不存在时返回空索引
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return &TaskIndex{
			Version:     "1.0",
			LastUpdated: time.Now(),
			Tasks:       []TaskMeta{},
		}, nil
	}

	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, fmt.Errorf("读取索引文件失败: %w", err)
	}

	var index TaskIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("解析索引文件失败: %w", err)
	}

	return &index, nil
}

// ListTasks 列出所有任务的元数据
func (js *JSONStorage) ListTasks() ([]TaskMeta, error) {
	index, err := js.LoadIndex()
	if err != nil {
		return nil, err
	}
	return index.Tasks, nil
}

// CleanOldTasks 清理指定时间之前的旧任务
func (js *JSONStorage) CleanOldTasks(beforeTime time.Time) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	// 加载索引
	index, err := js.LoadIndex()
	if err != nil {
		return err
	}

	var keptTasks []TaskMeta
	deletedCount := 0

	for _, task := range index.Tasks {
		// 检查是否需要删除
		if task.EndTime.Before(beforeTime) || task.Status == "failed" {
			// 删除任务文件
			taskFile := js.getTaskFilePath(task.TaskID)
			js.createBackup(taskFile)
			os.Remove(taskFile)
			deletedCount++
		} else {
			keptTasks = append(keptTasks, task)
		}
	}

	if deletedCount > 0 {
		// 更新索引
		index.Tasks = keptTasks
		index.LastUpdated = time.Now()
		data, _ := json.MarshalIndent(index, "", "  ")
		os.WriteFile(js.getIndexFilePath()+".tmp", data, 0644)
		os.Rename(js.getIndexFilePath()+".tmp", js.getIndexFilePath())
	}

	return nil
}

// getTaskFilePath 获取任务数据文件路径
func (js *JSONStorage) getTaskFilePath(taskID string) string {
	return filepath.Join(js.tasksDir, taskID+".json")
}

// getIndexFilePath 获取索引文件路径
func (js *JSONStorage) getIndexFilePath() string {
	return filepath.Join(js.tasksDir, "tasks.json")
}

// createBackup 创建备份文件
func (js *JSONStorage) createBackup(filePath string) {
	timestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(js.tasksDir, ".backup")
	backupName := filepath.Base(filePath) + ".bak." + timestamp
	backupPath := filepath.Join(backupDir, backupName)

	os.Rename(filePath, backupPath)

	// 清理超过 10 个的旧备份
	js.cleanOldBackups(backupDir, filepath.Base(filePath)+".bak.")
}

// cleanOldBackups 清理旧备份文件，保留最新的 10 个
func (js *JSONStorage) cleanOldBackups(backupDir, prefix string) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			backups = append(backups, filepath.Join(backupDir, name))
		}
	}

	// 按修改时间排序（最新的在前）
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			infoI, _ := os.Stat(backups[i])
			infoJ, _ := os.Stat(backups[j])
			if infoI.ModTime().Before(infoJ.ModTime()) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	// 删除超过 10 个的旧备份
	for i := 10; i < len(backups); i++ {
		os.Remove(backups[i])
	}
}
