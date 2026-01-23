# Bilibili 评论爬取系统 - 渐进式重构执行计划

## 🎯 总体目标
一次性完成第1-3阶段，解决核心问题，保持 API 兼容性。

---

## 用户选择的配置

- **日志方案**：增强 `pkg/utils/logger.go`（零依赖）
- **中间件优先级**：先加 Logging（方便调试）
- **Prometheus**：暂不需要，够用就行
- **内存优化**：实现懒加载
- **错误处理**：定义统一的 APIError 结构
- **重构范围**：一次性完成第1-3阶段

---

## 第一阶段：生命周期管理（核心）

### 1.1 增强 `pkg/utils/logger.go`
**目标**：零依赖的结构化日志系统

**增强内容**：
- 添加日志级别（DEBUG, INFO, WARN, ERROR）
- 添加 JSON 格式支持（可选）
- 添加请求 ID 追踪支持
- 添加格式化日志方法（WithFields）

**新增 API**：
```go
type Level int
const (
    DEBUG Level = iota
    INFO
    WARN
    ERROR
)

func LogWithFields(level Level, fields map[string]interface{}, message string)
func LogRequest(method, path string, statusCode int, duration time.Duration)
func LogErrorWithFields(err error, fields map[string]interface{}, message string)
```

**影响文件**：
- `pkg/utils/logger.go` - 增强

---

### 1.2 改造 `CommentService` - 添加 context 控制
**目标**：所有 goroutine 支持优雅取消

**修改内容**：

1. **添加字段**：
```go
type CommentService struct {
    ctx        context.Context
    cancel     context.CancelFunc
    wg         sync.WaitGroup
    tasks      map[string]*ScrapeTask
    mu         sync.RWMutex
    storage    storage.TaskStorage
    dirty      map[string]bool
}
```

2. **修改 `NewCommentService`**：
```go
func NewCommentService(ctx context.Context, storage storage.TaskStorage) *CommentService {
    serviceCtx, cancel := context.WithCancel(ctx)
    cs := &CommentService{
        ctx:    serviceCtx,
        cancel: cancel,
        // ... 其他字段
    }

    // 使用 wg 追踪所有 goroutine
    cs.wg.Add(2)
    go func() {
        defer cs.wg.Done()
        cs.persistWorker()
    }()
    go func() {
        defer cs.wg.Done()
        cs.cleanupWorker()
    }()

    return cs
}
```

3. **修改 `persistWorker`**：
```go
func (cs *CommentService) persistWorker() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-cs.ctx.Done():
            utils.LogInfo("persistWorker stopped")
            return
        case <-ticker.C:
            cs.persistDirtyTasks()
        }
    }
}
```

4. **修改 `cleanupWorker`**：
```go
func (cs *CommentService) cleanupWorker() {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-cs.ctx.Done():
            utils.LogInfo("cleanupWorker stopped")
            return
        case <-ticker.C:
            cs.CleanOldTasks()
        }
    }
}
```

5. **修改 `StartScrapeTask`**：
```go
func (cs *CommentService) StartScrapeTask(...) (string, error) {
    // ... 创建任务 ...

    cs.wg.Add(1)
    go func() {
        defer cs.wg.Done()
        cs.executeScrapingTask(taskID)
    }()

    return taskID, nil
}
```

6. **修改 `executeScrapingTask`**：
```go
func (cs *CommentService) executeScrapingTask(taskID string) {
    cs.mu.RLock()
    task := cs.tasks[taskID]
    cs.mu.RUnlock()

    if task == nil {
        return
    }

    // ... 获取视频信息 ...

    for page := 1; page <= task.PageLimit; page++ {
        // 检查是否被取消
        select {
        case <-cs.ctx.Done():
            utils.LogInfo("Scraping task cancelled: " + taskID)
            cs.mu.Lock()
            task.Status = "cancelled"
            task.Error = "Task cancelled by shutdown"
            task.EndTime = time.Now()
            cs.mu.Unlock()
            return
        default:
        }

        // ... 获取评论逻辑 ...
    }

    // ... 完成逻辑 ...
}
```

7. **添加 `Shutdown` 方法**：
```go
func (cs *CommentService) Shutdown(ctx context.Context) error {
    utils.LogInfo("Shutting down CommentService...")

    // 取消 context
    cs.cancel()

    // 等待所有 goroutine 结束
    done := make(chan struct{})
    go func() {
        cs.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        utils.LogInfo("CommentService shutdown complete")
        return nil
    case <-ctx.Done():
        utils.LogError("CommentService shutdown timeout")
        return ctx.Err()
    }
}
```

