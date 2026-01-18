package storage

import "time"

// TaskStorage 任务存储接口
// 定义了任务持久化的抽象层，支持不同的存储实现
type TaskStorage interface {
	// SaveTask 保存单个任务的完整数据
	SaveTask(task *TaskData) error

	// LoadTask 根据任务ID加载任务的完整数据
	LoadTask(taskID string) (*TaskData, error)

	// DeleteTask 删除单个任务的数据文件和索引
	DeleteTask(taskID string) error

	// SaveIndex 保存任务索引文件
	SaveIndex(index *TaskIndex) error

	// LoadIndex 加载任务索引文件
	LoadIndex() (*TaskIndex, error)

	// ListTasks 列出所有任务的元数据（通过索引）
	ListTasks() ([]TaskMeta, error)

	// CleanOldTasks 清理指定时间之前的旧任务
	CleanOldTasks(beforeTime time.Time) error

	// Initialize 初始化存储（创建目录结构等）
	Initialize() error
}
