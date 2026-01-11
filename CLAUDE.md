# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此仓库中工作时提供指导。

## 构建和运行命令

```bash
# 构建项目
go build -o bin/app ./cmd/app

# 运行 Web 服务器
go run ./cmd/app

# 运行示例
go run examples/bilibili_example.go
go run examples/file_example.go

# 运行测试
go test ./...

# 运行单个测试
go test -v ./pkg/bilibili -run TestFunctionName
```

## 架构概述

这是一个用于访问 Bilibili 公开 API 的 Go 项目，同时提供可选的 HTTP 服务器。

### 核心组件

**Bilibili API 客户端** (`pkg/bilibili/`)
- `client.go` - 支持 Cookie/APP 认证的 HTTP 客户端
- `wbi.go` - WBI 签名实现（大部分接口需要此签名）
- `comment.go` - 视频评论获取，支持分页、游标导航和备用接口
- `video.go` - 通过 BVID 或 AID 获取视频信息
- `user.go` - 通过 mid 获取用户信息
- `models.go` - 所有 API 响应结构体

**HTTP 服务器** (`cmd/app/`, `api/`, `internal/`)
- 基于 Gin 框架的 Web 服务器，端口 8080
- 路由定义在 `api/api.go`
- 业务逻辑在 `internal/services/`
- HTTP 处理器在 `internal/handlers/`

**文件工具** (`pkg/file/`)
- 使用 excelize 库进行 Excel 读写
- CSV 读写

### 关键模式

**WBI 签名**: Bilibili 大部分 API 需要 WBI 签名。`wbi.go` 负责：
1. 从 `/x/web-interface/nav` 获取 WBI 密钥
2. 生成带有 `w_rid` 和 `wts` 的签名参数

**认证选项**: 评论 API 使用函数式选项模式：
```go
bilibili.GetComments(oid, pn, ps, next, bilibili.WithCookie(sessdata))
bilibili.GetComments(oid, pn, ps, next, bilibili.WithAppAuth(appkey, appsec))
```

**排序模式**: 评论支持按时间或热度排序：
```go
bilibili.GetComments(oid, pn, ps, next, bilibili.WithSortMode("time"))  // 按时间排序（默认）
bilibili.GetComments(oid, pn, ps, next, bilibili.WithSortMode("hot"))   // 按热度排序
// 可组合使用：
bilibili.GetComments(oid, pn, ps, next, bilibili.WithCookie(sessdata), bilibili.WithSortMode("hot"))
```

**分页机制**: 评论使用游标分页，通过响应中 `Cursor` 字段的 `next`（整数）和 `nextOffset`（字符串）进行翻页，而非传统页码。

## 依赖

- `github.com/gin-gonic/gin` - HTTP 框架
- `github.com/xuri/excelize/v2` - Excel 文件处理