**影响文件**：
- `internal/services/comment.go` - 大量修改

---

### 1.3 改造 `ExportService` - 添加 context 控制
**目标**：清理 goroutine 支持取消

**修改内容**：

1. **添加字段**：
```go
type ExportService struct {
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
    exportDir string
    files     map[string]*ExportFile
    mu        sync.RWMutex
}
```

2. **修改 `NewExportService`** 和 `cleanupWorker`（类似 CommentService）
3. **添加 `Shutdown` 方法**

**影响文件**：
- `internal/services/export.go` - 修改

---

### 1.4 改造 `AnalysisService` - 添加 context 控制
**目标**：流式分析支持取消

**修改内容**：

1. **修改 `CallLLMStream` 添加 context 参数**：
```go
func (s *AnalysisService) CallLLMStream(ctx context.Context, callback ChunkCallback, prompt string) (string, error) {
    // ... 准备请求 ...

    req = req.WithContext(ctx) // 使用传入的 context

    // ... 读取流时检查取消 ...
    for {
        select {
        case <-ctx.Done():
            return "", ctx.Err()
        default:
        }
        // ... 读取逻辑 ...
    }
}
```

2. **修改 handler 传递 context**：
```go
go func() {
    // 格式化评论数据
    commentsText := h.analysisService.FormatComments(task.Comments, req.CommentLimit)

    // 渲染 Prompt
    prompt := h.analysisService.RenderTemplate(template, commentsText, task.VideoTitle, len(task.Comments))

    // 调用流式 LLM，传递 context
    _, err := h.analysisService.CallLLMStream(c.Request.Context(), func(chunk string) {
        streamChan <- chunk
    }, prompt)

    if err != nil {
        errorChan <- err
        return
    }

    streamChan <- "[DONE]"
}()
```

**影响文件**：
- `internal/services/analysis.go` - 修改
- `internal/handlers/analysis.go` - 修改

---

### 1.5 改造 `main.go` - 优雅关闭
**目标**：处理信号，顺序关闭服务

**完整实现**：
```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "bilibili/api"
    "bilibili/pkg/utils"
)

func main() {
    utils.LogInfo("Starting Bilibili Comment Scraper...")

    // 创建 root context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 设置路由
    router := api.SetupRoutes(ctx)

    // 创建 HTTP 服务器
    server := &http.Server{
        Addr:         ":8080",
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // 启动 HTTP 服务器（goroutine）
    errChan := make(chan error, 1)
    go func() {
        utils.LogInfo("Server listening on :8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errChan <- fmt.Errorf("server failed: %w", err)
        }
    }()

    // 等待信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-errChan:
        utils.LogError("Server error: " + err.Error())
        os.Exit(1)
    case sig := <-sigChan:
        utils.LogInfo("Received signal: " + sig.String())
    }

    // 优雅关闭
    utils.LogInfo("Shutting down gracefully...")
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    // 1. 关闭 HTTP 服务器（不再接受新请求）
    utils.LogInfo("Shutting down HTTP server...")
    if err := server.Shutdown(shutdownCtx); err != nil {
        utils.LogError("HTTP server shutdown error: " + err.Error())
    }

    // 2. 关闭所有服务
    api.ShutdownServices(shutdownCtx)

    utils.LogInfo("Shutdown complete")
}
```

**影响文件**：
- `cmd/app/main.go` - 完全重写

---

### 1.6 修改 `api.SetupRoutes` 签名
**目标**：接受 context 并返回服务实例

**修改内容**：
```go
type Services struct {
    CommentService  *services.CommentService
    ExportService   *services.ExportService
    AnalysisService *services.AnalysisService
}

func SetupRoutes(ctx context.Context) (*gin.Engine, *Services) {
    r := gin.Default()

    // 加载配置
    cfg, _ := config.LoadDefault()

    // 初始化存储
    taskStorage := storage.NewJSONStorage(cfg.Storage.DataDir)

    // 初始化服务（传递 context）
    commentService := services.NewCommentService(ctx, taskStorage)
    exportService := services.NewExportService(ctx, "./exports")
    analysisService := services.NewAnalysisService(
        cfg.AI.APIURL,
        cfg.AI.APIKey,
        cfg.AI.Model,
    )

    services := &Services{
        CommentService:  commentService,
        ExportService:   exportService,
        AnalysisService: analysisService,
    }

    // ... 设置路由 ...

    return r, services
}

func ShutdownServices(ctx context.Context) {
    // 获取服务实例并调用 Shutdown
}
```

**影响文件**：
- `api/api.go` - 大量修改

---

