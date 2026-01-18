# AI 评论分析功能更新

**日期**: 2026-01-17
**版本**: v1.3.0
**类型**: Feature

## 概述

新增了 AI 评论分析功能，支持调用大模型对抓取到的评论数据进行智能分析。包含独立的分析页面、多种预设 Prompt 模板、自定义 Prompt 支持以及结果预览功能。

## 新增功能

### 1. AI 分析服务

- ✅ **默认使用智谱AI (GLM-4.7)**: 国内大模型，无需代理即可访问
- ✅ **OpenAI 兼容 API**: 支持调用任何兼容 OpenAI 格式的大模型
- ✅ **灵活配置**: 通过环境变量配置 API URL、密钥和模型
- ✅ **模拟模式**: 未配置密钥时返回模拟响应，便于开发调试
- ✅ **超时保护**: 120秒请求超时，防止长时间等待

### 2. Prompt 模板系统

#### 预设模板
- ✅ **评论总结** (`summary`): 生成视频评论的整体总结
- ✅ **情感分析** (`sentiment`): 分析评论的情感倾向和分布
- ✅ **关键词提取** (`keywords`): 提取评论中的关键词和话题
- ✅ **问答分析** (`qa`): 从评论中提取问题和答案
- ✅ **自定义分析** (`custom`): 使用用户自定义的 Prompt

#### 模板变量
- `{{comments}}`: 评论数据列表
- `{{video_title}}`: 视频标题
- `{{comment_count}}`: 评论数量

### 3. 分析页面 (`/analysis`)

#### 配置面板（左侧）
- ✅ 任务选择：显示所有已完成的爬取任务
- ✅ 模板选择：卡片式展示预设模板
- ✅ 自定义 Prompt：文本编辑器（自定义模板时显示）
- ✅ 分析设置：评论数量限制（0=全部）
- ✅ Prompt 预览：查看渲染后的完整 Prompt

#### 结果面板（右侧）
- ✅ Markdown 渲染：使用 marked.js 渲染分析结果
- ✅ 结果操作：复制到剪贴板、下载 Markdown 文件
- ✅ 加载状态：分析中显示加载动画

### 4. API 端点

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/api/analysis/templates` | 获取所有预设 Prompt 模板 |
| GET | `/api/analysis/tasks/completed` | 获取所有已完成的任务列表 |
| GET | `/api/analysis/tasks/:task_id` | 获取指定任务的评论摘要 |
| POST | `/api/analysis/analyze` | 执行评论分析 |
| POST | `/api/analysis/preview` | 预览 Prompt 渲染结果 |

## 技术实现

### 后端架构

```
internal/services/analysis.go      # AI 分析服务
  ├── NewAnalysisService()          # 创建服务（支持环境变量配置）
  ├── AnalyzeComments()             # 执行分析
  ├── GetPresetTemplates()          # 获取预设模板
  ├── renderTemplate()               # 渲染模板
  ├── formatComments()              # 格式化评论数据
  └── callLLM()                    # 调用 LLM API / 返回模拟响应

internal/handlers/analysis.go      # 分析处理器
  ├── GetTemplatesHandler()         # 获取模板
  ├── CompletedTasksHandler()       # 获取任务列表
  ├── GetCommentsForAnalysisHandler()  # 获取评论摘要
  ├── AnalyzeHandler()             # 执行分析
  └── PreviewPromptHandler()        # 预览 Prompt
```

### 前端架构

```
static/analysis.html               # 分析页面
  ├── 配置面板（任务、模板、设置）
  ├── 结果面板（Markdown 渲染）
  └── Prompt 预览模态框

static/js/analysis.js              # 分析逻辑
  ├── AnalysisPage 类
  ├── loadTemplates()              # 加载模板
  ├── loadTasks()                  # 加载任务
  ├── selectTemplate()             # 选择模板
  ├── selectTask()                 # 选择任务
  ├── previewPrompt()              # 预览 Prompt
  ├── startAnalysis()              # 开始分析
  └── downloadMarkdown()           # 下载结果

static/css/analysis.css            # 分析页面样式
  ├── 双栏布局
  ├── Markdown 渲染样式
  └── 模态框样式
```

### 环境变量

配置文件 `configs/config.json`：
```json
{
  "ai": {
    "api_url": "https://open.bigmodel.cn/api/paas/v4/chat/completions",
    "api_key": "your-zhipu-api-key",
    "model": "glm-4.7"
  }
}
```

**获取智谱AI API Key：**
1. 访问 [智谱AI开放平台](https://open.bigmodel.cn/)
2. 注册并登录账户
3. 在 API Keys 管理页面创建 API Key
4. 将密钥填入配置文件的 `api_key` 字段

### 数据结构

```go
// PromptTemplate Prompt模板
type PromptTemplate struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Prompt      string   `json:"prompt"`
    Fields      []string `json:"fields"`
}

// AnalysisRequest 分析请求
type AnalysisRequest struct {
    TaskID       string                 `json:"task_id"`
    VideoTitle   string                 `json:"video_title"`
    Comments     []bilibili.CommentData `json:"comments"`
    Template     string                 `json:"template"`     // 模板ID或自定义Prompt
    CommentLimit int                    `json:"comment_limit"` // 0=全部
}

