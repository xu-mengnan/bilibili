# 新版页面设计与 V2 API 文档

## 概述

本文档记录了 Bilibili 评论爬取工具的全新设计版本，包括重新设计的前端界面和专用的 V2 API 接口。

### 设计理念

**"Cyber-Aqua Glass" 美学**

- **色调**: 青色/蓝绿色渐变（#06B6D4, #22D3EE）替代原版紫色
- **玻璃拟态**: 半透明卡片 + 背景模糊效果
- **动态背景**: 缓慢移动的网格渐变动画
- **现代字体**: Outfit (显示) + Plus Jakarta Sans (正文)
- **发光效果**: 悬停时的光晕阴影
- **平滑动画**: 300ms 缓动过渡

---

## 目录结构

```
static-new/
├── css/
│   ├── style.css      # 主设计系统
│   ├── app.css        # 评论爬取页面样式
│   ├── analysis.css   # AI分析页面样式
│   └── tasks.css      # 任务管理页面样式
├── js/
│   ├── api.js         # V2 API 客户端
│   ├── app.js         # 评论爬取逻辑
│   ├── analysis.js    # AI分析逻辑
│   ├── tasks.js       # 任务管理逻辑
│   └── charts.js      # 图表渲染
├── index-new.html     # 评论爬取页面
├── analysis-new.html  # AI分析页面
└── tasks-new.html     # 任务管理页面
```

---

## 设计系统

### CSS 变量

```css
/* 主色调 - 青色渐变 */
--primary-light: #22D3EE;
--primary: #06B6D4;
--primary-dark: #0891B2;
--primary-darker: #0E7490;

/* 玻璃拟态效果 */
--glass-bg: rgba(255, 255, 255, 0.85);
--glass-border: rgba(255, 255, 255, 0.3);
--glass-blur: blur(20px);

/* 状态颜色 */
--success: #10B981;
--warning: #F59E0B;
--error: #EF4444;

/* 字体 */
--font-display: 'Outfit', sans-serif;
--font-body: 'Plus Jakarta Sans', sans-serif;
```

### 动画

| 动画名称 | 时长 | 用途 |
|---------|------|------|
| fadeInUp | 0.6s | 页面元素进入 |
| shimmer | 3s | Logo 光泽效果 |
| pulse | 1.5s | 状态指示点 |
| spin | 0.8s | 加载旋转 |
| progressShimmer | 1.5s | 进度条流光 |

---

## 页面设计

### 1. 评论爬取页面 (`index-new.html`)

**布局结构:**
```
┌─────────────────────────────────────────────┐
│ Sidebar          │ Main Content             │
│ - Logo           │ - Hero Section            │
│ - 评论爬取 (active)│ - Video Input Card       │
│ - 任务管理        │ - Progress Card           │
│ - AI 分析         │ - Results Card            │
│                  │   ├─ Filter Toolbar        │
│                  │   ├─ Charts                │
│                  │   ├─ Comments List         │
│                  │   └─ Export Section        │
└─────────────────────────────────────────────┘
```

**功能组件:**
- 视频输入框（带链接图标）
- 认证方式选择（None/Cookie/APP）
- 排序模式选择（时间/热度）
- 进度显示（当前页/总评论/已用时间）
- 图表展示（时间分布/点赞分布）
- 评论卡片（可展开子评论）
- 导出功能（Excel/CSV）

### 2. 任务管理页面 (`tasks-new.html`)

**布局结构:**
```
┌─────────────────────────────────────────────┐
│ Sidebar          │ Stats + Task List         │
│ - Logo           │ ┌─────────────────────────┐│
│ - 评论爬取        │ │ Stats Cards (4列)       ││
│ - 任务管理 (active)│ │ 总数/运行/完成/失败     ││
│ - AI 分析         │ └─────────────────────────┘│
│                  │ ┌─────────────────────────┐│
│                  │ │ Task List              ││
│                  │ │ ├─ 状态筛选             ││
│                  │ │ ├─ 任务卡片             ││
│                  │ │ └─ Modal 详情           ││
│                  │ └─────────────────────────┘│
└─────────────────────────────────────────────┘
```

**统计卡片:**
- 图标 + 渐变数值
- 状态颜色区分
- 悬停抬起效果

