# 评论排序模式功能更新

**日期**: 2026-01-11
**版本**: v1.1.0
**类型**: Feature

## 概述

增加了按热度抓取Bilibili视频评论的功能，现在支持按时间或热度两种排序模式。

## 新增功能

### 1. 评论排序模式支持

- ✅ **按时间排序（默认）**: 获取最新发布的评论
- ✅ **按热度排序**: 获取点赞数最高的热门评论

### 2. Web界面更新

- ✅ 在Web界面添加了排序模式选择器
- ✅ 提供用户友好的单选按钮界面
- ✅ 默认选中"按时间排序"，保持向后兼容

### 3. API更新

#### pkg/bilibili包
- 新增 `CommentOptions` 结构体管理评论请求选项
- 新增 `WithSortMode(sortMode string)` 函数式选项
- 更新所有评论获取函数支持排序模式
- `GetComments()` 支持通过选项设置排序模式
- `GetAllComments()` 支持排序模式参数
- `GetHotComments()` 改为调用 `WithSortMode("hot")`

#### internal/services包
- `ScrapeTask` 结构体新增 `SortMode` 字段
- `StartScrapeTask()` 函数新增 `sortMode` 参数
- `executeScrapingTask()` 传递排序模式到API调用

#### internal/handlers包
- `ScrapeRequest` 结构体新增 `SortMode` 字段
- 添加排序模式验证逻辑（仅允许"time"或"hot"）
- 设置默认值为"time"

#### static前端
- `index.html` 添加排序模式单选按钮组
- `app.js` 获取并传递排序模式参数到API

## 技术实现

### Bilibili API参数映射

**main端点** (`/x/v2/reply/main`):
- `mode=2`: 按时间排序
- `mode=3`: 按热度排序

**reply端点** (`/x/v2/reply`) - 备用接口:
- `sort=2`: 按时间倒序排序
- `sort=1`: 按热度排序

## 使用示例

### Go API
```go
// 按时间排序（默认）
comments, err := bilibili.GetComments(oid, 1, 20, 0)

// 按热度排序
comments, err := bilibili.GetComments(oid, 1, 20, 0, bilibili.WithSortMode("hot"))

// 结合认证使用
comments, err := bilibili.GetComments(
    oid, 1, 20, 0,
    bilibili.WithCookie(sessdata),
    bilibili.WithSortMode("hot"),
)
```

### HTTP API
```bash
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "sort_mode": "hot",
    "page_limit": 10,
    "delay_ms": 300
  }'
```

### Web界面
1. 访问 http://localhost:8080
2. 选择"按热度排序（热门评论）"单选按钮
3. 输入视频BV号
4. 点击"开始爬取"

## 文档更新

- ✅ 创建 `docs/comment_sort_mode.md` - 完整功能说明文档
- ✅ 创建 `examples/comment_sortmode_example.go` - 代码示例
- ✅ 更新 `README.md` - 添加排序模式使用说明
- ✅ 更新 `CLAUDE.md` - 添加排序模式架构说明

## 向后兼容性

- ✅ 完全向后兼容
- ✅ 默认行为保持不变（按时间排序）
- ✅ 现有代码无需修改即可继续使用
- ✅ `AuthOption` 保持为 `CommentOption` 的别名

## 测试

- ✅ 代码成功编译
- ✅ 前端界面更新完成
- ✅ HTTP API参数验证通过

## 相关文件

### 核心代码
- `pkg/bilibili/comment.go`
- `internal/services/comment.go`
- `internal/handlers/comment.go`

### 前端
- `static/index.html`
- `static/js/app.js`

### 文档
- `docs/comment_sort_mode.md`
- `examples/comment_sortmode_example.go`
- `README.md`
- `CLAUDE.md`

## 已知问题

无

## 下一步计划

- 可考虑添加更多排序选项（如按回复数排序）
- 可添加评论筛选功能（如按用户等级筛选）