// AnalysisResult 分析结果
type AnalysisResult struct {
    Analysis  string `json:"analysis"`  // Markdown 格式的分析结果
    TaskID    string `json:"task_id"`
    Timestamp string `json:"timestamp"`
}
```

## 使用示例

### 配置

编辑 `configs/config.json`：

```json
{
  "ai": {
    "api_url": "https://open.bigmodel.cn/api/paas/v4/chat/completions",
    "api_key": "your-zhipu-api-key",
    "model": "glm-4.7"
  }
}
```

或使用其他兼容的 API（如 OpenAI、Claude 等）：
```json
{
  "ai": {
    "api_url": "https://api.openai.com/v1/chat/completions",
    "api_key": "sk-xxx",
    "model": "gpt-3.5-turbo"
  }
}
```

### 启动服务

```bash
go run ./cmd/app
```

### Web 界面使用

1. 访问 http://localhost:8080 进行评论爬取
2. 爬取完成后，访问 http://localhost:8080/analysis
3. 在左侧选择已完成的任务
4. 选择分析模板或自定义 Prompt
5. 可点击"预览 Prompt"查看渲染结果
6. 设置分析评论数量（可选）
7. 点击"开始分析"
8. 在右侧查看分析结果，支持复制或下载

### HTTP API 使用

```bash
# 获取模板列表
curl http://localhost:8080/api/analysis/templates

# 获取已完成任务
curl http://localhost:8080/api/analysis/tasks/completed

# 预览 Prompt
curl -X POST http://localhost:8080/api/analysis/preview \
  -H "Content-Type: application/json" \
  -d '{"task_id": "xxx", "template_id": "summary"}'

# 执行分析
curl -X POST http://localhost:8080/api/analysis/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "xxx",
    "template_id": "summary",
    "comment_limit": 100
  }'

# 使用自定义 Prompt
curl -X POST http://localhost:8080/api/analysis/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "xxx",
    "template_id": "custom",
    "custom_prompt": "请分析这些评论的主要观点：{{comments}}",
    "comment_limit": 0
  }'
```

### JavaScript API 使用

```javascript
// 获取模板
const templates = await API.getTemplates();

// 获取任务列表
const tasks = await API.getCompletedTasks();

// 预览 Prompt
const preview = await API.previewPrompt({
    task_id: 'xxx',
    template_id: 'summary'
});

// 执行分析
const result = await API.analyze({
    task_id: 'xxx',
    template_id: 'summary',
    comment_limit: 100
});

console.log(result.analysis);
```

## 新增文件

### 后端
- `internal/services/analysis.go` - AI 分析服务
- `internal/handlers/analysis.go` - 分析处理器

### 前端
- `static/analysis.html` - 分析页面
- `static/js/analysis.js` - 分析逻辑
- `static/css/analysis.css` - 分析页面样式

### 文档
- `changelogs/2026-01-17-ai-analysis-feature.md` - 本文档

## 修改文件

- `api/api.go` - 新增分析服务和路由
- `static/js/api.js` - 新增 AI 分析 API 方法

## 文档更新

- ✅ 更新 `CLAUDE.md` - 添加 AI 分析架构说明

## 向后兼容性

- ✅ 完全向后兼容
- ✅ 新增功能不影响现有功能
- ✅ 未配置 API 密钥时自动使用模拟模式

## 测试验证

- ✅ 模拟模式正常工作
- ✅ 模板加载和选择正常
- ✅ 任务列表加载正常
- ✅ Prompt 预览功能正常
- ✅ 分析请求发送正常
- ✅ Markdown 渲染正常
- ✅ 复制和下载功能正常

## 已知限制

1. **单次分析**: 目前单次分析最多处理 1000 条评论
2. **超时时间**: AI API 调用超时设置为 120 秒
3. **同步请求**: 分析请求是同步的，大量评论可能需要较长时间
4. **无流式输出**: 不支持流式输出，需要等待完整结果

## 使用建议

### 选择合适的模板

| 场景 | 推荐模板 |
|------|----------|
| 快速了解评论概况 | 评论总结 |
| 了解用户情感倾向 | 情感分析 |
| 识别热门话题 | 关键词提取 |
| 找出用户疑问 | 问答分析 |
| 特定分析需求 | 自定义 Prompt |

### 性能优化

- 对于评论数量较多的任务（>500条），建议设置 `comment_limit` 限制分析数量
- 选择合适的模型：gpt-3.5-turbo 速度较快，gpt-4 效果更好
- 使用预览功能确认 Prompt 是否符合预期后再执行分析

### API 密钥安全

- 不要将 API 密钥提交到代码仓库
- 使用环境变量或配置文件管理密钥
- 考虑使用本地部署的模型保护数据隐私

## 下一步计划

- 可考虑支持流式输出，实时显示分析结果
- 可考虑添加分析历史记录功能
- 可考虑支持更多大模型提供商
- 可考虑添加对比分析功能（多个任务的对比）
- 可考虑添加分析结果导出为 PDF 功能