### 1.7 创建 `internal/server/shutdown.go`（可选）
**目标**：统一管理服务关闭

**实现内容**：
```go
package server

type ShutdownManager struct {
    services []Shutdownable
}

type Shutdownable interface {
    Shutdown(ctx context.Context) error
}

func NewShutdownManager() *ShutdownManager {
    return &ShutdownManager{
        services: make([]Shutdownable, 0),
    }
}

func (sm *ShutdownManager) Register(service Shutdownable) {
    sm.services = append(sm.services, service)
}

func (sm *ShutdownManager) ShutdownAll(ctx context.Context) error {
    for _, service := range sm.services {
        if err := service.Shutdown(ctx); err != nil {
            return err
        }
    }
    return nil
}
```

**新增文件**：
- `internal/server/shutdown.go`

---

## 第二阶段：并发安全改进

### 2.1 修复 `GetTaskProgress` 锁升级问题
**目标**：消除锁升级风险

**修改方案**：完全重写，避免先读后写

```go
func (cs *CommentService) GetTaskProgress(taskID string) (*ScrapeTask, error) {
    // 方案：先尝试从内存获取
    cs.mu.RLock()
    task, exists := cs.tasks[taskID]

    // 如果任务不存在，尝试从存储加载
    if !exists {
        cs.mu.RUnlock()
        // 获取读锁从索引加载元数据
        cs.mu.RLock()
        defer cs.mu.RUnlock()

        index, err := cs.storage.LoadIndex()
        if err != nil {
            return nil, err
        }

        for _, meta := range index.Tasks {
            if meta.TaskID == taskID {
                // 构建任务对象（不含评论数据）
                task = &ScrapeTask{
                    TaskID:     meta.TaskID,
                    VideoID:    meta.VideoID,
                    VideoTitle: meta.VideoTitle,
                    Status:     meta.Status,
                    Comments:   nil, // 懒加载
                    Progress: TaskProgress{
                        TotalComments: meta.CommentCount,
                    },
                    StartTime: meta.StartTime,
                    EndTime:   meta.EndTime,
                    Error:     meta.Error,
                }
                exists = true
                break
            }
        }

        if !exists {
            return nil, fmt.Errorf("task not found: %s", taskID)
        }

        return task, nil
    }

    // 对于 completed 状态的任务，检查评论数据
    if task.Status == "completed" && (task.Comments == nil || len(task.Comments) == 0) {
        // 需要加载评论数据
        // 直接从存储加载，不持有锁
        cs.mu.RUnlock()

        taskData, err := cs.storage.LoadTask(taskID)
        if err != nil {
            return nil, fmt.Errorf("failed to load task comments: %w", err)
        }

        comments := cs.convertFromStorageFormat(taskData.Comments)

        // 再次获取锁并更新（双重检查）
        cs.mu.Lock()
        task = cs.tasks[taskID] // 重新获取（可能已被删除或加载）
        if task != nil && (task.Comments == nil || len(task.Comments) == 0) {
            task.Comments = comments
            task.Progress.TotalComments = len(comments)
        }
        cs.mu.Unlock()

        return task, nil
    }

    cs.mu.RUnlock()
    return task, nil
}
```

**影响文件**：
- `internal/services/comment.go:114-138` - 重写

---

### 2.2 完善懒加载机制
**目标**：任务完成后释放大块评论数据

**修改内容**：

1. **修改 `executeScrapingTask` 完成逻辑**：
```go
// 标记任务完成
cs.mu.Lock()
task.Status = "completed"
task.Comments = comments // 临时保存，用于持久化
task.Progress.TotalComments = len(comments)
task.EndTime = time.Now()
cs.mu.Unlock()

// 立即持久化完成的任务
cs.saveTask(task)

// 持久化后释放内存
cs.mu.Lock()
task.Comments = nil // 释放内存，下次查询时懒加载
cs.mu.Unlock()
```

2. **修改 `GetTaskResult` 确保懒加载**：
```go
// 懒加载：如果评论数据未加载，从存储加载
if task.Comments == nil || len(task.Comments) == 0 {
    taskData, err := cs.storage.LoadTask(taskID)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to load comments: %w", err)
    }

    comments := cs.convertFromStorageFormat(taskData.Comments)

    // 更新任务
    cs.mu.Lock()
    task.Comments = comments
    cs.mu.Unlock()
}
```

**影响文件**：
- `internal/services/comment.go:325-335` - 修改

---

## 第三阶段：基础设施增强

### 3.1 添加 Logging 中间件
**目标**：结构化请求日志

**新建文件**：`internal/handlers/middleware/logging.go`

