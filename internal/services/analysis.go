package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"bilibili/pkg/bilibili"
)

// AnalysisService AI分析服务
type AnalysisService struct {
	httpClient *http.Client
	apiURL     string
	apiKey     string
	model      string
}

// NewAnalysisService 创建分析服务
func NewAnalysisService(apiURL, apiKey, model string) *AnalysisService {
	// 默认使用智谱AI API（OpenAI兼容格式）
	if apiURL == "" {
		apiURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	}
	// 默认使用更快的模型
	if model == "" {
		model = "glm-4-flash" // 闪速版，响应更快
	}

	transport := &http.Transport{
		Proxy:                 nil,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		ForceAttemptHTTP2:     false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second, // 增加响应头超时，等待API处理
		DisableKeepAlives:     false,
		MaxConnsPerHost:       10,
		// 禁用压缩，减少处理开销
		DisableCompression: false,
		// 设置连接超时
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &AnalysisService{
		httpClient: &http.Client{
			Timeout:   180 * time.Second, // 增加总超时
			Transport: transport,
		},
		apiURL: apiURL,
		apiKey: apiKey,
		model:  model,
	}
}

// PromptTemplate Prompt模板
type PromptTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Prompt      string   `json:"prompt"`
	Fields      []string `json:"fields"` // 可用字段: comments, video_title, comment_count
}

// 预设的Prompt模板
var presetTemplates = []PromptTemplate{
	{
		ID:          "summary",
		Name:        "评论总结",
		Description: "生成视频评论的整体总结",
		Prompt: `请分析以下视频的评论数据，提供一个全面的分析总结：

## 视频信息
视频标题：{{video_title}}
评论数量：{{comment_count}}条

## 评论数据
{{comments}}

## 分析要求
请从以下几个方面进行分析：
1. **整体情感倾向**：评论整体是正面、负面还是中性
2. **主要观点**：用户评论中的主要观点和看法
3. **热门话题**：评论中出现频率较高的话题或关键词
4. **用户反馈**：用户对视频内容的具体意见和建议
5. **总结**：用3-5句话总结评论的总体情况

## 输出格式要求
请严格按照以下 Markdown 格式输出（注意每个标题和列表项前要有换行）：

# 分析总结

1. **整体情感倾向**
（在这里描述整体情感倾向）

2. **主要观点**
**观点一**：详细描述
**观点二**：详细描述

3. **热门话题**
- 话题一
- 话题二

4. **用户反馈**
正面反馈：具体内容
负面反馈：具体内容

5. **总结**
总结内容`,
		Fields: []string{"comments", "video_title", "comment_count"},
	},
	{
		ID:          "sentiment",
		Name:        "情感分析",
		Description: "分析评论的情感倾向",
		Prompt: `请对以下视频评论进行情感分析：

## 视频信息
视频标题：{{video_title}}

## 评论数据
{{comments}}

## 分析要求
1. 统计正面、负面、中性评论的比例
2. 提取每类评论的典型代表（各2-3条）
3. 分析情感分布的原因
4. 给出可视化数据表格

## 输出格式要求
请严格按照以下 Markdown 格式输出：

# 情感分析报告

1. **统计正面、负面、中性评论的比例**

| 情感类别 | 评论数量 | 比例 |
| --- | --- | --- |
| 正面 | 数字 | 百分比 |
| 负面 | 数字 | 百分比 |
| 中性 | 数字 | 百分比 |

2. **提取每类评论的典型代表**

**正面评论**：
- 评论内容

**负面评论**：
- 评论内容

注意：表格的每一行都要单独一行，不要连在一起。`,
		Fields: []string{"comments", "video_title"},
	},
	{
		ID:          "keywords",
		Name:        "关键词提取",
		Description: "提取评论中的关键词和话题",
		Prompt: `请提取以下视频评论中的关键词和话题：

## 视频信息
视频标题：{{video_title}}

## 评论数据
{{comments}}

## 分析要求
1. 提取主要关键词（按重要性排序，至少20个）
2. 识别主要话题/讨论点（至少5个）
3. 统计每个关键词/话题的出现频率
4. 给出关键词云（用不同大小表示频率）

请用Markdown格式输出。`,
		Fields: []string{"comments", "video_title"},
	},
	{
		ID:          "qa",
		Name:        "问答分析",
		Description: "从评论中提取问题和答案",
		Prompt: `请从以下视频评论中提取问题和相关的答案：

## 视频信息
视频标题：{{video_title}}

## 评论数据
{{comments}}

## 分析要求
1. 提取评论中的问题
2. 为每个问题找到相关的回复或答案
3. 如果没有答案，标注为"待解答"
4. 按问题类型分类整理

请用Markdown格式输出，使用列表格式。`,
		Fields: []string{"comments", "video_title"},
	},
	{
		ID:          "custom",
		Name:        "自定义分析",
		Description: "使用自定义Prompt进行分析",
		Prompt: `请根据以下视频评论进行分析：

## 视频信息
视频标题：{{video_title}}
评论数量：{{comment_count}}条

## 评论数据
{{comments}}

请在下方添加你的分析要求：
---
（请在此处输入你的分析要求）
---

请用Markdown格式输出结果。`,
		Fields: []string{"comments", "video_title", "comment_count"},
	},
}

