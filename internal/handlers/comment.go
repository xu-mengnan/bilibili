package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"bilibili/internal/services"
	"bilibili/pkg/bilibili"
	"github.com/gin-gonic/gin"
)

// CommentHandlers 评论处理器集合
type CommentHandlers struct {
	commentService *services.CommentService
	exportService  *services.ExportService
}

// NewCommentHandlers 创建评论处理器
func NewCommentHandlers(commentService *services.CommentService, exportService *services.ExportService) *CommentHandlers {
	return &CommentHandlers{
		commentService: commentService,
		exportService:  exportService,
	}
}

// ScrapeRequest 爬取请求
type ScrapeRequest struct {
	VideoID        string `json:"video_id" binding:"required"`
	AuthType       string `json:"auth_type"` // none, cookie, app
	Cookie         string `json:"cookie"`
	AppKey         string `json:"app_key"`
	AppSecret      string `json:"app_secret"`
	PageLimit      int    `json:"page_limit"`
	DelayMs        int    `json:"delay_ms"`
	SortMode       string `json:"sort_mode"`       // time(按时间), hot(按热度)
	IncludeReplies bool   `json:"include_replies"` // 是否抓取子评论
}

// ScrapeResponse 爬取响应
type ScrapeResponse struct {
	TaskID   string                `json:"task_id"`
	VideoID  string                `json:"video_id"`
	Status   string                `json:"status"`
	Progress services.TaskProgress `json:"progress"`
}

// ScrapeCommentsHandler 启动爬取任务
func (h *CommentHandlers) ScrapeCommentsHandler(c *gin.Context) {
	var req ScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 设置默认值
	if req.PageLimit == 0 {
		req.PageLimit = 2
	}
	if req.DelayMs == 0 {
		req.DelayMs = 300
	}
	if req.AuthType == "" {
		req.AuthType = "none"
	}
	if req.SortMode == "" {
		req.SortMode = "time" // 默认按时间排序
	}

	// 验证排序模式
	if req.SortMode != "time" && req.SortMode != "hot" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort_mode: must be 'time' or 'hot'"})
		return
	}

	// 启动爬取任务
	taskID, err := h.commentService.StartScrapeTask(
		req.VideoID,
		req.AuthType,
		req.Cookie,
		req.AppKey,
		req.AppSecret,
		req.SortMode,
		req.IncludeReplies,
		req.PageLimit,
		req.DelayMs,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start scraping: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, ScrapeResponse{
		TaskID:  taskID,
		VideoID: req.VideoID,
		Status:  "running",
		Progress: services.TaskProgress{
			CurrentPage:   0,
			TotalComments: 0,
			PageLimit:     req.PageLimit,
		},
	})
}

// ProgressResponse 进度响应
type ProgressResponse struct {
	TaskID         string                 `json:"task_id"`
	Status         string                 `json:"status"`
	Progress       services.TaskProgress  `json:"progress"`
	VideoTitle     string                 `json:"video_title,omitempty"`
	StartTime      string                 `json:"start_time"`
	EndTime        string                 `json:"end_time,omitempty"`
	ElapsedSeconds int64                  `json:"elapsed_seconds"`
	Error          string                 `json:"error,omitempty"`
	VideoID        string                 `json:"video_id,omitempty"`
	Comments       []bilibili.CommentData `json:"comments,omitempty"` // 添加评论数据
}

