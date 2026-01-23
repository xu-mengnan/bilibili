# Bilibili 评论爬取系统 - 后端完全重构计划

## 概述

基于对现有代码的深入分析，本计划将从零开始重构整个后端服务，解决并发安全、资源泄漏、代码质量等关键问题，并添加 Prometheus 监控能力。

## 目标

1. **健壮性**：解决 goroutine 泄漏、锁升级死锁、内存泄漏等严重问题
2. **代码质量**：统一错误处理、结构化日志、清晰的分层架构
3. **可观测性**：集成 Prometheus 监控和健康检查
4. **V2 优先**：专注于 V2 API，保持简洁

---

## 核心问题总结

| 问题 | 位置 | 严重程度 | 影响 |
|------|------|----------|------|
| Goroutine 泄漏 | `comment.go:66-69` | 严重 | 服务关闭时任务无法中断 |
| 锁升级死锁 | `comment.go:114-138` | 严重 | 潜在死锁，数据竞态 |
| 脏标记失效 | `comment.go:682` | 严重 | 持久化不可靠 |
| 内存泄漏 | `comment.go:29` | 中等 | 完成任务数据常驻内存 |
| 并发安全 | `comment.go:206` | 严重 | task 指针逃逸 |
| 错误处理分散 | 各 handler | 中等 | 格式不统一，无错误码 |
| 缺少中间件 | `api.go` | 中等 | 无 CORS、日志、恢复等 |
| 日志简陋 | 各服务 | 低 | 使用 fmt.Printf |

---

## 新架构设计

### 目录结构

```
bilibili/
├── cmd/server/
│   └── main.go                    # 新入口，支持优雅关闭
│
├── internal/
│   ├── config/
│   │   └── config.go              # 配置结构体（YAML + 环境变量）
│   │
│   ├── domain/
│   │   ├── task.go                # 任务领域模型
│   │   ├── comment.go             # 评论领域模型
│   │   └── errors.go              # 统一错误定义
│   │
│   ├── repository/
│   │   ├── task_repository.go     # 任务仓储接口
│   │   ├── task_json_repo.go      # JSON 实现
│   │   └── task_cache.go          # LRU 缓存装饰器
│   │
│   ├── service/
│   │   ├── task_service.go        # 任务服务（核心）
│   │   ├── analysis_service.go    # AI 分析服务
│   │   └── export_service.go      # 导出服务
│   │
│   ├── handler/
│   │   ├── middleware/
│   │   │   ├── cors.go            # CORS 中间件
│   │   │   ├── logging.go         # 结构化日志中间件
│   │   │   ├── recovery.go        # Panic 恢复中间件
│   │   │   └── metrics.go         # Prometheus 指标中间件
│   │   ├── v2/
│   │   │   ├── task_handler.go    # V2 任务处理器
│   │   │   └── analysis_handler.go # V2 分析处理器
│   │   └── health_handler.go      # 健康检查处理器
│   │
│   ├── server/
│   │   ├── server.go              # HTTP 服务器封装
│   │   └── graceful.go            # 优雅关闭实现
│   │
│   └── telemetry/
│       ├── metrics.go             # Prometheus 指标定义
│       └── logger.go              # 结构化日志（zap）
│
├── pkg/
│   ├── bilibili/                  # 保留（B站 API 客户端）
│   ├── errors/
│   │   ├── errors.go              # 错误包装和码定义
│   │   └── http_errors.go         # HTTP 错误响应
│   └── context/
│       └── context.go             # 自定义 context 键
│
└── configs/
    └── config.yaml                # 新配置格式（YAML）
```

---

## 实现步骤

### 第一阶段：基础设施（可并行创建）

#### 1.1 错误处理系统
**文件**: `internal/domain/errors.go`, `pkg/errors/errors.go`

```go
// 错误码定义
type ErrorCode string

const (
    ErrCodeTaskNotFound      ErrorCode = "TASK_NOT_FOUND"
    ErrCodeTaskInvalidState  ErrorCode = "TASK_INVALID_STATE"
    ErrCodeBilibiliAPI       ErrorCode = "BILIBILI_API_ERROR"
)

type AppError struct {
    Code    ErrorCode
    Message string
    Cause   error
}
```

#### 1.2 结构化日志（zap）
**文件**: `internal/telemetry/logger.go`

- 使用 `go.uber.org/zap`
- 开发环境：console 格式
- 生产环境：JSON 格式

#### 1.3 Prometheus 指标
**文件**: `internal/telemetry/metrics.go`

- `bilibili_tasks_active` - 活跃任务数
- `bilibili_tasks_total` - 任务总数
- `bilibili_task_duration_seconds` - 任务执行时长
- `bilibili_comments_fetched_total` - 评论获取总数
- `bilibili_api_errors_total` - API 错误数

