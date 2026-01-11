# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 构建和运行命令

```bash
# 构建项目
go build -o bin/app ./cmd/app

# 运行 Web 服务器（端口 8080）
go run ./cmd/app
# 或
./bin/app

# 运行示例
go run examples/bilibili_example.go
go run examples/comment_sortmode_example.go

# 运行测试
go test ./...

# 运行单个测试
go test -v ./pkg/bilibili -run TestFunctionName
```

## 架构概述

这是一个 Bilibili 评论爬取和分析工具，包含：
1. Bilibili API 客户端库（可独立使用）
2. 基于 Gin 的 Web 服务器（提供 RESTful API 和可视化界面）
3. 异步任务管理系统（评论爬取任务）
4. 数据导出功能（Excel/CSV）

### 核心组件

**Bilibili API 客户端** (`pkg/bilibili/`)
- `client.go` - 统一 HTTP 客户端，支持 Cookie/APP 认证
- `wbi.go` - WBI 签名实现（必需）
- `comment.go` - 评论获取，支持：
  - 按时间/热度排序
  - 游标分页
  - 子评论获取（最多3条）
  - 主端点 `/x/v2/reply/main` 和备用端点 `/x/v2/reply`
- `models.go` - 所有 API 响应结构体
- `video.go` - 视频信息获取
- `user.go` - 用户信息获取

**Web 服务器** (`cmd/app/`, `api/`, `internal/`)
- 端口：8080
- 路由定义：`api/api.go`
- 处理器：`internal/handlers/comment.go`
- 服务层：`internal/services/comment.go`（任务管理）、`internal/services/export.go`（导出）
- 静态文件：`static/` - Web 界面（HTML/JS/CSS）

**任务管理系统** (`internal/services/comment.go`)
- `CommentService` 管理所有爬取任务
- 每个任务是独立的 goroutine
- 任务状态：running, completed, failed
- 自动清理 1 小时前的旧任务
- 支持进度查询和结果获取

**导出系统** (`internal/services/export.go`)
- `ExportService` 管理导出文件
- 支持 Excel 和 CSV 格式
- 子评论层级展示（"主评论" / "└ 回复 (L1)"）
- 自动清理 2 小时前的导出文件

### 关键模式和约定

**WBI 签名**（必需）
Bilibili 的大部分 API 都需要 WBI 签名。流程：
1. `GetWBIKey()` 从 `/x/web-interface/nav` 获取密钥
2. `SignParams()` 对查询参数进行签名，添加 `w_rid` 和 `wts`

**函数式选项模式**
评论 API 使用函数式选项支持灵活配置：
```go
// 组合使用多个选项
bilibili.GetComments(oid, 1, 20, 0,
    bilibili.WithCookie(sessdata),      // 认证
    bilibili.WithSortMode("hot"),        // 排序
)
```

可用选项：
- `WithCookie(sessdata)` - Cookie 认证
- `WithAppAuth(key, secret)` - APP 认证
- `WithSortMode("time"|"hot")` - 排序模式

**游标分页机制**
评论 API 不使用传统页码，而是游标：
- `next` (int) - 下一页游标
- `nextOffset` (string) - 备用游标
- 使用 `GetCommentsWithOffset()` 传递这两个值

**子评论抓取**
- 子评论通过独立 API `/x/v2/reply/reply` 获取
- 每条主评论最多获取 3 条子评论
- 子评论数据存储在 `CommentData.Replies` 字段
- 前端默认折叠，点击展开/折叠

**排序模式映射**
- Main 端点：`mode=2` (时间) / `mode=3` (热度)
- Fallback 端点：`sort=2` (时间) / `sort=1` (热度)

### 数据流

**评论爬取流程**：
1. 用户通过 Web 界面或 API 提交爬取请求
2. `CommentHandlers.ScrapeCommentsHandler` 验证参数
3. `CommentService.StartScrapeTask` 创建任务并启动 goroutine
4. `executeScrapingTask` 执行实际爬取：
   - 获取视频信息（获取 AID）
   - 循环调用 `bilibili.GetComments` 获取评论
   - 如果开启子评论，为每条评论调用 `bilibili.GetSubComments`
   - 使用 map 去重（key 为 RPID）
   - 更新进度
5. 任务完成，数据存储在 `ScrapeTask.Comments`
6. 用户通过 `/api/comments/result/:task_id` 获取结果
7. 结果转换为 `CommentItem`（包含子评论层级）

**前端交互**：
- 主评论默认显示
- 子评论默认折叠（`collapsed` 类）
- 点击主评论前的箭头图标切换展开/折叠
- 箭头有 0.3s 旋转动画
- 子评论有浅灰背景和左侧紫色边框

### 重要约定

**默认值**：
- 最大爬取页数：2 页
- 请求间隔：300ms
- 排序模式：time（按时间）
- 子评论抓取：开启
- 子评论延迟：200ms

**API 端点选择**：
- 优先使用 `main` 端点（支持更多功能）
- 如果返回 -403 错误，自动回退到 `reply` 端点

**错误处理**：
- 子评论获取失败不影响主评论
- API 错误通过 `updateTaskError` 记录到任务
- 失败的任务不会自动重试

**性能考虑**：
- 使用 map 去重评论
- 子评论抓取会显著增加请求数（每条评论 +1 请求）
- 建议子评论模式下增加延迟到 500ms+

## 依赖

- `github.com/gin-gonic/gin` - Web 框架
- `github.com/xuri/excelize/v2` - Excel 处理
- `github.com/google/uuid` - UUID 生成

## 项目特色

1. **完整的任务管理系统**：异步爬取、进度追踪、结果查询
2. **灵活的认证方式**：支持无认证、Cookie、APP 三种方式
3. **智能去重**：使用 RPID 作为唯一键去重评论
4. **子评论层级展示**：前端折叠/展开交互，Excel 层级标识
5. **自动资源清理**：定期清理旧任务和导出文件
6. **API 容错**：主端点失败自动切换备用端点