// GetProgressHandler 获取任务进度
func (h *CommentHandlers) GetProgressHandler(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := h.commentService.GetTaskProgress(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	elapsed := time.Since(task.StartTime).Seconds()

	// 格式化结束时间
	endTime := ""
	if !task.EndTime.IsZero() {
		endTime = task.EndTime.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, ProgressResponse{
		TaskID:         task.TaskID,
		Status:         task.Status,
		Progress:       task.Progress,
		VideoTitle:     task.VideoTitle,
		StartTime:      task.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:        endTime,
		ElapsedSeconds: int64(elapsed),
		Error:          task.Error,
		VideoID:        task.VideoID,
		Comments:       task.Comments, // 包含评论数据
	})
}

// CommentItem 评论项（简化版）
type CommentItem struct {
	RPID    int64         `json:"rpid"`
	Author  string        `json:"author"`
	Avatar  string        `json:"avatar"`
	Content string        `json:"content"`
	Likes   int           `json:"likes"`
	Time    string        `json:"time"`
	Level   int           `json:"level"`
	Replies []CommentItem `json:"replies,omitempty"` // 子评论
}

// ResultResponse 结果响应
type ResultResponse struct {
	TaskID     string        `json:"task_id"`
	TotalCount int           `json:"total_count"`
	Comments   []CommentItem `json:"comments"`
}

// convertToCommentItem 转换为CommentItem（包含子评论）
func convertToCommentItem(comment bilibili.CommentData) CommentItem {
	// 处理头像URL：B站返回的是 protocol-relative URL (//开头)
	avatar := comment.Member.Avatar
	if strings.HasPrefix(avatar, "//") {
		avatar = "https:" + avatar
	}

	item := CommentItem{
		RPID:    comment.RPID,
		Author:  comment.Member.Uname,
		Avatar:  avatar,
		Content: comment.Content.Message,
		Likes:   comment.Like,
		Time:    time.Unix(int64(comment.Ctime), 0).Format("2006-01-02 15:04:05"),
		Level:   comment.Member.LevelInfo.CurrentLevel,
	}

	// 递归处理子评论
	if len(comment.Replies) > 0 {
		item.Replies = make([]CommentItem, 0, len(comment.Replies))
		for _, reply := range comment.Replies {
			item.Replies = append(item.Replies, convertToCommentItem(reply))
		}
	}

	return item
}

// GetResultHandler 获取爬取结果
func (h *CommentHandlers) GetResultHandler(c *gin.Context) {
	taskID := c.Param("task_id")
	sortBy := c.DefaultQuery("sort", "time_desc")
	keyword := c.Query("keyword")
	limitStr := c.DefaultQuery("limit", "1000")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 1000
	}

	comments, totalCount, err := h.commentService.GetTaskResult(taskID, sortBy, keyword, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为简化格式（包含子评论）
	items := make([]CommentItem, 0, len(comments))
	for _, comment := range comments {
		items = append(items, convertToCommentItem(comment))
	}

	c.JSON(http.StatusOK, ResultResponse{
		TaskID:     taskID,
		TotalCount: totalCount,
		Comments:   items,
	})
}

// ExportRequest 导出请求
type ExportRequest struct {
	TaskID   string `json:"task_id" binding:"required"`
	Format   string `json:"format" binding:"required"`
	SortBy   string `json:"sort"`
	Filename string `json:"filename"`
}

// ExportResponse 导出响应
type ExportResponse struct {
	FileID      string `json:"file_id"`
	Filename    string `json:"filename"`
	DownloadURL string `json:"download_url"`
	CreatedAt   string `json:"created_at"`
}

// ExportCommentsHandler 导出评论
func (h *CommentHandlers) ExportCommentsHandler(c *gin.Context) {
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取评论数据
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "time_desc"
	}

	comments, _, err := h.commentService.GetTaskResult(req.TaskID, sortBy, "", 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 导出
	exportFile, err := h.exportService.ExportComments(comments, req.Format, req.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, ExportResponse{
		FileID:      exportFile.FileID,
		Filename:    exportFile.Filename,
		DownloadURL: "/api/download/" + exportFile.FileID,
		CreatedAt:   exportFile.CreatedAt.Format(time.RFC3339),
	})
}

// DownloadFileHandler 下载导出文件
func (h *CommentHandlers) DownloadFileHandler(c *gin.Context) {
	fileID := c.Param("file_id")

	exportFile, err := h.exportService.GetExportFile(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.FileAttachment(exportFile.FilePath, exportFile.Filename)
}

// GetCommentsStatsHandler 获取评论统计
func (h *CommentHandlers) GetCommentsStatsHandler(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := h.commentService.GetTaskProgress(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task not completed yet"})
		return
	}

	// 统计数据
	stats := map[string]interface{}{
		"task_id":        taskID,
		"total_comments": len(task.Comments),
		"by_date":        make(map[string]int),
		"by_likes": map[string]int{
			"0-10":   0,
			"11-50":  0,
			"51-100": 0,
			"100+":   0,
		},
		"top_keywords": []map[string]interface{}{},
	}

	// 统计日期分布
	dateMap := make(map[string]int)
	//wordCount := make(map[string]int)

	for _, comment := range task.Comments {
		// 日期统计
		date := time.Unix(int64(comment.Ctime), 0).Format("2006-01-02")
		dateMap[date]++

		// 点赞数分布
		likesMap := stats["by_likes"].(map[string]int)
		if comment.Like <= 10 {
			likesMap["0-10"]++
		} else if comment.Like <= 50 {
			likesMap["11-50"]++
		} else if comment.Like <= 100 {
			likesMap["51-100"]++
		} else {
			likesMap["100+"]++
		}

		// 简单的关键词统计（按空格分词）
		// 注：这里可以使用更复杂的中文分词库
		// words := strings.Fields(comment.Content.Message)
		// for _, word := range words {
		// 	if len(word) > 1 {
		// 		wordCount[word]++
		// 	}
		// }
	}

	stats["by_date"] = dateMap

	c.JSON(http.StatusOK, stats)
}

// GetAllTasksHandler 获取所有任务列表
func (h *CommentHandlers) GetAllTasksHandler(c *gin.Context) {
	tasks := h.commentService.GetAllTasks()

	result := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		// 使用 Progress.TotalComments 作为评论数
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
			"start_time":    task.StartTime.Format("2006-01-02 15:04:05"),
			"end_time":      task.EndTime.Format("2006-01-02 15:04:05"),
			"error":         task.Error,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tasks": result})
}
