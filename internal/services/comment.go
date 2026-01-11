package services

import (
	"fmt"
	"github.com/google/uuid"
	"sort"
	"strings"
	"sync"
	"time"

	"bilibili/pkg/bilibili"
)

// CommentService 评论服务，管理爬取任务
type CommentService struct {
	tasks map[string]*ScrapeTask
	mu    sync.RWMutex
}

// ScrapeTask 爬取任务
type ScrapeTask struct {
	TaskID         string
	VideoID        string
	VideoTitle     string
	Status         string // running, completed, failed
	Comments       []bilibili.CommentData
	Progress       TaskProgress
	StartTime      time.Time
	EndTime        time.Time
	Error          string
	AuthType       string
	Cookie         string
	AppKey         string
	AppSecret      string
	PageLimit      int
	DelayMs        int
	SortMode       string // "time" 按时间, "hot" 按热度
	IncludeReplies bool   // 是否包含子评论
}

// TaskProgress 任务进度
type TaskProgress struct {
	CurrentPage   int `json:"current_page"`
	TotalComments int `json:"total_comments"`
	PageLimit     int `json:"page_limit"`
}

// NewCommentService 创建评论服务
func NewCommentService() *CommentService {
	cs := &CommentService{
		tasks: make(map[string]*ScrapeTask),
	}
	// 启动清理goroutine
	go cs.cleanupWorker()
	return cs
}

// StartScrapeTask 启动爬取任务
func (cs *CommentService) StartScrapeTask(videoID, authType, cookie, appKey, appSecret, sortMode string, includeReplies bool, pageLimit, delayMs int) (string, error) {
	taskID := uuid.New().String()

	// 设置默认排序模式
	if sortMode == "" {
		sortMode = "time"
	}

	task := &ScrapeTask{
		TaskID:         taskID,
		VideoID:        videoID,
		Status:         "running",
		Comments:       []bilibili.CommentData{},
		Progress:       TaskProgress{CurrentPage: 0, TotalComments: 0, PageLimit: pageLimit},
		StartTime:      time.Now(),
		AuthType:       authType,
		Cookie:         cookie,
		AppKey:         appKey,
		AppSecret:      appSecret,
		PageLimit:      pageLimit,
		DelayMs:        delayMs,
		SortMode:       sortMode,
		IncludeReplies: includeReplies,
	}

	cs.mu.Lock()
	cs.tasks[taskID] = task
	cs.mu.Unlock()

	// 在后台执行爬取
	go cs.executeScrapingTask(taskID)

	return taskID, nil
}