**实现内容**：
```go
package middleware

import (
    "time"

    "bilibili/pkg/utils"
    "github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        // 处理请求
        c.Next()

        // 计算耗时
        latency := time.Since(start)
        statusCode := c.Writer.Status()

        // 记录日志
        fields := map[string]interface{}{
            "method":     c.Request.Method,
            "path":       path,
            "query":      query,
            "status":     statusCode,
            "ip":         c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
            "latency":    latency,
        }

        if statusCode >= 500 {
            utils.LogErrorWithFields(nil, fields, "Server error")
        } else if statusCode >= 400 {
            utils.LogWithFields(utils.WARN, fields, "Client error")
        } else {
            utils.LogWithFields(utils.INFO, fields, "Request completed")
        }
    }
}
```

**新增文件**：
- `internal/handlers/middleware/logging.go`

---

### 3.2 添加 Recovery 中间件
**目标**：捕获 panic 并优雅恢复

**新建文件**：`internal/handlers/middleware/recovery.go`

**实现内容**：
```go
package middleware

import (
    "fmt"
    "net/http"
    "runtime/debug"

    "bilibili/pkg/utils"
    "github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 记录 panic 信息
                stack := debug.Stack()
                utils.LogError(fmt.Sprintf("Panic recovered: %v\n%s", err, string(stack)))

                // 返回错误响应
                c.JSON(http.StatusInternalServerError, gin.H{
                    "code":    "INTERNAL_SERVER_ERROR",
                    "message": "Internal server error",
                })
                c.Abort()
            }
        }()

        c.Next()
    }
}
```

**新增文件**：
- `internal/handlers/middleware/recovery.go`

---

### 3.3 添加 CORS 中间件
**目标**：支持跨域请求

**新建文件**：`internal/handlers/middleware/cors.go`

**实现内容**：
```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

**新增文件**：
- `internal/handlers/middleware/cors.go`

---

### 3.4 统一错误处理
**目标**：定义统一的 APIError 结构

**新建文件**：`internal/errors/api_errors.go`

**实现内容**：
```go
package errors

import "net/http"

// APIError 统一错误响应结构
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// 错误码定义
const (
    ErrCodeBadRequest         = "BAD_REQUEST"
    ErrCodeUnauthorized       = "UNAUTHORIZED"
    ErrCodeNotFound           = "NOT_FOUND"
    ErrCodeConflict           = "CONFLICT"
    ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
    ErrCodeTaskNotFound       = "TASK_NOT_FOUND"
    ErrCodeTaskInvalidState   = "TASK_INVALID_STATE"
    ErrCodeBilibiliAPI        = "BILIBILI_API_ERROR"
)

// NewAPIError 创建错误
func NewAPIError(code, message, details string) *APIError {
    return &APIError{
        Code:    code,
        Message: message,
        Details: details,
    }
}

// NewBadRequest 400 错误
func NewBadRequest(message string) *APIError {
    return NewAPIError(ErrCodeBadRequest, message, "")
}

// NewNotFound 404 错误
func NewNotFound(message string) *APIError {
    return NewAPIError(ErrCodeNotFound, message, "")
}

// NewInternalError 500 错误
func NewInternalError(message string) *APIError {
    return NewAPIError(ErrCodeInternalServerError, message, "")
}

// GetHTTPStatus 获取错误对应的 HTTP 状态码
func (e *APIError) GetHTTPStatus() int {
    switch e.Code {
    case ErrCodeBadRequest:
        return http.StatusBadRequest
    case ErrCodeNotFound:
        return http.StatusNotFound
    case ErrCodeTaskNotFound:
        return http.StatusNotFound
    case ErrCodeConflict:
        return http.StatusConflict
    case ErrCodeInternalServerError:
        return http.StatusInternalServerError
    default:
        return http.StatusInternalServerError
    }
}
```

**新增文件**：
- `internal/errors/api_errors.go`

---

### 3.5 修改 Handler 使用统一错误处理
**目标**：所有 handler 使用统一的错误响应

**示例修改**：
```go
func (h *CommentHandlers) GetProgressHandler(c *gin.Context) {
    taskID := c.Param("task_id")

    task, err := h.commentService.GetTaskProgress(taskID)
    if err != nil {
        apiErr := apierrors.NewNotFound("Task not found")
        c.JSON(apiErr.GetHTTPStatus(), apiErr)
        return
    }

    c.JSON(http.StatusOK, task)
}
```

**影响文件**：
- `internal/handlers/comment.go` - 修改多处
- `internal/handlers/analysis.go` - 修改多处
- `internal/handlers/v2_api.go` - 修改多处

---

### 3.6 添加健康检查端点
**目标**：提供 `/health` 端点

**新建文件**：`internal/handlers/health.go`

**实现内容**：
```go
package handlers