#### 1.4 配置管理（YAML）
**文件**: `internal/config/config.go`, `configs/config.yaml`

```yaml
server:
  host: 0.0.0.0
  port: 8080
  shutdown_timeout: 30s

logging:
  level: info
  format: json

storage:
  data_dir: ./data
  cache_size: 100

ai:
  api_url: https://open.bigmodel.cn/api/paas/v4/chat/completions
  api_key: ${AI_API_KEY}
  model: glm-4.7
```

---

### 第二阶段：领域层

#### 2.1 任务领域模型
**文件**: `internal/domain/task.go`

```go
type Task struct {
    ID          string
    VideoID     string
    Status      TaskStatus  // Pending, Running, Completed, Failed, Cancelled
    Progress    Progress
    Comments    []Comment
    CreatedAt   time.Time
    CompletedAt time.Time
    Error       error
}

type TaskStatus int
const (
    StatusPending TaskStatus = iota
    StatusRunning
    StatusCompleted
    StatusFailed
    StatusCancelled
)
```

#### 2.2 评论领域模型
**文件**: `internal/domain/comment.go`

- 从 `pkg/bilibili/models.go` 提取核心结构
- 去除 JSON 标签等序列化细节

---

### 第三阶段：仓储层

#### 3.1 仓储接口
**文件**: `internal/repository/task_repository.go`

```go
type TaskRepository interface {
    Save(task *domain.Task) error
    Load(id string) (*domain.Task, error)
    UpdateMeta(meta *domain.TaskMeta) error
    List() ([]*domain.TaskMeta, error)
    Delete(id string) error
}
```

#### 3.2 JSON 实现
**文件**: `internal/repository/task_json_repo.go`

- 重构现有的 `pkg/storage/json_storage.go`

#### 3.3 LRU 缓存装饰器
**文件**: `internal/repository/task_cache.go`

- 使用 `github.com/hashicorp/golang-lru`
- 缓存大小：100 个任务
- 自动驱逐旧任务

---

### 第四阶段：服务层（核心）

#### 4.1 任务服务（最复杂）
**文件**: `internal/service/task_service.go`

**关键改进**：
- Context + WaitGroup 生命周期管理
- 读写分离，避免锁升级
- 异步持久化
- 任务状态机

```go
type TaskService struct {
    ctx        context.Context
    cancel     context.CancelFunc
    wg         sync.WaitGroup
    repo       repository.TaskRepository
    logger     *zap.Logger
    metrics    *telemetry.Metrics
    semaphore  chan struct{}  // 并发限制
}

func (s *TaskService) Start(ctx context.Context, req *StartTaskRequest) (string, error)
func (s *TaskService) GetTask(id string) (*domain.Task, error)
func (s *TaskService) Cancel(id string) error
func (s *TaskService) Shutdown(ctx context.Context) error
```

#### 4.2 分析服务
**文件**: `internal/service/analysis_service.go`

- 保留现有 LLM 集成
- 添加结构化日志
- 添加指标

---

### 第五阶段：Handler 层

#### 5.1 中间件
**文件**: `internal/handler/middleware/*.go`

| 中间件 | 职责 |
|--------|------|
| cors.go | CORS 头设置 |
| logging.go | 请求日志记录 |
| recovery.go | Panic 恢复 |
| metrics.go | Prometheus 指标收集 |

#### 5.2 V2 处理器
**文件**: `internal/handler/v2/task_handler.go`

```go
func (h *TaskHandler) StartTask(c *gin.Context)
func (h *TaskHandler) GetTask(c *gin.Context)
func (h *TaskHandler) ListTasks(c *gin.Context)
func (h *TaskHandler) CancelTask(c *gin.Context)
```

#### 5.3 健康检查
**文件**: `internal/handler/health_handler.go`

- `/health` - 健康检查
- `/metrics` - Prometheus 指标

---

### 第六阶段：服务器层

#### 6.1 服务器封装
**文件**: `internal/server/server.go`

```go
type Server struct {
    httpServer  *http.Server
    taskService *service.TaskService
    logger      *zap.Logger
    metrics     *telemetry.Metrics
}

func (s *Server) Start() error
func (s *Server) Shutdown(ctx context.Context) error
```

#### 6.2 优雅关闭
**文件**: `internal/server/graceful.go`

- 处理 SIGTERM/SIGINT 信号
- 先关闭 HTTP 服务（不再接受新请求）
- 再关闭任务服务（等待任务完成）

---

### 第七阶段：主入口

**文件**: `cmd/server/main.go`

