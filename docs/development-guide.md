# 开发指南

本文档帮助开发者快速了解项目结构、开发流程和最佳实践。

## 目录

- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [开发环境配置](#开发环境配置)
- [代码规范](#代码规范)
- [开发流程](#开发流程)
- [测试](#测试)
- [调试技巧](#调试技巧)
- [常见问题](#常见问题)

---

## 快速开始

### 前置要求

- **Go**: 1.19 或更高版本
- **Git**: 用于版本控制
- **浏览器**: 用于测试Web界面

### 克隆项目

```bash
git clone <repository-url>
cd bilibili
```

### 安装依赖

```bash
go mod download
```

### 运行项目

```bash
# 开发模式运行
go run ./cmd/app

# 或者先构建再运行
go build -o bin/app ./cmd/app
./bin/app  # Linux/Mac
# 或
bin\app.exe  # Windows
```

访问 `http://localhost:8080` 查看Web界面。

### 运行示例

```bash
# 运行Bilibili API示例
go run examples/bilibili_example.go

# 运行文件处理示例
go run examples/file_example.go
```

---

## 项目结构

```
bilibili/
├── api/                    # API路由定义
│   └── api.go             # 主路由配置
├── cmd/                    # 程序入口
│   └── app/
│       └── main.go        # 主程序入口
├── configs/                # 配置文件
│   └── config.json        # 服务器配置
├── changelogs/             # 变更日志
│   ├── README.md          # 日志索引
│   └── *.md               # 具体变更记录
├── docs/                   # 项目文档
│   ├── api-reference.md   # API参考
│   ├── comment_sort_mode.md  # 功能说明
│   └── development-guide.md  # 本文档
├── examples/               # 示例代码
│   ├── bilibili_example.go
│   └── file_example.go
├── exports/                # 导出文件存储目录
├── internal/               # 私有应用代码
│   ├── handlers/          # HTTP处理器
│   │   ├── comment.go     # 评论相关处理器
│   │   └── video.go       # 视频相关处理器
│   └── services/          # 业务逻辑层
│       ├── comment.go     # 评论服务
│       └── export.go      # 导出服务
├── pkg/                    # 可被外部引用的公共代码
│   ├── bilibili/          # Bilibili API客户端
│   │   ├── client.go      # HTTP客户端
│   │   ├── wbi.go         # WBI签名
│   │   ├── comment.go     # 评论API
│   │   ├── video.go       # 视频API
│   │   ├── user.go        # 用户API
│   │   └── models.go      # 数据模型
│   ├── file/              # 文件处理
│   │   ├── excel.go       # Excel操作
│   │   └── csv.go         # CSV操作
│   └── utils/             # 工具类
├── static/                 # 静态资源
│   ├── index.html         # 主页面
│   ├── css/
│   │   └── style.css      # 样式文件
│   └── js/
│       ├── api.js         # API调用
│       ├── app.js         # 应用逻辑
│       └── charts.js      # 图表渲染
├── go.mod                  # Go模块文件
├── go.sum                  # 依赖校验和
├── README.md               # 项目说明
└── CLAUDE.md               # Claude Code指南
```

### 目录说明

#### `pkg/` - 公共包
存放可被外部项目引用的代码，应保持稳定的API。

- **pkg/bilibili**: Bilibili API封装
  - 使用函数式选项模式提供灵活配置
  - 支持Cookie和APP认证
  - 自动处理WBI签名
- **pkg/file**: 文件处理工具
  - Excel读写（基于excelize）
  - CSV读写
- **pkg/utils**: 通用工具函数

#### `internal/` - 私有包
仅供本项目内部使用的代码。

- **internal/handlers**: HTTP请求处理器
  - 处理HTTP请求和响应
  - 参数验证和转换
  - 调用服务层执行业务逻辑
- **internal/services**: 业务逻辑层
  - 实现核心业务逻辑
  - 任务管理和进度跟踪
  - 数据处理和导出

#### `api/` - 路由定义
- 集中管理所有HTTP路由
- 配置中间件
- 静态文件服务

#### `static/` - 前端资源
- 单页应用（SPA）
- 使用原生JavaScript（无框架依赖）
- Chart.js用于数据可视化

---

## 开发环境配置

### IDE推荐

- **VS Code**: 配合Go扩展
- **GoLand**: JetBrains的Go IDE

### VS Code配置

安装扩展：
- Go (golang.go)
- REST Client (humao.rest-client)

推荐的 `.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintOnSave": "package",
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

### 环境变量

可选的环境变量：
```bash
# 服务器端口（默认8080）
export PORT=8080

# 日志级别（debug/info/warn/error）
export LOG_LEVEL=info
```

---

## 代码规范

### Go代码规范

遵循官方Go代码规范：
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

**关键点**：

1. **命名**:
   - 包名: 小写单词，无下划线
   - 导出函数: 大写开头（PascalCase）
   - 私有函数: 小写开头（camelCase）
   - 常量: 大写开头或全大写

2. **注释**:
   - 导出的类型、函数、常量必须有注释
   - 注释以名称开头
   ```go
   // GetComments 获取视频评论
   func GetComments(oid int64) ([]CommentData, error) { ... }
   ```

3. **错误处理**:
   - 始终检查错误
   - 使用 `fmt.Errorf` 包装错误添加上下文
   ```go
   if err != nil {
       return nil, fmt.Errorf("failed to get comments: %w", err)
   }
   ```

4. **包组织**:
   ```go
   import (
       // 标准库
       "fmt"
       "time"

       // 第三方库
       "github.com/gin-gonic/gin"

       // 本地包
       "bilibili/pkg/bilibili"
   )
   ```

### 提交规范

使用约定式提交（Conventional Commits）：

```
<类型>(<范围>): <简短描述>

<详细描述>

<尾注>
```

**类型**：
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式调整（不影响功能）
- `refactor`: 重构（不是新功能也不是修复bug）
- `test`: 添加或修改测试
- `chore`: 构建过程或辅助工具的变动

**示例**：
```
feat(bilibili): 添加热门评论抓取功能

新增按热度排序评论的功能，支持通过WithSortMode选项设置。

- 添加CommentOptions结构体
- 实现WithSortMode函数
- 更新GetComments函数支持排序模式
```

---

## 开发流程

### 添加新功能

1. **创建功能分支**
   ```bash
   git checkout -b feature/new-feature
   ```

2. **编写代码**
   - 在 `pkg/` 或 `internal/` 中实现功能
   - 添加必要的测试
   - 更新文档

3. **测试功能**
   ```bash
   go test ./...
   go run ./cmd/app
   ```

4. **提交代码**
   ```bash
   git add .
   git commit -m "feat: 添加新功能描述"
   ```

5. **合并分支**
   ```bash
   git checkout main
   git merge feature/new-feature
   ```

### 修复Bug

1. **创建修复分支**
   ```bash
   git checkout -b fix/bug-description
   ```

2. **定位问题**
   - 添加测试重现bug
   - 使用调试工具定位原因

3. **修复并测试**
   ```bash
   go test ./...
   ```

4. **提交**
   ```bash
   git commit -m "fix: 修复具体问题描述"
   ```

### 添加新的API端点

1. **定义路由** (`api/api.go`)
   ```go
   router.GET("/api/new-endpoint", handlers.NewEndpointHandler)
   ```

2. **实现处理器** (`internal/handlers/`)
   ```go
   func NewEndpointHandler(c *gin.Context) {
       // 参数解析
       // 调用服务层
       // 返回响应
   }
   ```

3. **实现服务** (`internal/services/`)
   ```go
   func (s *Service) NewOperation() error {
       // 业务逻辑
       return nil
   }
   ```

4. **更新文档** (`docs/api-reference.md`)

---

## 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行指定包的测试
go test ./pkg/bilibili

# 运行指定测试函数
go test -v ./pkg/bilibili -run TestGetComments

# 查看测试覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 编写测试

**单元测试示例**:
```go
// pkg/bilibili/comment_test.go
package bilibili

import "testing"

func TestGetComments(t *testing.T) {
    tests := []struct {
        name    string
        oid     int64
        wantErr bool
    }{
        {"valid oid", 123456, false},
        {"invalid oid", -1, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := GetComments(tt.oid, 1, 20, 0)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetComments() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**集成测试示例**:
```go
func TestScrapingWorkflow(t *testing.T) {
    // 启动任务
    taskID, err := service.StartScrapeTask(...)
    if err != nil {
        t.Fatal(err)
    }

    // 等待完成
    time.Sleep(5 * time.Second)

    // 验证结果
    task, err := service.GetTaskProgress(taskID)
    if err != nil {
        t.Fatal(err)
    }

    if task.Status != "completed" {
        t.Errorf("expected completed, got %s", task.Status)
    }
}
```

---

## 调试技巧

### 使用Delve调试器

安装Delve：
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

调试命令：
```bash
# 调试主程序
dlv debug ./cmd/app

# 调试测试
dlv test ./pkg/bilibili -- -test.run TestGetComments
```

### 日志调试

添加详细日志：
```go
import "log"

log.Printf("Debug: oid=%d, page=%d\n", oid, page)
```

### HTTP请求调试

使用curl测试API：
```bash
# 启动任务并查看响应
curl -v -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{"video_id": "BV1xx411c7mu", "page_limit": 1}'
```

### 前端调试

浏览器开发者工具：
- **Console**: 查看JavaScript日志和错误
- **Network**: 查看API请求和响应
- **Elements**: 检查DOM结构和样式

在 `static/js/app.js` 中添加调试日志：
```javascript
console.log('API Response:', response);
```

---

## 常见问题

### Q: 如何获取Bilibili的SESSDATA Cookie?

A:
1. 登录 bilibili.com
2. 打开浏览器开发者工具（F12）
3. 切换到 Application/存储 标签页
4. 查看 Cookies -> https://www.bilibili.com
5. 找到 SESSDATA 字段并复制其值

### Q: 为什么评论抓取很慢？

A:
- 默认每次请求间隔300ms，这是为了避免触发API限流
- 如果开启子评论抓取，每条有回复的评论会增加一次请求
- 可以适当增加 `delay_ms` 参数以提高稳定性

### Q: 如何处理"too many requests"错误？

A:
1. 增加 `delay_ms` 参数（如500-1000ms）
2. 减少 `page_limit` 参数
3. 使用Cookie认证可以获得更高的限额

### Q: Excel导出的中文乱码怎么办？

A:
- 使用Excel打开时选择UTF-8编码
- 或者导出为CSV格式（已包含UTF-8 BOM）

### Q: 如何添加新的Bilibili API接口？

A:
1. 在 `pkg/bilibili/models.go` 定义响应结构体
2. 在对应文件（如 `video.go`）添加API函数
3. 使用 `BilibiliClient.Get()` 发送请求
4. 如需WBI签名，使用 `AddWbiSign()` 处理参数
5. 添加测试和文档

### Q: 如何修改服务器端口？

A:
编辑 `configs/config.json`:
```json
{
  "server": {
    "port": 8080
  }
}
```

或设置环境变量：
```bash
export PORT=3000
```

### Q: 前端修改后如何刷新？

A:
静态文件直接被Gin serve，修改后刷新浏览器即可（Ctrl+F5 强制刷新）。

---

## 性能优化建议

### 后端

1. **并发请求**: 使用goroutine并发获取评论
   ```go
   var wg sync.WaitGroup
   for page := 1; page <= limit; page++ {
       wg.Add(1)
       go func(p int) {
           defer wg.Done()
           // 获取评论
       }(page)
   }
   wg.Wait()
   ```

2. **缓存**: 缓存视频信息和用户信息
   ```go
   var cache sync.Map
   cache.LoadOrStore(key, value)
   ```

3. **连接池**: HTTP客户端使用连接池（已默认启用）

### 前端

1. **虚拟滚动**: 对于大量评论，使用虚拟滚动渲染
2. **防抖**: 搜索输入使用防抖
   ```javascript
   let timeout;
   input.addEventListener('input', () => {
       clearTimeout(timeout);
       timeout = setTimeout(() => {
           // 执行搜索
       }, 300);
   });
   ```

---

## 部署

### 生产环境构建

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/app ./cmd/app

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/app.exe ./cmd/app

# 优化构建（减小体积）
go build -ldflags="-s -w" -o bin/app ./cmd/app
```

### Docker部署

创建 `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app ./cmd/app

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .
COPY --from=builder /app/static ./static
COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./app"]
```

构建和运行：
```bash
docker build -t bilibili-scraper .
docker run -p 8080:8080 bilibili-scraper
```

---

## 贡献指南

欢迎贡献！请遵循以下步骤：

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

请确保：
- 代码通过所有测试
- 遵循代码规范
- 更新相关文档
- 添加变更日志

---

## 参考资源

- [Go官方文档](https://golang.org/doc/)
- [Gin框架文档](https://gin-gonic.com/docs/)
- [Bilibili API文档](https://github.com/SocialSisterYi/bilibili-API-collect)
- [Excelize文档](https://xuri.me/excelize/)

---

## 联系方式

如有问题或建议，请通过以下方式联系：
- 提交Issue
- 发起Discussion
- 查看项目README

---

**祝开发愉快！🚀**
