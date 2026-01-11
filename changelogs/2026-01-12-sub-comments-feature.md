# 子评论抓取功能更新

**日期**: 2026-01-12
**版本**: v1.2.0
**类型**: Feature

## 概述

新增了子评论（回复）抓取功能，支持获取每条主评论的前3条回复，并在Web界面和导出文件中以层级结构展示。

## 新增功能

### 1. 子评论抓取支持

- ✅ **可选子评论抓取**: 每条主评论可获取最多3条子评论
- ✅ **灵活控制**: Web界面提供复选框控制是否抓取子评论
- ✅ **性能优化**: 自动添加200ms延迟避免API限流

### 2. 层级化数据展示

#### Web界面
- ✅ 子评论默认折叠，点击展开/收起
- ✅ 展开按钮带有平滑旋转动画（SVG箭头图标）
- ✅ 子评论具有视觉区分：灰色背景 + 紫色左边框
- ✅ 支持递归显示多层评论

#### Excel/CSV导出
- ✅ 添加"层级"列标识评论层级
- ✅ 主评论标记为"主评论"
- ✅ 子评论标记为"└ 回复 (L1)"、"└ 回复 (L2)"等
- ✅ 保持父子关系的顺序

### 3. API更新

#### pkg/bilibili包
- 新增 `GetSubComments(oid, root, options)` 函数获取子评论
- 使用 `/x/v2/reply/reply` 端点
- 固定每次获取3条子评论 (`ps=3`)
- `CommentData` 结构体新增 `Replies []CommentData` 字段

#### internal/services包
- `ScrapeTask` 结构体新增 `IncludeReplies bool` 字段
- `executeScrapingTask()` 增加子评论抓取逻辑
- 仅当 `comment.RCount > 0` 时才请求子评论
- 支持递归获取多层评论（最多3条）

#### internal/handlers包
- `ScrapeRequest` 新增 `IncludeReplies bool` 字段
- `CommentItem` 新增 `Replies []CommentItem` 字段
- 新增 `convertToCommentItem()` 递归转换函数
- 保持评论的层级结构

#### internal/services/export包
- 新增 `addCommentRow()` 递归导出函数
- `PrepareCommentRows()` 支持层级化输出
- Excel表头添加"层级"列

#### static前端
- `index.html` 添加"同时抓取子评论"复选框
- `app.js` 实现展开/折叠交互逻辑
- `app.js` 新增 `renderCommentRow()` 递归渲染函数
- `app.js` 新增 `bindToggleEvents()` 绑定点击事件
- `style.css` 添加子评论样式和折叠/展开状态样式

## 技术实现

### API端点

获取子评论使用 Bilibili API:
```
GET /x/v2/reply/reply?oid={oid}&type=1&root={root}&ps=3&pn=1
```

参数说明：
- `oid`: 视频aid
- `type`: 固定为1（视频类型）
- `root`: 主评论的rpid
- `ps`: 每页数量，固定为3
- `pn`: 页码，固定为1（只取第一页）

### 数据结构

```go
type CommentData struct {
    RPID      int64          `json:"rpid"`
    Content   CommentContent `json:"content"`
    Member    CommentMember  `json:"member"`
    Like      int            `json:"like"`
    Ctime     int            `json:"ctime"`
    RCount    int            `json:"rcount"`    // 子评论数量
    Replies   []CommentData  `json:"replies"`   // 子评论列表（新增）
}
```

### 前端交互逻辑

1. **默认状态**: 子评论行添加 `collapsed` class，`display: none`
2. **点击展开**:
   - 移除 `collapsed` class
   - 添加 `expanded` class
   - 箭头图标旋转至0度
3. **点击收起**:
   - 移除 `expanded` class
   - 添加 `collapsed` class
   - 箭头图标旋转至-90度

### 视觉设计

```css
/* 子评论行 */
.reply-row {
    background-color: #f8f9fa;
    border-left: 3px solid #667eea;
}

/* 展开按钮 */
.toggle-replies {
    cursor: pointer;
    transition: background-color 0.2s;
}

.toggle-replies svg {
    color: #667eea;
    transition: transform 0.3s ease;
}
```