```go
func main() {
    // 1. 加载配置
    cfg := config.Load()

    // 2. 初始化日志
    logger, _ := telemetry.NewLogger(cfg.Env)
    defer logger.Sync()

    // 3. 初始化指标
    metrics := telemetry.NewMetrics()
    metrics.Register()

    // 4. 初始化仓储
    repo := repository.NewCachedTaskRepository(...)

    // 5. 初始化服务
    taskService := service.NewTaskService(repo, logger, metrics)

    // 6. 初始化服务器
    server := server.New(cfg, taskService, logger, metrics)

    // 7. 启动服务器（goroutine）
    go func() {
        if err := server.Start(); err != nil {
            logger.Fatal("Server failed", zap.Error(err))
        }
    }()

    // 8. 等待信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    // 9. 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    server.Shutdown(ctx)
}
```

---

## 关键设计决策

### 解决 Goroutine 泄漏

```go
// 启动任务时创建独立 context
taskCtx, taskCancel := context.WithCancel(ctx)

s.wg.Add(1)
go func() {
    defer s.wg.Done()
    defer taskCancel()

    for page := 1; page <= limit; page++ {
        select {
        case <-taskCtx.Done():
            return // 任务被取消
        default:
        }
        // 执行爬取逻辑
    }
}()

// 优雅关闭时取消所有任务
func (s *TaskService) Shutdown(ctx context.Context) error {
    s.cancel()  // 取消 context
    s.wg.Wait() // 等待所有 goroutine
    return nil
}
```

### 解决锁升级死锁

```go
// 读写分离，不再混用
func (s *TaskService) GetTask(id string) (*domain.Task, error) {
    // 读操作只用读锁
    s.mu.RLock()
    handle, exists := s.tasks[id]
    s.mu.RUnlock()

    if !exists {
        return s.repo.Load(id) // 从存储加载
    }
    return &domain.Task{Meta: *handle.meta}, nil
}

func (s *TaskService) updateProgress(id string, p int) error {
    // 写操作只用写锁
    s.mu.Lock()
    defer s.mu.Unlock()
    // 更新逻辑
    go s.repo.UpdateMeta(handle.meta) // 异步持久化
    return nil
}
```

### 解决内存泄漏

```go
// LRU 缓存自动驱逐
type CachedTaskRepository struct {
    base  repository.TaskRepository
    cache *lru.Cache
}

func NewCachedTaskRepository(base repository.TaskRepository, size int) *CachedTaskRepository {
    cache, _ := lru.NewWithEvict(size, func(key, value interface{}) {
        log.Debug("Task evicted from cache", "id", key)
    })
    return &CachedTaskRepository{base: base, cache: cache}
}
```

---

## 验证计划

### 单元测试

```bash
# 并发安全测试
go test -race ./internal/...

# 覆盖率
go test -cover ./internal/...
```

### 压力测试

```bash
# 使用 hey 进行压测
hey -n 1000 -c 100 -m POST \
  -H "Content-Type: application/json" \
  -d '{"video_id":"test123"}' \
  http://localhost:8080/api/v2/tasks
```

### 监控验证

```bash
# 检查 Prometheus 指标
curl http://localhost:8080/metrics | grep bilibili_

# 检查健康状态
curl http://localhost:8080/health
```

---

## 依赖关系

```
第一阶段（基础设施）
├── domain/errors.go
├── telemetry/logger.go
├── telemetry/metrics.go
└── config/config.go

第二阶段（领域层，依赖第一阶段）
├── domain/task.go
└── domain/comment.go

第三阶段（仓储层，依赖第二阶段）
├── repository/task_repository.go
├── repository/task_json_repo.go
└── repository/task_cache.go

第四阶段（服务层，依赖第三阶段）
├── service/task_service.go
└── service/analysis_service.go

第五阶段（Handler 层，依赖第四阶段）
├── handler/middleware/*.go
├── handler/v2/task_handler.go
└── handler/health_handler.go

第六阶段（服务器层，依赖第五阶段）
├── server/server.go
└── server/graceful.go

第七阶段（入口点）
└── cmd/server/main.go
```

---

## 关键文件路径

| 文件 | 职责 |
|------|------|
| `internal/service/task_service.go` | 核心任务服务，解决并发/泄漏问题 |
| `internal/handler/v2/task_handler.go` | V2 API 处理器 |
| `internal/repository/task_cache.go` | LRU 缓存，解决内存泄漏 |
| `internal/domain/errors.go` | 统一错误定义 |
| `internal/telemetry/metrics.go` | Prometheus 监控 |
| `cmd/server/main.go` | 新入口，优雅关闭 |

---

## 预期改进

| 指标 | 改进 |
|------|------|
| 并发安全 | 消除死锁和数据竞态风险 |
| 内存使用 | LRU 缓存限制在 100 个任务 |
| 可观测性 | Prometheus 指标 + 结构化日志 |
| 可维护性 | 清晰的分层架构 + 统一错误处理 |
| 可靠性 | 优雅关闭 + 任务取消支持 |