// GetPresetTemplates 获取预设的Prompt模板
func (s *AnalysisService) GetPresetTemplates() []PromptTemplate {
	return presetTemplates
}

// GetTemplateByID 根据ID获取模板
func (s *AnalysisService) GetTemplateByID(id string) *PromptTemplate {
	for _, t := range presetTemplates {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

// AnalyzeComments 分析评论
func (s *AnalysisService) AnalyzeComments(req *AnalysisRequest) (*AnalysisResult, error) {
	// 格式化评论数据
	commentsText := s.formatComments(req.Comments, req.CommentLimit)

	// 获取模板
	template := req.Template
	if template == "" {
		template = s.renderTemplate(presetTemplates[0], commentsText, req.VideoTitle, len(req.Comments))
	} else {
		// 检查是否是预设模板ID
		if t := s.GetTemplateByID(template); t != nil {
			template = s.renderTemplate(*t, commentsText, req.VideoTitle, len(req.Comments))
		}
	}

	// 调用大模型API
	response, err := s.callLLM(template)
	if err != nil {
		return nil, fmt.Errorf("LLM API调用失败: %w", err)
	}

	return &AnalysisResult{
		Analysis:  response,
		TaskID:    req.TaskID,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// renderTemplate 渲染模板
func (s *AnalysisService) renderTemplate(template PromptTemplate, commentsText, videoTitle string, commentCount int) string {
	result := template.Prompt
	result = strings.ReplaceAll(result, "{{comments}}", commentsText)
	result = strings.ReplaceAll(result, "{{video_title}}", videoTitle)
	result = strings.ReplaceAll(result, "{{comment_count}}", fmt.Sprintf("%d", commentCount))
	return result
}

// formatComments 格式化评论数据
func (s *AnalysisService) formatComments(comments []bilibili.CommentData, limit int) string {
	if limit > 0 && limit < len(comments) {
		comments = comments[:limit]
	}

	var builder strings.Builder
	builder.WriteString("```\n")
	for i, c := range comments {
		builder.WriteString(fmt.Sprintf("[%d] %s (点赞:%d)\n", i+1, c.Content.Message, c.Like))
		if len(c.Replies) > 0 {
			builder.WriteString("    回复:\n")
			for _, r := range c.Replies {
				builder.WriteString(fmt.Sprintf("    - %s\n", r.Content.Message))
			}
		}
		builder.WriteString("\n")
	}
	builder.WriteString("```\n")
	return builder.String()
}

// LLMRequest LLM API请求
type LLMRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse LLM API响应
type LLMResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// callLLM 调用LLM API（流式输出）
func (s *AnalysisService) callLLM(prompt string) (string, error) {
	if s.apiKey == "" {
		return s.getMockResponse(prompt), nil
	}

	reqBody := LLMRequest{
		Model: s.model,
		Messages: []Message{
			{Role: "system", Content: "你是一个专业的数据分析师，擅长分析社交媒体评论。"},
			{Role: "user", Content: prompt},
		},
		Stream: true, // 启用流式输出
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API返回错误: %s, Body: %s", resp.Status, string(body))
	}

	// 读取流式响应
	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 6 || line[:6] != "data: " {
			continue
		}
		data := line[6:]
		if data == "[DONE]" {
			break
		}

		var streamResp struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
			fullContent.WriteString(streamResp.Choices[0].Delta.Content)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取流失败: %w", err)
	}

	result := fullContent.String()
	if result == "" {
		return "", fmt.Errorf("API返回空响应")
	}

	return result, nil
}

// getMockResponse 返回模拟响应（当未配置API密钥时）
func (s *AnalysisService) getMockResponse(prompt string) string {
	return `# 评论分析结果（模拟）

> 注意：这是模拟响应。如需使用真实AI分析，请配置API密钥。

## 整体情感倾向
基于分析，评论整体呈现 **中性偏正面** 的倾向。大多数用户对视频内容表示认可，同时也提出了一些建设性的建议。

## 主要观点
1. **内容质量**：用户普遍认为视频制作精良，信息量充足
2. **节奏把控**：部分用户认为视频节奏适中，易于跟随
3. **实用价值**：很多用户表示视频内容对实际应用有帮助

## 热门话题
- 视频制作技巧
- 内容优化建议
- 后续期待

## 用户反馈
- "视频讲解很清晰，受益匪浅"
- "希望能有更多类似的教程"
- "节奏可以稍微慢一点"

## 总结
该视频获得了用户的积极反馈，整体评价良好。建议继续保持当前的创作风格，同时可以考虑增加更多互动环节。

---
*以上为模拟分析结果。要获取真实的AI分析，请在` + "`" + `configs/config.json` + "`" + `中配置` + "`" + `api_key` + "`" + `字段。*`
}

// ChunkCallback 流式输出回调函数类型
type ChunkCallback func(chunk string)

// CallLLMStream 调用LLM API（流式输出，带回调）
func (s *AnalysisService) CallLLMStream(callback ChunkCallback, prompt string) (string, error) {
	if s.apiKey == "" {
		// 模拟响应也使用流式输出，每次发送累积的完整内容
		mockResponse := s.getMockResponse(prompt)
		chunks := splitIntoChunks(mockResponse, 20) // 每20个字符为一个chunk
		var accumulated strings.Builder
		for _, chunk := range chunks {
			accumulated.WriteString(chunk)
			callback(accumulated.String()) // 发送累积的完整内容
		}
		return mockResponse, nil
	}

	startTime := time.Now()

	reqBody := LLMRequest{
		Model: s.model,
		Messages: []Message{
			{Role: "system", Content: "你是一个专业的数据分析师，擅长分析社交媒体评论。"},
			{Role: "user", Content: prompt},
		},
		Stream: true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	fmt.Printf("[LLM] 请求体大小: %d bytes\n", len(jsonBody))

	httpReq, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")

	fmt.Printf("[LLM] 发起请求到: %s\n", s.apiURL)
	reqSendTime := time.Now()

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respRecvTime := time.Now()
	fmt.Printf("[LLM] 响应时间: %v, 状态码: %d\n", respRecvTime.Sub(reqSendTime), resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API返回错误: %s, Body: %s", resp.Status, string(body))
	}

	// 使用原始 reader 读取流，避免 bufio.Scanner 的缓冲延迟
	var fullContent strings.Builder
	reader := bufio.NewReader(resp.Body)
	firstChunkTime := time.Time{}
	chunkCount := 0

	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("读取流失败: %w", err)
		}

		// 处理超长行（虽然 SSE 通常不会有超长行）
		if isPrefix {
			for {
				_, isPrefix, err := reader.ReadLine()
				if err != nil || !isPrefix {
					break
				}
			}
			continue
		}

		// 解析 SSE 格式
		lineStr := string(line)
		if len(lineStr) < 6 || lineStr[:6] != "data: " {
			continue
		}
		data := lineStr[6:]
		if data == "[DONE]" {
			fmt.Printf("[LLM] 收到 [DONE] 信号\n")
			break
		}

		var streamResp struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
			chunk := streamResp.Choices[0].Delta.Content
			chunkCount++
			if firstChunkTime.IsZero() {
				firstChunkTime = time.Now()
				fmt.Printf("[LLM] 首个 chunk 延迟: %v\n", firstChunkTime.Sub(respRecvTime))
			}
			fullContent.WriteString(chunk)
			// 发送累积的完整内容到前端
			callback(fullContent.String())
		}
	}

	endTime := time.Now()
	fmt.Printf("[LLM] 流式读取完成，总耗时: %v, chunk数: %d, 内容长度: %d\n",
		endTime.Sub(startTime), chunkCount, fullContent.Len())

	result := fullContent.String()
	if result == "" {
		return "", fmt.Errorf("API返回空响应")
	}

	return result, nil
}

// splitIntoChunks 将文本分割成块
func splitIntoChunks(text string, chunkSize int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	for i := 0; i < len(text); i += chunkSize {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}

// FormatComments 格式化评论数据
func (s *AnalysisService) FormatComments(comments []bilibili.CommentData, limit int) string {
	return s.formatComments(comments, limit)
}

// RenderTemplate 渲染模板
func (s *AnalysisService) RenderTemplate(template string, commentsText, videoTitle string, commentCount int) string {
	promptTemplate := PromptTemplate{Prompt: template}
	return s.renderTemplate(promptTemplate, commentsText, videoTitle, commentCount)
}

// AnalysisRequest 分析请求
type AnalysisRequest struct {
	TaskID       string                 `json:"task_id"`
	VideoTitle   string                 `json:"video_title"`
	Comments     []bilibili.CommentData `json:"comments"`
	Template     string                 `json:"template"`      // 可以是模板ID或自定义Prompt
	CommentLimit int                    `json:"comment_limit"` // 限制分析评论数量，0表示全部
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Analysis  string `json:"analysis"`
	TaskID    string `json:"task_id"`
	Timestamp string `json:"timestamp"`
}