// GetTaskProgress 获取任务进度
func (cs *CommentService) GetTaskProgress(taskID string) (*ScrapeTask, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	task, exists := cs.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// GetTaskResult 获取任务结果（带筛选排序）
func (cs *CommentService) GetTaskResult(taskID, sortBy, keyword string, limit int) ([]bilibili.CommentData, int, error) {
	cs.mu.RLock()
	task, exists := cs.tasks[taskID]
	cs.mu.RUnlock()

	if !exists {
		return nil, 0, fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status != "completed" {
		return nil, 0, fmt.Errorf("task not completed yet")
	}

	// 复制评论数据，避免修改原始数据
	comments := make([]bilibili.CommentData, len(task.Comments))
	copy(comments, task.Comments)

	// 筛选
	if keyword != "" {
		comments = cs.FilterComments(comments, keyword)
	}

	totalCount := len(comments)

	// 排序
	if sortBy != "" {
		cs.SortComments(comments, sortBy)
	}

	// 限制数量
	if limit > 0 && limit < len(comments) {
		comments = comments[:limit]
	}

	return comments, totalCount, nil
}

// executeScrapingTask 执行爬取任务（后台goroutine）
func (cs *CommentService) executeScrapingTask(taskID string) {
	cs.mu.RLock()
	task := cs.tasks[taskID]
	cs.mu.RUnlock()

	if task == nil {
		return
	}

	// 首先获取视频信息
	videoResp, err := bilibili.GetVideoByBVID(task.VideoID)
	if err != nil {
		cs.updateTaskError(taskID, fmt.Sprintf("failed to get video info: %v", err))
		return
	}

	if videoResp.Code != 0 {
		cs.updateTaskError(taskID, fmt.Sprintf("video API error: %s", videoResp.Message))
		return
	}

	// 更新视频标题
	cs.mu.Lock()
	task.VideoTitle = videoResp.Data.Title
	cs.mu.Unlock()

	// 准备认证选项
	var opts []bilibili.CommentOption
	switch task.AuthType {
	case "cookie":
		if task.Cookie != "" {
			opts = append(opts, bilibili.WithCookie(task.Cookie))
		}
	case "app":
		if task.AppKey != "" && task.AppSecret != "" {
			opts = append(opts, bilibili.WithAppAuth(task.AppKey, task.AppSecret))
		}
	}

	// 添加排序模式选项
	if task.SortMode != "" {
		opts = append(opts, bilibili.WithSortMode(task.SortMode))
	}

	// 爬取评论
	oid := videoResp.Data.AID
	pageSize := 20
	nextCursor := 0
	nextOffset := ""
	commentMap := make(map[int64]bilibili.CommentData) // 用于去重

	for page := 1; page <= task.PageLimit; page++ {
		// 获取评论
		var commentsResp *bilibili.CommentResponse
		var err error

		if nextOffset != "" {
			commentsResp, err = bilibili.GetCommentsWithOffset(oid, page, pageSize, nextCursor, nextOffset, opts...)
		} else {
			commentsResp, err = bilibili.GetComments(oid, page, pageSize, nextCursor, opts...)
		}

		if err != nil {
			cs.updateTaskError(taskID, fmt.Sprintf("failed to get comments on page %d: %v", page, err))
			return
		}

		if commentsResp.Code != 0 {
			cs.updateTaskError(taskID, fmt.Sprintf("comment API error on page %d: %s", page, commentsResp.Message))
			return
		}

		// 添加评论（去重）
		if commentsResp.Data.Replies != nil {
			for _, comment := range commentsResp.Data.Replies {
				// 如果需要获取子评论
				if task.IncludeReplies && comment.RCount > 0 {
					// 添加延迟避免请求过快
					time.Sleep(200 * time.Millisecond)

					// 获取前3条子评论
					subComments, err := bilibili.GetSubComments(oid, comment.RPID, opts...)
					if err == nil && len(subComments) > 0 {
						// 只取前3条
						if len(subComments) > 3 {
							subComments = subComments[:3]
						}
						comment.Replies = subComments
					}
				}
				commentMap[comment.RPID] = comment
			}
		}

		// 更新进度
		cs.mu.Lock()
		task.Progress.CurrentPage = page
		task.Progress.TotalComments = len(commentMap)
		cs.mu.Unlock()

		// 检查是否有更多评论
		if commentsResp.Data.Cursor.Next == 0 && commentsResp.Data.Cursor.PaginationReply.NextOffset == "" {
			break
		}

		// 更新游标
		nextCursor = commentsResp.Data.Cursor.Next
		nextOffset = commentsResp.Data.Cursor.PaginationReply.NextOffset

		// 延迟下次请求
		if task.DelayMs > 0 && page < task.PageLimit {
			time.Sleep(time.Duration(task.DelayMs) * time.Millisecond)
		}
	}

	// 将map转为slice
	comments := make([]bilibili.CommentData, 0, len(commentMap))
	for _, comment := range commentMap {
		comments = append(comments, comment)
	}

	// 标记任务完成
	cs.mu.Lock()
	task.Status = "completed"
	task.Comments = comments
	task.Progress.TotalComments = len(comments)
	task.EndTime = time.Now()
	cs.mu.Unlock()
}

// updateTaskError 更新任务错误状态
func (cs *CommentService) updateTaskError(taskID, errMsg string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if task, exists := cs.tasks[taskID]; exists {
		task.Status = "failed"
		task.Error = errMsg
		task.EndTime = time.Now()
	}
}

// SortComments 排序评论
func (cs *CommentService) SortComments(comments []bilibili.CommentData, sortBy string) {
	switch sortBy {
	case "time_desc":
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].Ctime > comments[j].Ctime
		})
	case "time_asc":
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].Ctime < comments[j].Ctime
		})
	case "like_desc":
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].Like > comments[j].Like
		})
	case "like_asc":
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].Like < comments[j].Like
		})
	}
}

// FilterComments 筛选评论（关键词搜索）
func (cs *CommentService) FilterComments(comments []bilibili.CommentData, keyword string) []bilibili.CommentData {
	if keyword == "" {
		return comments
	}

	keyword = strings.ToLower(keyword)
	filtered := []bilibili.CommentData{}

	for _, comment := range comments {
		// 搜索评论内容
		if strings.Contains(strings.ToLower(comment.Content.Message), keyword) {
			filtered = append(filtered, comment)
			continue
		}
		// 搜索用户名
		if strings.Contains(strings.ToLower(comment.Member.Uname), keyword) {
			filtered = append(filtered, comment)
			continue
		}
	}

	return filtered
}

// cleanupWorker 定期清理旧任务（1小时前）
func (cs *CommentService) cleanupWorker() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cs.CleanOldTasks()
	}
}

// CleanOldTasks 清理旧任务
func (cs *CommentService) CleanOldTasks() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for taskID, task := range cs.tasks {
		if task.EndTime.Before(cutoff) && !task.EndTime.IsZero() {
			delete(cs.tasks, taskID)
		}
	}
}
