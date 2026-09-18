package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bilibili/internal/services"
	"bilibili/pkg/bilibili"
	"github.com/gin-gonic/gin"
)

// AnalysisHandlers 分析处理器集合
type AnalysisHandlers struct {
	commentService  *services.CommentService
	analysisService *services.AnalysisService
}

// NewAnalysisHandlers 创建分析处理器
func NewAnalysisHandlers(commentService *services.CommentService, analysisService *services.AnalysisService) *AnalysisHandlers {
	return &AnalysisHandlers{
		commentService:  commentService,
		analysisService: analysisService,
	}
}

// GetTemplatesHandler 获取预设Prompt模板
func (h *AnalysisHandlers) GetTemplatesHandler(c *gin.Context) {
	templates := h.analysisService.GetPresetTemplates()
	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
	})
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	TaskID       string `json:"task_id" binding:"required"`
	TemplateID   string `json:"template_id"`   // 模板ID
	CustomPrompt string `json:"custom_prompt"` // 自定义Prompt（template_id为custom时使用）
	CommentLimit int    `json:"comment_limit"` // 限制分析的评论数量，0表示全部
}

// AnalyzeHandler 执行评论分析
func (h *AnalysisHandlers) AnalyzeHandler(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取任务数据
	task, err := h.commentService.GetTaskWithComments(req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found: " + err.Error()})
		return
	}

	if task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task is not completed yet"})
		return
	}

	if len(task.Comments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No comments to analyze"})
		return
	}

	// 获取模板或使用自定义Prompt
	template := ""
	if req.TemplateID == "custom" && req.CustomPrompt != "" {
		template = req.CustomPrompt
	} else if req.TemplateID != "" {
		t := h.analysisService.GetTemplateByID(req.TemplateID)
		if t == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Template not found"})
			return
		}
		template = t.Prompt
	}

	// 构建分析请求
	analysisReq := &services.AnalysisRequest{
		TaskID:       req.TaskID,
		VideoTitle:   task.VideoTitle,
		Comments:     task.Comments,
		Template:     template,
		CommentLimit: req.CommentLimit,
	}

	// 执行分析
	result, err := h.analysisService.AnalyzeComments(analysisReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":   result.TaskID,
		"analysis":  result.Analysis,
		"timestamp": result.Timestamp,
	})
}