**任务卡片:**
- 视频标题 + 状态徽章
- 元信息（ID/视频ID/时间/评论数/错误）
- 点击查看详情（模态框）

### 3. AI 分析页面 (`analysis-new.html`)

**布局结构:**
```
┌─────────────────────────────────────────────┐
│ Sidebar          │ Config + Results         │
│ - Logo           │ ┌─────────────────────────┐│
│ - 评论爬取        │ │ Left Panel (配置)       ││
│ - 任务管理        │ │ ├─ Step 1: 选择任务     ││
│ - AI 分析 (active)│ │ ├─ Step 2: 选择模板     ││
│                  │ │ ├─ Step 3: 开始分析     ││
│                  │ └─────────────────────────┘│
│                  │ ┌─────────────────────────┐│
│                  │ │ Right Panel (结果)      ││
│                  │ │ ├─ 分析中动画           ││
│                  │ │ ├─ 流式预览             ││
│                  │ │ ├─ Markdown渲染         ││
│                  │ │ └─ 复制/下载按钮        ││
│                  │ └─────────────────────────┘│
└─────────────────────────────────────────────┘
```

**分析流程:**
1. 选择已完成的任务
2. 选择预设模板或自定义 Prompt
3. 可选：限制分析评论数量
4. 开始分析（流式显示进度）
5. 完成：Markdown 渲染 + 复制/下载

---

## V2 API 接口

### 设计原则

与老版 API (`/api/*`) 不同，V2 API (`/api/v2/*`) 采用更简洁的设计：

| 特性 | 老API | 新API (V2) |
|------|---------------|-----------------|
| 响应格式 | `{data: [...]}` | 直接返回数组 `[...]` |
| SSE格式 | `event: content\ndata: "JSON"` | `data: 文本内容` |
| 完成信号 | `event: done\ndata:` | `data: [DONE]` |
| 错误信号 | `event: error\ndata: ...` | `data: [ERROR] ...` |

### 端点列表

#### 任务管理

**获取所有任务**
```
GET /api/v2/tasks

Response 200:
[
  {
    "task_id": "uuid",
    "video_id": "BV1xxx",
    "video_title": "视频标题",
    "status": "completed",
    "comment_count": 50,
    "start_time": "2026-01-23 10:30",
    "end_time": "2026-01-23 10:31",
    "error": "",
    "progress": {
      "current_page": 2,
      "page_limit": 2,
      "total_comments": 50
    }
  }
]
```

**获取单个任务详情**
```
GET /api/v2/tasks/:id

Response 200:
{
  "task_id": "uuid",
  "video_id": "BV1xxx",
  "video_title": "视频标题",
  "status": "completed",
  "comment_count": 50,
  "start_time": "2026-01-23 10:30:05",
  "end_time": "2026-01-23 10:31:20",
  "error": "",
  "progress": {
    "current_page": 2,
    "page_limit": 2,
    "total_comments": 50
  },
  "comments": [
    {
      "rpid": 12345678,
      "author": "用户名",
      "avatar": "https://...",
      "content": "评论内容",
      "likes": 10,
      "time": "2小时前",
      "level": 5
    }
  ]
}
```

#### 模板管理

**获取所有模板**
```
GET /api/v2/templates

Response 200:
[
  {
    "id": "summary",
    "name": "评论总结",
    "description": "对评论内容进行总结",
    "prompt": "请总结以下评论..."
  },
  {
    "id": "custom",
    "name": "自定义分析",
    "description": "使用你自己编写的分析Prompt",
    "prompt": ""
  }
]
```

#### AI 分析

**流式分析**
```
POST /api/v2/analyze-stream

Request:
{
  "task_id": "uuid",
  "template_id": "summary",
  "custom_prompt": "",
  "comment_limit": 0
}

Response (SSE):
data: 根据评论内容
data: 可以看到以下观点
data: [DONE]
```

**预览 Prompt**
```
POST /api/v2/preview

Request:
{
  "task_id": "uuid",
  "template_id": "summary"
}

Response 200:
{
  "prompt": "渲染后的完整Prompt...",
  "count": 10
}
```

---

## 与老版本对比

### 访问地址