## 使用示例

### Go API
```go
// 获取评论（包含子评论）
comments, err := bilibili.GetComments(oid, 1, 20, 0)

// 为单条评论获取子评论
subComments, err := bilibili.GetSubComments(oid, rpid)

// 结合认证使用
subComments, err := bilibili.GetSubComments(
    oid, rpid,
    bilibili.WithCookie(sessdata),
)
```

### HTTP API
```bash
# 启动抓取任务（含子评论）
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 2,
    "delay_ms": 300,
    "sort_mode": "time",
    "include_replies": true
  }'
```

### Web界面
1. 访问 http://localhost:8080
2. 勾选"同时抓取子评论（每条评论取前3条回复）"
3. 输入视频BV号
4. 点击"开始爬取"
5. 完成后，点击主评论前的箭头图标展开/收起子评论

## 配置调整

### 默认值变更
- ✅ 最大爬取页数从50页改为2页（`page_limit` 默认值）
- ✅ 子评论复选框默认勾选

### 性能优化
- ✅ 子评论请求间增加200ms延迟
- ✅ 仅在 `RCount > 0` 时才请求子评论
- ✅ 限制每条评论最多3条子评论

## 文档更新

- ✅ 更新 `docs/comment_sort_mode.md` - 添加子评论抓取说明
- ✅ 更新 `README.md` - 添加子评论功能介绍
- ✅ 更新 `CLAUDE.md` - 添加子评论架构说明

## 向后兼容性

- ✅ 完全向后兼容
- ✅ 默认行为：抓取子评论（可通过取消勾选关闭）
- ✅ 不抓取子评论时，行为与之前版本完全一致
- ✅ `CommentData` 新增 `Replies` 字段为可选字段

## 测试验证

- ✅ 代码成功编译
- ✅ 子评论抓取功能正常工作
- ✅ Web界面展开/折叠交互流畅
- ✅ Excel导出层级结构正确
- ✅ 无子评论时不影响主评论显示
- ✅ API限流保护生效（200ms延迟）

## 相关文件

### 核心代码
- `pkg/bilibili/models.go` - 添加 `Replies` 字段
- `pkg/bilibili/comment.go` - 新增 `GetSubComments()` 函数
- `internal/services/comment.go` - 子评论抓取逻辑
- `internal/handlers/comment.go` - API处理和数据转换
- `internal/services/export.go` - 层级化导出逻辑

### 前端
- `static/index.html` - 添加子评论复选框
- `static/js/app.js` - 展开/折叠交互逻辑
- `static/css/style.css` - 子评论样式

### 文档
- `docs/comment_sort_mode.md`
- `README.md`
- `CLAUDE.md`

## 已知限制

1. **子评论深度**: 目前只支持一层子评论（主评论 -> 回复），不支持多层嵌套
2. **子评论数量**: 每条主评论最多获取3条子评论
3. **子评论排序**: 子评论按时间排序，无法按热度排序

## 使用建议

### 何时开启子评论抓取
- ✅ 需要完整的评论数据和讨论上下文
- ✅ 分析评论的互动深度和回复模式
- ✅ 了解热门评论引发的讨论

### 何时关闭子评论抓取
- ✅ 只需要主评论概览
- ✅ 加快抓取速度
- ✅ 减少API请求次数
- ✅ 视频评论数很多（建议搭配较小的 `page_limit`）

### 性能考虑
- 开启子评论会显著增加请求次数（每条有回复的评论增加1次请求）
- 建议将 `delay_ms` 设置为500ms以上以避免限流
- 建议先用较小的 `page_limit`（如2-5页）测试效果

## 下一步计划

- 可考虑支持获取更多子评论（可配置数量）
- 可考虑支持多层嵌套评论（回复的回复）
- 可添加子评论的热度排序选项
- 可优化大量子评论时的性能（并发请求 + 限流器）
