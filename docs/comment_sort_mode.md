# 评论排序模式与子评论抓取功能说明

## 功能概述

现在支持以下功能：
1. **排序模式**: 按**时间**或**热度**两种方式抓取 Bilibili 视频评论
2. **子评论抓取**: 可选择同时抓取每条评论的前3条子评论（回复）

这些都是可选功能，默认按时间排序，默认抓取子评论。

## API 使用方法

### 1. 基础 API 使用（pkg/bilibili）

#### 按时间排序（默认）
```go
import "bilibili/pkg/bilibili"

// 方式1: 显式指定按时间排序
comments, err := bilibili.GetComments(oid, 1, 20, 0, bilibili.WithSortMode("time"))

// 方式2: 不指定则默认按时间排序
comments, err := bilibili.GetComments(oid, 1, 20, 0)
```

#### 按热度排序
```go
// 方式1: 使用 WithSortMode 选项（推荐）
comments, err := bilibili.GetComments(oid, 1, 20, 0, bilibili.WithSortMode("hot"))

// 方式2: 使用 GetHotComments（向后兼容，内部调用 WithSortMode）
comments, err := bilibili.GetHotComments(oid, 1, 20)
```

#### 结合认证使用
```go
// 同时设置 Cookie 认证和排序模式
comments, err := bilibili.GetComments(
    oid, 1, 20, 0,
    bilibili.WithCookie(sessdata),
    bilibili.WithSortMode("hot"),
)

// 或使用 APP 认证
comments, err := bilibili.GetComments(
    oid, 1, 20, 0,
    bilibili.WithAppAuth(appkey, appsec),
    bilibili.WithSortMode("hot"),
)
```

#### 批量获取所有评论
```go
// 按时间获取所有评论（默认）
allComments, err := bilibili.GetAllComments(oid)

// 按热度获取所有评论
allComments, err := bilibili.GetAllComments(oid, bilibili.WithSortMode("hot"))

// 带认证按热度获取
allComments, err := bilibili.GetAllComments(
    oid,
    bilibili.WithCookie(sessdata),
    bilibili.WithSortMode("hot"),
)
```

### 2. HTTP API 使用

#### 启动爬取任务

**请求示例：**

```bash
# 按时间排序抓取（默认）
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 10,
    "delay_ms": 300,
    "sort_mode": "time",
    "include_replies": true
  }'

# 按热度排序抓取
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 10,
    "delay_ms": 300,
    "sort_mode": "hot",
    "include_replies": true
  }'

# 不抓取子评论
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 10,
    "delay_ms": 300,
    "sort_mode": "time",
    "include_replies": false
  }'

# 带 Cookie 认证按热度抓取（含子评论）
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "auth_type": "cookie",
    "cookie": "your_sessdata_here",
    "page_limit": 10,
    "delay_ms": 300,
    "sort_mode": "hot",
    "include_replies": true
  }'
```

**请求参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| video_id | string | 是 | - | 视频 BVID |
| sort_mode | string | 否 | "time" | 排序模式：`"time"` 按时间，`"hot"` 按热度 |
| include_replies | boolean | 否 | false | 是否抓取子评论（每条评论取前3条回复） |
| auth_type | string | 否 | "none" | 认证类型：`"none"`、`"cookie"`、`"app"` |
| cookie | string | 否 | - | SESSDATA Cookie（auth_type=cookie 时需要） |
| app_key | string | 否 | - | APP Key（auth_type=app 时需要） |
| app_secret | string | 否 | - | APP Secret（auth_type=app 时需要） |
| page_limit | int | 否 | 50 | 最大抓取页数 |
| delay_ms | int | 否 | 300 | 请求间隔（毫秒） |

**响应示例：**

```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "video_id": "BV1uT4y1P7CX",
  "status": "running",
  "progress": {
    "current_page": 0,
    "total_comments": 0,
    "page_limit": 10
  }
}
```

### 3. Web 界面使用

#### 启动 Web 服务器

```bash
# 编译项目
go build -o bin/app ./cmd/app

# 运行服务器
./bin/app  # Linux/Mac
# 或
bin\app.exe  # Windows

# 服务器默认运行在 http://localhost:8080
```

#### 在浏览器中使用

1. **打开界面**
   - 在浏览器中访问 `http://localhost:8080`
   - 看到"Bilibili 评论爬取与可视化"页面

2. **选择排序模式**

   在"1. 输入视频信息"区域，找到"排序模式"选项：

   - **按时间排序（最新评论）**：默认选项，获取最新发布的评论
   - **按热度排序（热门评论）**：获取点赞数最高的热门评论

3. **选择是否抓取子评论**

   在排序模式下方，有一个复选框：
   - **同时抓取子评论（每条评论取前3条回复）**：勾选后会为每条主评论获取最多3条子评论

4. **配置其他参数**
   - **视频链接或BV号**：输入要爬取的视频ID
   - **认证方式**：可选择无认证、Cookie认证或APP认证
   - **最大爬取页数**：设置要爬取的页数（默认10页）
   - **请求间隔**：设置每次请求间隔时间（默认300毫秒）