import (
    "net/http"

    "bilibili/internal/services"
    "github.com/gin-gonic/gin"
)

type HealthHandler struct {
    services *Services
}

func NewHealthHandler(services *Services) *HealthHandler {
    return &HealthHandler{
        services: services,
    }
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
    status := gin.H{
        "status": "healthy",
        "services": gin.H{
            "comment_service":  "ok",
            "export_service":   "ok",
            "analysis_service": "ok",
        },
    }

    c.JSON(http.StatusOK, status)
}
```

**影响文件**：
- `api/api.go` - 添加路由

---

## 📊 执行顺序

| 顺序 | 任务 | 预计工作量 | 依赖 |
|------|------|-----------|------|
| 1 | 增强 logger.go | 30分钟 | 无 |
| 2 | 改造 CommentService | 2小时 | 1 |
| 3 | 改造 ExportService | 1小时 | 2 |
| 4 | 改造 AnalysisService | 1小时 | 2 |
| 5 | 修改 main.go | 30分钟 | 2,3,4 |
| 6 | 修改 api.go | 1小时 | 5 |
| 7 | 修复 GetTaskProgress | 1小时 | 2 |
| 8 | 完善懒加载 | 30分钟 | 7 |
| 9 | 添加 Logging 中间件 | 30分钟 | 1 |
| 10 | 添加 Recovery 中间件 | 20分钟 | 无 |
| 11 | 添加 CORS 中间件 | 20分钟 | 无 |
| 12 | 统一错误处理 | 2小时 | 11 |
| 13 | 添加健康检查 | 20分钟 | 6 |

**总计**：约 10-12 小时

---

## ✅ 验证清单

### 功能验证
- [ ] 启动服务正常
- [ ] 创建爬取任务成功
- [ ] 获取任务进度正常
- [ ] 懒加载评论数据正常
- [ ] 导出功能正常
- [ ] AI 分析功能正常
- [ ] 健康检查端点正常

### 优雅关闭验证
- [ ] 按 Ctrl+C 服务正常退出
- [ ] 所有 goroutine 正常结束
- [ ] 正在运行的任务被正确取消

### 日志验证
- [ ] 日志输出包含时间、级别、消息
- [ ] 请求日志记录方法、路径、状态码、耗时
- [ ] 错误日志包含详细信息

### 并发安全验证
- [ ] 使用 `go test -race` 无数据竞态
- [ ] 压力测试无 panic

---

## 📝 新增文件列表

```
internal/
├── errors/
│   └── api_errors.go           # 统一错误定义
├── handlers/
│   └── middleware/
│       ├── cors.go            # CORS 中间件
│       ├── logging.go         # 日志中间件
│       └── recovery.go        # 恢复中间件
├── handlers/
│   └── health.go              # 健康检查（可能已存在）
└── server/
    └── shutdown.go            # 优雅关闭管理器（可选）
```

---

## 🔄 修改文件列表

```
修改文件：
- pkg/utils/logger.go                    # 增强
- internal/services/comment.go           # 大量修改
- internal/services/export.go            # 修改
- internal/services/analysis.go          # 修改
- cmd/app/main.go                        # 完全重写
- api/api.go                             # 大量修改
- internal/handlers/comment.go           # 错误处理
- internal/handlers/analysis.go          # 错误处理
- internal/handlers/v2_api.go            # 错误处理
```

---

## 📅 执行日志

### 第一阶段执行记录

- [x] 任务 1.1: 增强 `pkg/utils/logger.go`
- [ ] 任务 1.2: 改造 `CommentService`
- [ ] 任务 1.3: 改造 `ExportService`
- [ ] 任务 1.4: 改造 `AnalysisService`
- [ ] 任务 1.5: 改造 `main.go`
- [ ] 任务 1.6: 修改 `api.go`

### 第二阶段执行记录

- [ ] 任务 2.1: 修复 `GetTaskProgress`
- [ ] 任务 2.2: 完善懒加载

### 第三阶段执行记录

- [ ] 任务 3.1: 添加 Logging 中间件
- [ ] 任务 3.2: 添加 Recovery 中间件
- [ ] 任务 3.3: 添加 CORS 中间件
- [ ] 任务 3.4: 统一错误处理
- [ ] 任务 3.5: 修改 Handler
- [ ] 任务 3.6: 添加健康检查

---

**文档版本**: 1.0
**创建日期**: 2026-01-23
**最后更新**: 2026-01-23