// GetCommentsForAnalysisHandler 获取任务评论列表（用于分析页面选择）
func (h *AnalysisHandlers) GetCommentsForAnalysisHandler(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := h.commentService.GetTaskWithComments(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 返回简化的评论摘要
	summary := struct {
		TaskID         string `json:"task_id"`
		VideoTitle     string `json:"video_title"`
		Status         string `json:"status"`
		TotalCount     int    `json:"total_count"`
		SampleCount    int    `json:"sample_count"` // 返回样本数量
		SampleComments []struct {
			RPID    int64  `json:"rpid"`
			Author  string `json:"author"`
			Content string `json:"content"`
			Likes   int    `json:"likes"`
		} `json:"sample_comments"`
	}{
		TaskID:      task.TaskID,
		VideoTitle:  task.VideoTitle,
		Status:      task.Status,
		TotalCount:  len(task.Comments),
		SampleCount: 5, // 返回前5条作为样本
	}

	// 取前5条评论作为样本
	limit := 5
	if len(task.Comments) < limit {
		limit = len(task.Comments)
	}
	for i := 0; i < limit; i++ {
		c := task.Comments[i]
		summary.SampleComments = append(summary.SampleComments, struct {
			RPID    int64  `json:"rpid"`
			Author  string `json:"author"`
			Content string `json:"content"`
			Likes   int    `json:"likes"`
		}{
			RPID:    c.RPID,
			Author:  c.Member.Uname,
			Content: c.Content.Message,
			Likes:   c.Like,
		})
	}

	c.JSON(http.StatusOK, summary)
}

// CompletedTasksHandler 获取所有已完成的任务列表
func (h *AnalysisHandlers) CompletedTasksHandler(c *gin.Context) {
	tasks := h.commentService.GetAllTasks()

	completed := make([]gin.H, 0)
	for _, task := range tasks {
		if task.Status == "completed" {
			// 使用 Progress.TotalComments 而不是 len(task.Comments)
			// 因为评论数据可能未加载（懒加载）
			commentCount := task.Progress.TotalComments
			if commentCount == 0 && len(task.Comments) > 0 {
				commentCount = len(task.Comments)
			}

			completed = append(completed, gin.H{
				"task_id":       task.TaskID,
				"video_id":      task.VideoID,
				"video_title":   task.VideoTitle,
				"comment_count": commentCount,
				"start_time":    task.StartTime.Format("2006-01-02 15:04:05"),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": completed,
	})
}

// PreviewPromptHandler 预览Prompt渲染结果
func (h *AnalysisHandlers) PreviewPromptHandler(c *gin.Context) {
	var req struct {
		TaskID     string `json:"task_id" binding:"required"`
		TemplateID string `json:"template_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	task, err := h.commentService.GetTaskWithComments(req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found: " + err.Error()})
		return
	}

	template := h.analysisService.GetTemplateByID(req.TemplateID)
	if template == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	// 格式化部分评论用于预览
	sampleComments := task.Comments
	if len(sampleComments) > 10 {
		sampleComments = sampleComments[:10]
	}

	commentsText := h.formatCommentsForPreview(sampleComments)
	prompt := h.renderPrompt(*template, commentsText, task.VideoTitle, len(task.Comments))

	c.JSON(http.StatusOK, gin.H{
		"prompt":        prompt,
		"comment_count": len(sampleComments),
	})
}

// formatCommentsForPreview 格式化评论用于预览
func (h *AnalysisHandlers) formatCommentsForPreview(comments []bilibili.CommentData) string {
	var builder strings.Builder
	for i, c := range comments {
		builder.WriteString(fmt.Sprintf("[%d] %s (点赞:%d)\n", i+1, c.Content.Message, c.Like))
	}
	return builder.String()
}

// renderPrompt 渲染Prompt
func (h *AnalysisHandlers) renderPrompt(template services.PromptTemplate, commentsText, videoTitle string, commentCount int) string {
	result := template.Prompt
	result = strings.ReplaceAll(result, "{{comments}}", commentsText)
	result = strings.ReplaceAll(result, "{{video_title}}", videoTitle)
	result = strings.ReplaceAll(result, "{{comment_count}}", fmt.Sprintf("%d", commentCount))
	return result
}

// AnalyzeStreamHandler 执行流式评论分析（SSE）
func (h *AnalysisHandlers) AnalyzeStreamHandler(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取任务数据
	task, err := h.commentService.GetTaskWithComments(req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found: " + err.Error()})
		return
	}

	if task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task is not completed yet"})
		return
	}

	if len(task.Comments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No comments to analyze"})
		return
	}

	// 获取模板或使用自定义Prompt
	template := ""
	if req.TemplateID == "custom" && req.CustomPrompt != "" {
		template = req.CustomPrompt
	} else if req.TemplateID != "" {
		t := h.analysisService.GetTemplateByID(req.TemplateID)
		if t == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Template not found"})
			return
		}
		template = t.Prompt
	}

	commentsText := h.analysisService.FormatComments(task.Comments, req.CommentLimit)
	promptTemplate := template
	if promptTemplate == "" {
		promptTemplate = h.analysisService.GetPresetTemplates()[0].Prompt
	}
	prompt := h.analysisService.RenderTemplate(promptTemplate, commentsText, task.VideoTitle, len(task.Comments))

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	events := runAnalysisStream(c.Request.Context(), h.analysisService, prompt)
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-events:
			if !ok {
				return false
			}
			if event.Err != nil {
				data, _ := json.Marshal(event.Err.Error())
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", data)
				flusher.Flush()
				return false
			}
			if event.Done {
				fmt.Fprint(w, "event: done\ndata: \n\n")
				flusher.Flush()
				return false
			}
			data, _ := json.Marshal(event.Chunk)
			fmt.Fprintf(w, "event: content\ndata: %s\n\n", data)
			flusher.Flush()
			return true
		case <-heartbeat.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