5. **开始爬取**
   - 点击"开始爬取"按钮
   - 系统会实时显示爬取进度
   - 完成后自动显示结果

6. **查看结果**

   爬取完成后，界面会显示：
   - **评论时间分布图表**：可视化评论发布时间
   - **点赞数分布图表**：可视化评论热度
   - **评论列表**：详细的评论数据表格
     - 主评论默认显示
     - 子评论默认折叠，点击主评论前的 ▶ 图标可展开/折叠子评论
     - 子评论有浅灰色背景和左侧紫色边框标识
   - **导出功能**：可导出为Excel或CSV格式

#### Web 界面截图示例

**按时间排序示例**：
```
✓ 按时间排序（最新评论）  ← 选中此项
○ 按热度排序（热门评论）
```
获取结果：最新发布的评论按时间倒序排列

**按热度排序示例**：
```
○ 按时间排序（最新评论）
✓ 按热度排序（热门评论）  ← 选中此项
```
获取结果：点赞数最高的热门评论

#### Web 界面技术实现

**前端文件**：
- `static/index.html` - 主页面，包含排序模式选择器
- `static/js/app.js` - 应用逻辑，获取排序模式并传递给API
- `static/js/api.js` - API调用封装
- `static/js/charts.js` - 图表渲染
- `static/css/style.css` - 样式文件

**关键代码**：
```javascript
// 获取排序模式
const sortMode = document.querySelector('input[name="sort-mode"]:checked').value;

// 传递给API
const response = await API.startScrape({
    video_id: videoInput,
    sort_mode: sortMode,  // "time" 或 "hot"
    // ... 其他参数
});
```

## 技术实现

### 1. Bilibili API 参数

- **main 端点** (`/x/v2/reply/main`)
  - `mode=2`: 按时间排序
  - `mode=3`: 按热度排序

- **reply 端点** (`/x/v2/reply`) - 备用接口
  - `sort=2`: 按时间倒序排序
  - `sort=1`: 按热度排序

### 2. 代码结构

#### pkg/bilibili/models.go
- `CommentData` 结构体新增 `Replies` 字段存储子评论列表

#### pkg/bilibili/comment.go
- 新增 `CommentOptions` 结构体管理评论请求选项
- 新增 `WithSortMode(sortMode string)` 函数式选项
- 新增 `GetSubComments()` 函数获取子评论（最多3条）
- 更新所有评论函数支持排序模式选项

#### internal/services/comment.go
- `ScrapeTask` 结构体新增 `SortMode` 和 `IncludeReplies` 字段
- `StartScrapeTask` 函数新增 `sortMode` 和 `includeReplies` 参数
- `executeScrapingTask` 在爬取评论时，如果开启子评论抓取，则为每条评论获取前3条子评论

#### internal/handlers/comment.go
- `ScrapeRequest` 结构体新增 `SortMode` 和 `IncludeReplies` 字段
- `ScrapeCommentsHandler` 处理和验证排序模式参数，传递子评论开关

#### static/index.html
- 添加排序模式单选按钮组件
- 添加子评论抓取复选框
- 提供"按时间排序"和"按热度排序"两个选项
- 默认选中"按时间排序"和"同时抓取子评论"

#### static/js/app.js
- `startScraping` 函数获取排序模式和子评论开关值
- 将 `sort_mode` 和 `include_replies` 参数传递给 API 请求

## 使用建议

### 排序模式选择

1. **按时间排序**适合：
   - 想看最新评论
   - 追踪实时讨论动态
   - 按时间顺序分析评论趋势

2. **按热度排序**适合：
   - 想看最受欢迎的评论
   - 了解大众观点
   - 分析热门话题

### 子评论抓取选择

1. **开启子评论抓取**适合：
   - 想要更完整的评论数据
   - 分析评论的讨论深度
   - 了解评论引发的互动内容
   - 研究评论的回复模式

2. **关闭子评论抓取**适合：
   - 只需要主评论信息
   - 加快抓取速度
   - 减少API请求次数
   - 节省存储空间

### 注意事项

1. **排序模式**：
   - 排序模式必须是 `"time"` 或 `"hot"`，其他值会返回错误
   - 不同排序模式可能返回不同数量的评论
   - 热门评论通常点赞数更高

2. **子评论抓取**：
   - 每条主评论最多获取3条子评论
   - 子评论按时间排序（最新的3条）
   - 开启子评论会增加请求次数，建议适当增加延迟时间
   - 子评论获取失败不会影响主评论的抓取

3. **页面交互**：
   - 子评论默认折叠不显示
   - 点击主评论前的 ▶ 图标展开子评论
   - 再次点击 ▼ 图标折叠子评论
   - 图标会有平滑的旋转动画效果

4. **性能优化**：
   - 默认最大爬取页数为2页，可根据需要调整
   - 如果视频评论数很多且开启了子评论抓取，建议将`delay_ms`设置为500ms以上
   - 子评论抓取会为每条有回复的评论增加一次API请求
   - 建议先用小的`page_limit`测试，确认效果后再增加
