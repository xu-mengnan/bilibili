# Bilibili

这是一个使用Go语言开发的项目。

## 目录结构

```
.
├── api/              # API相关代码
├── cmd/              # 程序入口
│   └── app/          # 主应用程序
├── configs/          # 配置文件
├── examples/         # 示例代码
├── go.mod            # Go模块文件
├── internal/         # 私有应用程序代码
│   ├── handlers/     # HTTP处理器
│   └── services/     # 业务逻辑层
└── pkg/              # 可被外部引用的公共代码
    ├── bilibili/     # Bilibili API接口
    ├── file/         # 文件处理（Excel、CSV）
    └── utils/        # 工具类
```

## 快速开始

### 构建项目

```bash
go build -o bin/app ./cmd/app
```

### 运行项目

```bash
go run ./cmd/app
```

或者

```bash
./bin/app
```

## API接口

### Web界面
- `GET /` - Web可视化界面首页
- `GET /static/*` - 静态资源（JS、CSS、图片）

### 评论爬取API
- `POST /api/comments/scrape` - 启动评论爬取任务（支持sort_mode参数）
- `GET /api/comments/progress/:task_id` - 获取爬取进度
- `GET /api/comments/result/:task_id` - 获取爬取结果
- `GET /api/comments/stats/:task_id` - 获取评论统计
- `POST /api/comments/export` - 导出评论数据
- `GET /api/download/:file_id` - 下载导出文件

### 其他API
- `GET /hello` - 简单的问候接口
- `GET /user/{id}` - 根据ID获取用户信息

## Bilibili API功能

项目提供了访问Bilibili公开API的功能：

- 获取视频信息：使用`pkg/bilibili/video.go`
- 获取用户信息：使用`pkg/bilibili/user.go`
- 获取评论信息：使用`pkg/bilibili/comment.go`（支持按时间/热度排序）

### 核心特性

#### 1. 评论排序模式
支持两种评论排序方式：
- **按时间排序**（默认）：获取最新发布的评论
- **按热度排序**：获取点赞数最高的热门评论

#### 2. 子评论抓取
- **可选抓取子评论**：每条评论可获取最多3条子评论（回复）
- **灵活控制**：可在页面上选择是否抓取子评论

#### 3. 认证支持
- 无认证：访问公开评论
- Cookie认证：使用SESSDATA访问完整评论
- APP认证：使用APP Key/Secret访问

#### 4. Web可视化界面
- 实时爬取进度显示
- 评论数据可视化图表
- 支持数据筛选和排序
- 支持子评论抓取（点击展开/折叠）
- 导出Excel/CSV格式

### 示例

#### 按时间排序获取评论（默认）
```go
// 获取视频评论 (oid为视频aid, pn为页码, ps为每页数量, next为游标)
comments, err := bilibili.GetComments(123456, 1, 20, 0)
if err != nil {
    log.Fatal("获取评论失败:", err)
}

// 显式指定按时间排序
comments, err := bilibili.GetComments(123456, 1, 20, 0, bilibili.WithSortMode("time"))
```

#### 按热度排序获取评论
```go
// 获取热门评论
comments, err := bilibili.GetComments(123456, 1, 20, 0, bilibili.WithSortMode("hot"))
if err != nil {
    log.Fatal("获取评论失败:", err)
}

// 结合Cookie认证使用
comments, err := bilibili.GetComments(
    123456, 1, 20, 0,
    bilibili.WithCookie("your_sessdata"),
    bilibili.WithSortMode("hot"),
)
```

#### 获取所有评论
```go
// 按时间获取所有评论（默认）
allComments, err := bilibili.GetAllComments(123456)

// 按热度获取所有评论
allComments, err := bilibili.GetAllComments(123456, bilibili.WithSortMode("hot"))
```

#### 其他API示例

获取用户信息：
```go
// 获取用户信息 (mid为用户ID)
user, err := bilibili.GetUser(123456)
if err != nil {
    log.Fatal("获取用户信息失败:", err)
}
```