| 页面 | 老版本 | 新版本 |
|------|--------|--------|
| 评论爬取 | `http://localhost:8080/` | `http://localhost:8080/index-new.html` |
| 任务管理 | `http://localhost:8080/tasks` | `http://localhost:8080/tasks-new.html` |
| AI 分析 | `http://localhost:8080/analysis` | `http://localhost:8080/analysis-new.html` |

### API 对比

| 功能 | 老API | 新API |
|------|-------|-------|
| 获取任务列表 | `/api/tasks/all` | `/api/v2/tasks` |
| 获取任务详情 | `/api/comments/progress/:id` | `/api/v2/tasks/:id` |
| 获取模板 | `/api/analysis/templates` | `/api/v2/templates` |
| 流式分析 | `/api/analysis/analyze-stream` | `/api/v2/analyze-stream` |
| 预览 Prompt | `/api/analysis/preview` | `/api/v2/preview` |

### 响应格式对比

**老版本:**
```json
{
  "tasks": [
    { "task_id": "...", ... }
  ]
}
```

**新版本:**
```json
[
  { "task_id": "...", ... }
]
```

### SSE 格式对比

**老版本:**
```
event: content
data: "文本内容"

event: done
data:
```

**新版本:**
```
data: 文本内容
data: [DONE]
```

---

## 技术栈

### 前端
- **框架**: 原生 JavaScript (ES6+)
- **图表**: Chart.js 4.4.1
- **Markdown**: markdown-it 14.0.0
- **字体**: Google Fonts (Outfit + Plus Jakarta Sans)

### 后端
- **框架**: Gin (Go Web Framework)
- **存储**: JSON 文件存储
- **AI**: OpenAI 兼容 API (智谱AI GLM-4 Flash)

---

## 开发说明

### 添加新的设计元素

1. **颜色**: 在 `style.css` 的 `:root` 中添加 CSS 变量
2. **动画**: 在 `style.css` 的 `@keyframes` 中定义
3. **组件**: 在对应的页面 CSS 文件中添加样式

### 修改 API

1. **后端**: 修改 `internal/handlers/v2_api.go`
2. **前端**: 修改 `static-new/js/api.js`

### 添加新页面

1. 创建 HTML 文件
2. 创建页面专属 CSS 文件
3. 创建页面专属 JS 文件
4. 在 `api/api.go` 中添加路由

---

## 性能优化

### 前端
- 懒加载图表库
- 使用 `requestAnimationFrame` 优化动画
- 防抖/节流用户输入

### 后端
- 流式响应，减少首字节时间
- 增大 SSE 缓冲区 (100)
- goroutine 异步处理

---

## 未来改进

- [ ] 添加深色模式切换
- [ ] 添加更多预设模板
- [ ] 支持导出 PDF 报告
- [ ] 添加评论情感分析图表
- [ ] 支持多任务并行分析
- [ ] 添加用户认证系统

---

## 文件清单

### 新建文件
- `internal/handlers/v2_api.go` - V2 API 处理器
- `static-new/css/style.css` - 主设计系统
- `static-new/css/app.css` - 评论爬取样式
- `static-new/css/analysis.css` - 分析页面样式
- `static-new/css/tasks.css` - 任务管理样式
- `static-new/js/api.js` - API 客户端
- `static-new/js/app.js` - 评论爬取逻辑
- `static-new/js/analysis.js` - 分析逻辑
- `static-new/js/tasks.js` - 任务管理逻辑
- `static-new/js/charts.js` - 图表渲染
- `static-new/index-new.html` - 评论爬取页面
- `static-new/analysis-new.html` - 分析页面
- `static-new/tasks-new.html` - 任务管理页面

### 修改文件
- `api/api.go` - 添加 V2 路由组和静态文件路由

---

## 快速开始

### 1. 启动服务器
```bash
cd C:\Users\x_mn\Desktop\go_code\bilibili
go run ./cmd/app
```

### 2. 访问新版页面
```
http://localhost:8080/index-new.html
http://localhost:8080/tasks-new.html
http://localhost:8080/analysis-new.html
```

### 3. 对比老版本
打开两个浏览器标签页，一个访问老版本，一个访问新版本进行对比。

---

## 许可证

本项目仅用于学习和研究目的。
