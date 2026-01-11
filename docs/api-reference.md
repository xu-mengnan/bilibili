# API 参考文档

本文档详细说明了项目提供的所有HTTP API端点。

## 基础信息

- **基础URL**: `http://localhost:8080`
- **默认端口**: 8080
- **内容类型**: `application/json`
- **字符编码**: UTF-8

## 目录

- [Web界面端点](#web界面端点)
- [评论爬取API](#评论爬取api)
- [任务管理API](#任务管理api)
- [数据导出API](#数据导出api)
- [示例API](#示例api)
- [错误响应](#错误响应)

---

## Web界面端点

### GET /

返回Web可视化界面主页。

**响应**:
- `Content-Type: text/html`
- 返回静态HTML页面

**示例**:
```bash
curl http://localhost:8080/
```

### GET /static/*

访问静态资源（JS、CSS、图片等）。

**路径参数**:
- `*` - 静态文件路径（如 `js/app.js`, `css/style.css`）

**示例**:
```bash
curl http://localhost:8080/static/js/app.js
```

---

## 评论爬取API

### POST /api/comments/scrape

启动一个评论爬取任务。

**请求体**:
```json
{
  "video_id": "BV1uT4y1P7CX",
  "auth_type": "none",
  "cookie": "",
  "app_key": "",
  "app_secret": "",
  "page_limit": 2,
  "delay_ms": 300,
  "sort_mode": "time",
  "include_replies": true
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| video_id | string | 是 | - | Bilibili视频BV号（如"BV1xx411c7mu"）或完整URL |
| auth_type | string | 否 | "none" | 认证类型：`none`（无认证）、`cookie`（Cookie认证）、`app`（APP认证） |
| cookie | string | 否 | "" | SESSDATA Cookie值（auth_type为cookie时必填） |
| app_key | string | 否 | "" | APP Key（auth_type为app时必填） |
| app_secret | string | 否 | "" | APP Secret（auth_type为app时必填） |
| page_limit | integer | 否 | 2 | 最大爬取页数（1-100） |
| delay_ms | integer | 否 | 300 | 请求间隔毫秒数（100-5000） |
| sort_mode | string | 否 | "time" | 排序模式：`time`（按时间）、`hot`（按热度） |
| include_replies | boolean | 否 | false | 是否抓取子评论（每条评论最多3条回复） |

**响应**:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "video_id": "BV1uT4y1P7CX",
  "status": "running",
  "progress": {
    "current_page": 0,
    "total_comments": 0,
    "page_limit": 2
  }
}
```

**响应字段**:
- `task_id` - 任务唯一标识符（UUID）
- `video_id` - 视频BV号
- `status` - 任务状态（"running"、"completed"、"failed"）
- `progress` - 任务进度信息

**错误响应**:
- `400 Bad Request` - 请求参数无效
- `500 Internal Server Error` - 服务器内部错误

**示例**:
```bash
# 按时间排序，不抓取子评论
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 2,
    "delay_ms": 300,
    "sort_mode": "time",
    "include_replies": false
  }'

# 按热度排序，抓取子评论
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 5,
    "delay_ms": 500,
    "sort_mode": "hot",
    "include_replies": true
  }'

# 使用Cookie认证
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

---

## 任务管理API

### GET /api/comments/progress/:task_id

获取指定任务的爬取进度。

**路径参数**:
- `task_id` - 任务ID（由爬取接口返回）

**响应**:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "progress": {
    "current_page": 5,
    "total_comments": 487,
    "page_limit": 10
  },
  "video_title": "【视频标题】",
  "start_time": "2026-01-12T10:30:00Z",
  "elapsed_seconds": 15,
  "error": ""
}
```

**响应字段**:
- `task_id` - 任务ID
- `status` - 任务状态：
  - `running` - 正在爬取
  - `completed` - 爬取完成
  - `failed` - 爬取失败
- `progress` - 进度信息
  - `current_page` - 当前页数
  - `total_comments` - 已获取评论总数
  - `page_limit` - 最大页数限制
- `video_title` - 视频标题
- `start_time` - 任务开始时间（RFC3339格式）
- `elapsed_seconds` - 已耗时（秒）
- `error` - 错误信息（失败时）

**错误响应**:
- `404 Not Found` - 任务不存在

**示例**:
```bash
curl http://localhost:8080/api/comments/progress/550e8400-e29b-41d4-a716-446655440000
```

### GET /api/comments/result/:task_id

获取指定任务的爬取结果（评论列表）。

**路径参数**:
- `task_id` - 任务ID

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| sort | string | 否 | "time_desc" | 排序方式：`time_desc`（时间降序）、`time_asc`（时间升序）、`like_desc`（点赞降序）、`like_asc`（点赞升序） |
| keyword | string | 否 | "" | 搜索关键词（搜索评论内容和用户名） |
| limit | integer | 否 | 1000 | 返回评论数量限制 |

**响应**:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_count": 487,
  "comments": [
    {
      "rpid": 123456789,
      "author": "用户名",
      "avatar": "https://i0.hdslb.com/bfs/face/xxx.jpg",
      "content": "评论内容",
      "likes": 125,
      "time": "2026-01-12 10:30:00",
      "level": 5,
      "replies": [
        {
          "rpid": 123456790,
          "author": "回复者",
          "avatar": "https://i0.hdslb.com/bfs/face/yyy.jpg",
          "content": "回复内容",
          "likes": 10,
          "time": "2026-01-12 10:35:00",
          "level": 3
        }
      ]
    }
  ]
}
```

**响应字段**:
- `task_id` - 任务ID
- `total_count` - 评论总数
- `comments` - 评论列表（包含子评论）
  - `rpid` - 评论ID
  - `author` - 作者用户名
  - `avatar` - 头像URL
  - `content` - 评论内容
  - `likes` - 点赞数
  - `time` - 发布时间（格式：YYYY-MM-DD HH:MM:SS）
  - `level` - 用户等级
  - `replies` - 子评论列表（结构同主评论）

**错误响应**:
- `400 Bad Request` - 任务未完成或参数错误
- `404 Not Found` - 任务不存在

**示例**:
```bash
# 获取所有评论（默认按时间降序）
curl http://localhost:8080/api/comments/result/550e8400-e29b-41d4-a716-446655440000

# 按点赞数降序
curl http://localhost:8080/api/comments/result/550e8400-e29b-41d4-a716-446655440000?sort=like_desc

# 搜索包含"精彩"的评论
curl http://localhost:8080/api/comments/result/550e8400-e29b-41d4-a716-446655440000?keyword=精彩

# 组合使用
curl "http://localhost:8080/api/comments/result/550e8400-e29b-41d4-a716-446655440000?sort=like_desc&keyword=精彩&limit=100"
```

### GET /api/comments/stats/:task_id

获取指定任务的评论统计信息。

**路径参数**:
- `task_id` - 任务ID

**响应**:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_comments": 487,
  "by_date": {
    "2026-01-10": 125,
    "2026-01-11": 234,
    "2026-01-12": 128
  },
  "by_likes": {
    "0-10": 350,
    "11-50": 100,
    "51-100": 25,
    "100+": 12
  },
  "top_keywords": []
}
```

**响应字段**:
- `task_id` - 任务ID
- `total_comments` - 评论总数
- `by_date` - 按日期分布（日期 -> 数量）
- `by_likes` - 按点赞数分布
  - `0-10` - 0-10赞
  - `11-50` - 11-50赞
  - `51-100` - 51-100赞
  - `100+` - 100赞以上
- `top_keywords` - 热门关键词（暂未实现）

**错误响应**:
- `400 Bad Request` - 任务未完成
- `404 Not Found` - 任务不存在

**示例**:
```bash
curl http://localhost:8080/api/comments/stats/550e8400-e29b-41d4-a716-446655440000
```

---

## 数据导出API

### POST /api/comments/export

导出评论数据为Excel或CSV文件。

**请求体**:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "format": "xlsx",
  "sort": "time_desc",
  "filename": "bilibili_comments"
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| task_id | string | 是 | - | 任务ID |
| format | string | 是 | - | 导出格式：`xlsx`（Excel）、`csv`（CSV） |
| sort | string | 否 | "time_desc" | 排序方式（同result接口） |
| filename | string | 否 | "comments" | 文件名（不含扩展名） |

**响应**:
```json
{
  "file_id": "export_20260112_103045_abc123",
  "filename": "bilibili_comments.xlsx",
  "download_url": "/api/download/export_20260112_103045_abc123",
  "created_at": "2026-01-12T10:30:45Z"
}
```

**响应字段**:
- `file_id` - 文件唯一标识符
- `filename` - 文件名
- `download_url` - 下载URL
- `created_at` - 创建时间（RFC3339格式）

**Excel格式说明**:
- **列**: 层级、评论ID、用户ID、用户名、等级、评论内容、点赞数、评论时间
- **层级标识**:
  - 主评论: "主评论"
  - 子评论: "└ 回复 (L1)"、"└ 回复 (L2)" 等

**CSV格式说明**:
- 编码: UTF-8 BOM
- 列同Excel格式

**错误响应**:
- `400 Bad Request` - 参数错误或任务未完成
- `500 Internal Server Error` - 导出失败

**示例**:
```bash
# 导出为Excel
curl -X POST http://localhost:8080/api/comments/export \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "format": "xlsx",
    "filename": "my_comments"
  }'

# 导出为CSV，按点赞数排序
curl -X POST http://localhost:8080/api/comments/export \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "format": "csv",
    "sort": "like_desc",
    "filename": "top_comments"
  }'
```

### GET /api/download/:file_id

下载导出的文件。

**路径参数**:
- `file_id` - 文件ID（由export接口返回）

**响应**:
- `Content-Type`: 根据文件格式设置
  - Excel: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
  - CSV: `text/csv; charset=utf-8`
- `Content-Disposition`: `attachment; filename="..."`
- 文件二进制内容

**错误响应**:
- `404 Not Found` - 文件不存在

**示例**:
```bash
# 下载文件
curl -O http://localhost:8080/api/download/export_20260112_103045_abc123

# 使用wget
wget http://localhost:8080/api/download/export_20260112_103045_abc123
```

---

## 示例API

这些是项目中的示例端点，可用于测试。

### GET /hello

简单的问候接口。

**响应**:
```json
{
  "message": "Hello, World!"
}
```

**示例**:
```bash
curl http://localhost:8080/hello
```

### GET /user/:id

根据ID获取用户信息（示例端点）。

**路径参数**:
- `id` - 用户ID

**响应**:
```json
{
  "id": "123",
  "name": "User 123"
}
```

**示例**:
```bash
curl http://localhost:8080/user/123
```

---

## 错误响应

所有API端点在出错时返回统一的错误格式。

**错误响应格式**:
```json
{
  "error": "错误信息描述"
}
```

**HTTP状态码**:

| 状态码 | 说明 |
|--------|------|
| 200 OK | 请求成功 |
| 400 Bad Request | 请求参数错误或无效 |
| 404 Not Found | 资源不存在（任务ID、文件ID等） |
| 500 Internal Server Error | 服务器内部错误 |

**常见错误示例**:

```json
// 无效的排序模式
{
  "error": "Invalid sort_mode: must be 'time' or 'hot'"
}

// 任务不存在
{
  "error": "task not found"
}

// 任务未完成
{
  "error": "Task not completed yet"
}

// 参数验证失败
{
  "error": "Invalid request: Key: 'ScrapeRequest.VideoID' Error:Field validation for 'VideoID' failed on the 'required' tag"
}
```

---

## 完整使用流程示例

以下是一个完整的API调用流程示例：

```bash
# 1. 启动爬取任务
TASK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1uT4y1P7CX",
    "page_limit": 5,
    "delay_ms": 300,
    "sort_mode": "hot",
    "include_replies": true
  }')

# 提取任务ID
TASK_ID=$(echo $TASK_RESPONSE | jq -r '.task_id')
echo "任务ID: $TASK_ID"

# 2. 轮询任务进度
while true; do
  PROGRESS=$(curl -s http://localhost:8080/api/comments/progress/$TASK_ID)
  STATUS=$(echo $PROGRESS | jq -r '.status')
  echo "状态: $STATUS"

  if [ "$STATUS" = "completed" ]; then
    break
  elif [ "$STATUS" = "failed" ]; then
    echo "任务失败"
    exit 1
  fi

  sleep 2
done

# 3. 获取评论结果
curl -s "http://localhost:8080/api/comments/result/$TASK_ID?sort=like_desc&limit=10" \
  | jq '.comments[] | {author, content, likes}'

# 4. 获取统计信息
curl -s http://localhost:8080/api/comments/stats/$TASK_ID | jq '.'

# 5. 导出为Excel
EXPORT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/comments/export \
  -H "Content-Type: application/json" \
  -d "{
    \"task_id\": \"$TASK_ID\",
    \"format\": \"xlsx\",
    \"sort\": \"like_desc\",
    \"filename\": \"bilibili_hot_comments\"
  }")

FILE_ID=$(echo $EXPORT_RESPONSE | jq -r '.file_id')
echo "文件ID: $FILE_ID"

# 6. 下载文件
curl -O http://localhost:8080/api/download/$FILE_ID
```

---

## 注意事项

### 性能建议

1. **请求间隔**: 建议设置 `delay_ms` >= 300ms，避免触发Bilibili API限流
2. **子评论抓取**: 开启子评论会显著增加请求次数，建议 `delay_ms` >= 500ms
3. **页数限制**: 默认 `page_limit` = 2，根据实际需求调整
4. **并发限制**: 避免同时运行过多爬取任务

### 认证说明

1. **无认证**: 可访问公开评论，但可能受到更严格的限流
2. **Cookie认证**: 使用浏览器的SESSDATA Cookie，获取完整评论数据
3. **APP认证**: 使用官方APP的Key和Secret（需自行获取）

### 数据说明

1. **评论顺序**:
   - `time` 模式: 按发布时间倒序（最新评论在前）
   - `hot` 模式: 按热度（点赞数）倒序
2. **子评论限制**: 每条主评论最多返回3条子评论
3. **子评论排序**: 子评论按时间排序（最新的3条）
4. **游标分页**: 内部使用Bilibili的游标分页机制，自动处理翻页

### 文件管理

1. **导出文件**: 存储在 `exports/` 目录
2. **文件清理**: 需要手动清理过期的导出文件
3. **文件命名**: `export_{时间戳}_{随机字符串}.{扩展名}`

---

## 更新日志

- **2026-01-12**: 添加子评论抓取功能
- **2026-01-11**: 添加排序模式功能
- **初版**: 基础评论爬取功能