获取视频信息：
```go
// 通过BVID获取视频信息
video, err := bilibili.GetVideoByBVID("BV1xx411c7mu")
if err != nil {
    log.Fatal("获取视频信息失败:", err)
}
```

## Web界面使用

### 启动服务器

```bash
# 编译
go build -o bin/app ./cmd/app

# 运行（Linux/Mac）
./bin/app

# 运行（Windows）
bin\app.exe
```

服务器默认运行在 `http://localhost:8080`

### 在浏览器中使用

1. 访问 `http://localhost:8080`
2. 输入视频BV号或链接
3. **选择排序模式**：
   - **按时间排序（最新评论）**：获取最新发布的评论
   - **按热度排序（热门评论）**：获取点赞数最高的热门评论
4. **选择是否抓取子评论**：勾选"同时抓取子评论"可获取每条评论的前3条回复
5. 配置其他参数（认证、页数、延迟等）
6. 点击"开始爬取"
7. 查看实时进度和可视化结果
8. 可导出为Excel或CSV格式

### HTTP API使用

#### 启动爬取任务

```bash
# 按时间排序（默认）
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 10,
    "delay_ms": 300,
    "sort_mode": "time",
    "include_replies": true
  }'

# 按热度排序，抓取子评论
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 10,
    "delay_ms": 300,
    "sort_mode": "hot",
    "include_replies": true
  }'
```

## 详细文档

### 功能文档
- [评论排序模式与子评论抓取功能说明](docs/comment_sort_mode.md) - 完整使用指南
- [代码示例](examples/comment_sortmode_example.go) - 排序模式的代码示例

### 开发文档
- [API参考文档](docs/api-reference.md) - 详细的HTTP API接口说明
- [开发指南](docs/development-guide.md) - 开发环境配置、代码规范、开发流程
- [故障排查指南](docs/troubleshooting.md) - 常见问题诊断和解决方案

### 变更日志
- [变更日志目录](changelogs/README.md) - 所有版本的更新记录
- [v1.2.0 - 子评论抓取功能](changelogs/2026-01-12-sub-comments-feature.md)
- [v1.1.0 - 评论排序模式功能](changelogs/2026-01-11-comment-sort-mode.md)

## 使用建议

### 何时使用按时间排序
- 想要查看最新的评论内容
- 追踪实时讨论动态
- 按时间顺序分析评论趋势
- 了解最近用户的反馈

### 何时使用按热度排序
- 想要查看最受欢迎的评论
- 了解大众的主流观点
- 分析热门话题和讨论焦点
- 找出最有价值的评论内容

### 何时使用子评论抓取
- **开启子评论抓取**：
  - 想要更完整的评论数据
  - 分析评论的讨论深度
  - 了解评论引发的互动内容
- **关闭子评论抓取**：
  - 只需要主评论信息
  - 加快抓取速度
  - 减少API请求次数

### 注意事项
- 排序模式必须是 `"time"` 或 `"hot"`
- 不同排序模式可能返回不同的评论集合
- 热门评论通常点赞数较高，但不一定是最新的
- 使用认证（Cookie或APP）可以获取更完整的评论数据
- 开启子评论抓取会增加请求次数，建议适当增加延迟时间（如500ms以上）
- 每条评论最多获取3条子评论
- 子评论在页面上默认折叠，点击主评论前的图标可展开/折叠
- 默认最大爬取页数为2页，可根据需要调整

## HTTP客户端优化

为了提高代码复用性和维护性，项目引入了统一的HTTP客户端：

- 统一的HTTP请求处理：使用`pkg/bilibili/client.go`
- 所有Bilibili API调用都通过该客户端发送请求
- 统一设置User-Agent、超时等参数

## 文件处理功能

项目提供了处理Excel和CSV文件的功能：

- Excel文件读写：使用`pkg/file/excel.go`
- CSV文件读写：使用`pkg/file/csv.go`

### 示例

运行文件处理示例：
```bash
go run examples/file_example.go
```

这将创建并读取示例的Excel和CSV文件。

## 配置

配置文件位于 `configs/config.json`，可以修改服务器端口和其他配置项。