package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bilibili/internal/services"
	"github.com/gin-gonic/gin"
)

// V2Handlers 新版API处理器（为重新设计的前端页面服务）
type V2Handlers struct {
	commentService  *services.CommentService
	analysisService *services.AnalysisService
}

// NewV2Handlers 创建V2处理器
func NewV2Handlers(commentService *services.CommentService, analysisService *services.AnalysisService) *V2Handlers {
	return &V2Handlers{
		commentService:  commentService,
		analysisService: analysisService,
	}
}

// ============================================================================
// 任务相关 API
// ============================================================================

// GetTasksHandler 获取所有任务
// GET /api/v2/tasks
// Response: 200 [{task对象}, ...]
func (h *V2Handlers) GetTasksHandler(c *gin.Context) {
	tasks := h.commentService.GetAllTasks()

	// 转换为前端友好的格式
	result := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		commentCount := task.Progress.TotalComments
		if commentCount == 0 && len(task.Comments) > 0 {
			commentCount = len(task.Comments)
		}

		result = append(result, gin.H{
			"task_id":       task.TaskID,
			"video_id":      task.VideoID,
			"video_title":   task.VideoTitle,
			"status":        task.Status,
			"comment_count": commentCount,
			"start_time":    task.StartTime.Format("2006-01-02 15:04"),
			"end_time":      task.EndTime.Format("2006-01-02 15:04"),
			"error":         task.Error,
			// 进度信息
			"progress": gin.H{
				"current_page":   task.Progress.CurrentPage,
				"page_limit":     task.Progress.PageLimit,
				"total_comments": task.Progress.TotalComments,
			},
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetTaskHandler 获取单个任务详情
// GET /api/v2/tasks/:id
// Response: 200 {task对象}
func (h *V2Handlers) GetTaskHandler(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.commentService.GetTaskProgress(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 构建评论预览数据
	commentsPreview := make([]gin.H, 0)
	if len(task.Comments) > 0 {
		limit := 5
		if len(task.Comments) < limit {
			limit = len(task.Comments)
		}
		for i := 0; i < limit; i++ {
			comment := task.Comments[i]
			commentsPreview = append(commentsPreview, gin.H{
				"rpid":    comment.RPID,
				"author":  comment.Member.Uname,
				"avatar":  comment.Member.Avatar,
				"content": comment.Content.Message,
				"likes":   comment.Like,
				"time":    formatTimestamp(comment.Ctime),
				"level":   comment.Member.LevelInfo.CurrentLevel,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":       task.TaskID,
		"video_id":      task.VideoID,
		"video_title":   task.VideoTitle,
		"status":        task.Status,
		"comment_count": len(task.Comments),
		"start_time":    task.StartTime.Format("2006-01-02 15:04:05"),
		"end_time":      task.EndTime.Format("2006-01-02 15:04:05"),
		"error":         task.Error,
		"progress": gin.H{
			"current_page":   task.Progress.CurrentPage,
			"page_limit":     task.Progress.PageLimit,
			"total_comments": task.Progress.TotalComments,
		},
		"comments": commentsPreview,
	})
}

// ============================================================================
// 模板相关 API
// ============================================================================

// GetTemplatesHandler 获取所有分析模板
// GET /api/v2/templates
// Response: 200 [{template对象}, ...]
func (h *V2Handlers) GetTemplatesHandler(c *gin.Context) {
	templates := h.analysisService.GetPresetTemplates()

	result := make([]gin.H, 0, len(templates))
	for _, t := range templates {
		result = append(result, gin.H{
			"id":          t.ID,
			"name":        t.Name,
			"description": t.Description,
			"prompt":      t.Prompt,
		})
	}

	// 添加自定义模板选项
	result = append(result, gin.H{
		"id":          "custom",
		"name":        "自定义分析",
		"description": "使用你自己编写的分析Prompt",
		"prompt":      "",
	})

	c.JSON(http.StatusOK, result)
}

// ============================================================================
// AI分析相关 API
// ============================================================================

// AnalyzeStreamHandlerV2 流式分析（简化SSE格式）
// POST /api/v2/analyze-stream
// SSE格式: data: 实际文本内容\n\n (不需要event:,不需要JSON编码)
func (h *V2Handlers) AnalyzeStreamHandlerV2(c *gin.Context) {
	var req struct {
		TaskID       string `json:"task_id" binding:"required"`
		TemplateID   string `json:"template_id" binding:"required"`
		CustomPrompt string `json:"custom_prompt"`
		CommentLimit int    `json:"comment_limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 获取任务
	task, err := h.commentService.GetTaskProgress(req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	if task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务尚未完成"})
		return
	}

	if len(task.Comments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可分析的评论"})
		return
	}

	// 获取模板
	template := ""
	if req.TemplateID == "custom" {
		template = req.CustomPrompt
	} else {
		t := h.analysisService.GetTemplateByID(req.TemplateID)
		if t == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "模板不存在"})
			return
		}
		template = t.Prompt
	}

	if template == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择模板或输入自定义Prompt"})
		return
	}

	// 设置SSE响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "流式传输不支持"})
		return
	}

	// 创建通道
	streamChan := make(chan string, 100)
	errorChan := make(chan error, 1)

	// 在goroutine中执行分析
	go func() {
		commentsText := h.analysisService.FormatComments(task.Comments, req.CommentLimit)
		prompt := h.analysisService.RenderTemplate(template, commentsText, task.VideoTitle, len(task.Comments))

		_, err := h.analysisService.CallLLMStream(func(chunk string) {
			streamChan <- chunk
		}, prompt)

		if err != nil {
			errorChan <- err
			return
		}

		// 发送完成标记
		streamChan <- "__DONE__"
	}()

	// 发送SSE（简化格式：直接发送文本内容）
	c.Stream(func(w io.Writer) bool {
		select {
		case chunk := <-streamChan:
			if chunk == "__DONE__" {
				// 发送完成信号
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				return false
			}
			// 直接发送文本内容，不进行JSON编码
			// 需要转义特殊字符
			escapedChunk := strings.ReplaceAll(chunk, "\n", "\\n")
			fmt.Fprintf(w, "data: %s\n\n", escapedChunk)
			flusher.Flush()
			return true
		case err := <-errorChan:
			// 发送错误
			errorMsg := strings.ReplaceAll(err.Error(), "\n", " ")
			fmt.Fprintf(w, "data: [ERROR] %s\n\n", errorMsg)
			flusher.Flush()
			return false
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// PreviewPromptHandlerV2 预览渲染后的Prompt
// POST /api/v2/preview
func (h *V2Handlers) PreviewPromptHandlerV2(c *gin.Context) {
	var req struct {
		TaskID     string `json:"task_id" binding:"required"`
		TemplateID string `json:"template_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	task, err := h.commentService.GetTaskProgress(req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	var template services.PromptTemplate
	if req.TemplateID == "custom" {
		template = services.PromptTemplate{
			ID:     "custom",
			Name:   "自定义",
			Prompt: "",
		}
	} else {
		t := h.analysisService.GetTemplateByID(req.TemplateID)
		if t == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "模板不存在"})
			return
		}
		template = *t
	}

	// 格式化评论预览
	sampleComments := task.Comments
	if len(sampleComments) > 10 {
		sampleComments = sampleComments[:10]
	}

	commentsText := h.analysisService.FormatComments(sampleComments, 0)
	prompt := h.analysisService.RenderTemplate(template.Prompt, commentsText, task.VideoTitle, len(task.Comments))

	c.JSON(http.StatusOK, gin.H{
		"prompt": prompt,
		"count":  len(sampleComments),
	})
}

// ============================================================================
// 辅助函数
// ============================================================================

// formatTimestamp 格式化时间戳
func formatTimestamp(ts int) string {
	if ts == 0 {
		return "未知"
	}
	// 将 Unix 时间戳转换为时间
	t := time.Unix(int64(ts), 0)
	// 如果是一年内的，显示相对时间
	if time.Since(t) < 365*24*time.Hour {
		duration := time.Since(t)
		if duration < 24*time.Hour {
			hours := int(duration.Hours())
			if hours == 0 {
				return "刚刚"
			}
			return fmt.Sprintf("%d小时前", hours)
		}
		days := int(duration.Hours() / 24)
		if days < 30 {
			return fmt.Sprintf("%d天前", days)
		}
		months := days / 30
		if months < 12 {
			return fmt.Sprintf("%d个月前", months)
		}
	}
	// 否则显示完整日期
	return t.Format("2006-01-02 15:04")
}
