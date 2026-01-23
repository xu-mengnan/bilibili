package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"sort"
	"strings"
	"sync"
	"time"

	"bilibili/pkg/bilibili"
	"bilibili/pkg/storage"
	"bilibili/pkg/utils"
)

// CommentService 评论服务，管理爬取任务
type CommentService struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	tasks   map[string]*ScrapeTask
	mu      sync.RWMutex
	storage storage.TaskStorage // 存储层
	dirty   map[string]bool     // 脏标记：记录需要持久化的任务
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
func NewCommentService(ctx context.Context, storage storage.TaskStorage) *CommentService {
	serviceCtx, cancel := context.WithCancel(ctx)

	cs := &CommentService{
		ctx:     serviceCtx,
		cancel:  cancel,
		tasks:   make(map[string]*ScrapeTask),
		storage: storage,
		dirty:   make(map[string]bool),
	}

	// 初始化存储
	storage.Initialize()

	// 启动时从存储加载任务
	cs.loadTasksFromStorage()

	// 启动持久化goroutine
	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		cs.persistWorker()
	}()

	// 启动清理goroutine
	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		cs.cleanupWorker()
	}()

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

	// 立即持久化新任务
	go cs.saveTask(task)

	// 在后台执行爬取
	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		cs.executeScrapingTask(taskID)
	}()

	return taskID, nil
}

// GetTaskProgress 获取任务进度
func (cs *CommentService) GetTaskProgress(taskID string) (*ScrapeTask, error) {
	// 先尝试从内存获取
	cs.mu.RLock()
	task, exists := cs.tasks[taskID]
	cs.mu.RUnlock()

	// 如果任务不存在，尝试从存储加载
	if !exists {
		// 从索引加载任务元数据
		index, err := cs.storage.LoadIndex()
		if err != nil {
			return nil, fmt.Errorf("failed to load index: %w", err)
		}

		var foundMeta *storage.TaskMeta
		for i := range index.Tasks {
			if index.Tasks[i].TaskID == taskID {
				foundMeta = &index.Tasks[i]
				break
			}
		}

		if foundMeta == nil {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}

		// 构建任务对象（不含评论数据，懒加载）
		task = &ScrapeTask{
			TaskID:     foundMeta.TaskID,
			VideoID:    foundMeta.VideoID,
			VideoTitle: foundMeta.VideoTitle,
			Status:     foundMeta.Status,
			Comments:   nil, // 懒加载
			Progress: TaskProgress{
				TotalComments: foundMeta.CommentCount,
				PageLimit:     2, // 默认值
			},
			StartTime: foundMeta.StartTime,
			EndTime:   foundMeta.EndTime,
			Error:     foundMeta.Error,
		}

		// 将任务添加到内存中
		cs.mu.Lock()
		cs.tasks[taskID] = task
		cs.mu.Unlock()

		return task, nil
	}

	// 对于 completed 状态的任务，检查评论数据
	if task.Status == "completed" && (task.Comments == nil || len(task.Comments) == 0) {
		// 需要加载评论数据，直接从存储加载，不持有锁
		taskData, err := cs.storage.LoadTask(taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to load task comments: %w", err)
		}

		comments := cs.convertFromStorageFormat(taskData.Comments)

		// 获取锁并更新（双重检查）
		cs.mu.Lock()
		task = cs.tasks[taskID] // 重新获取（可能已被删除或加载）
		if task != nil && (task.Comments == nil || len(task.Comments) == 0) {
			task.Comments = comments
			task.Progress.TotalComments = len(comments)
		}
		cs.mu.Unlock()

		return task, nil
	}

	return task, nil
}

// GetAllTasks 获取所有任务（按开始时间降序排序，最新的在前）
func (cs *CommentService) GetAllTasks() []*ScrapeTask {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	tasks := make([]*ScrapeTask, 0, len(cs.tasks))
	for _, task := range cs.tasks {
		tasks = append(tasks, task)
	}

	// 按开始时间降序排序（最新的在前）
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})

	return tasks
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

	// 懒加载：如果评论数据未加载，从存储加载
	if task.Comments == nil || len(task.Comments) == 0 {
		cs.mu.Lock()
		cs.loadTaskComments(task)
		cs.mu.Unlock()
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
		// 检查是否被取消
		select {
		case <-cs.ctx.Done():
			utils.LogInfo("Scraping task cancelled: " + taskID)
			cs.mu.Lock()
			task.Status = "cancelled"
			task.Error = "Task cancelled by shutdown"
			task.EndTime = time.Now()
			cs.mu.Unlock()
			return
		default:
		}

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
	task.Comments = comments // 临时保存，用于持久化
	task.Progress.TotalComments = len(comments)
	task.EndTime = time.Now()
	cs.mu.Unlock()

	// 立即持久化完成的任务
	cs.saveTask(task)

	// 持久化后释放内存（懒加载）
	cs.mu.Lock()
	task.Comments = nil
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

	// 立即持久化失败的任务
	if task, exists := cs.tasks[taskID]; exists {
		go cs.saveTask(task)
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

	for {
		select {
		case <-cs.ctx.Done():
			utils.LogInfo("cleanupWorker stopped")
			return
		case <-ticker.C:
			cs.CleanOldTasks()
		}
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
			// 同时删除存储中的任务
			cs.storage.DeleteTask(taskID)
		}
	}

	// 更新索引
	cs.updateIndex()
}

// loadTasksFromStorage 从存储加载任务
func (cs *CommentService) loadTasksFromStorage() {
	tasks, err := cs.storage.ListTasks()
	if err != nil {
		fmt.Printf("加载任务失败: %v\n", err)
		return
	}

	fmt.Printf("从存储加载 %d 个任务\n", len(tasks))

	for _, meta := range tasks {
		// 将 running 状态的任务标记为 failed（重启中断）
		if meta.Status == "running" {
			meta.Status = "failed"
			meta.Error = "任务被中断（服务器重启）"
		}

		// 只加载元数据到内存，评论数据懒加载
		task := &ScrapeTask{
			TaskID:     meta.TaskID,
			VideoID:    meta.VideoID,
			VideoTitle: meta.VideoTitle,
			Status:     meta.Status,
			Comments:   nil, // 懒加载
			Progress: TaskProgress{
				TotalComments: meta.CommentCount, // 使用索引中的评论数
				PageLimit:     2,                 // 默认值
			},
			StartTime: meta.StartTime,
			EndTime:   meta.EndTime,
			Error:     meta.Error,
		}

		cs.tasks[meta.TaskID] = task
	}

	// 更新索引（处理状态变更）
	if len(tasks) > 0 {
		cs.updateIndex()
	}
}

// persistWorker 后台持久化工作器
func (cs *CommentService) persistWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cs.ctx.Done():
			utils.LogInfo("persistWorker stopped")
			return
		case <-ticker.C:
			cs.persistDirtyTasks()
		}
	}
}

// persistDirtyTasks 持久化脏任务
func (cs *CommentService) persistDirtyTasks() {
	cs.mu.Lock()
	dirtyTasks := make(map[string]*ScrapeTask)
	for taskID := range cs.dirty {
		if task, exists := cs.tasks[taskID]; exists {
			dirtyTasks[taskID] = task
		}
		delete(cs.dirty, taskID)
	}
	cs.mu.Unlock()

	// 持久化脏任务
	for _, task := range dirtyTasks {
		if err := cs.saveTask(task); err != nil {
			fmt.Printf("持久化任务失败 %s: %v\n", task.TaskID, err)
		}
	}
}

// saveTask 保存单个任务
func (cs *CommentService) saveTask(task *ScrapeTask) error {
	// 转换为存储层格式
	taskData := cs.convertToStorageFormat(task)

	// 保存任务数据
	if err := cs.storage.SaveTask(taskData); err != nil {
		return err
	}

	// 更新索引
	return cs.updateIndex()
}

// updateIndex 更新任务索引
func (cs *CommentService) updateIndex() error {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// 构建索引
	var metas []storage.TaskMeta
	for _, task := range cs.tasks {
		// 使用 Progress.TotalComments 而不是 len(task.Comments)
		// 因为 Comments 可能未加载（懒加载）
		commentCount := task.Progress.TotalComments
		if commentCount == 0 && len(task.Comments) > 0 {
			commentCount = len(task.Comments)
		}

		meta := storage.TaskMeta{
			TaskID:       task.TaskID,
			VideoID:      task.VideoID,
			VideoTitle:   task.VideoTitle,
			Status:       task.Status,
			CommentCount: commentCount,
			StartTime:    task.StartTime,
			EndTime:      task.EndTime,
			DataFile:     task.TaskID + ".json",
			Error:        task.Error,
		}
		metas = append(metas, meta)
	}

	index := &storage.TaskIndex{
		Tasks: metas,
	}

	return cs.storage.SaveIndex(index)
}

// loadTaskComments 懒加载任务的评论数据
func (cs *CommentService) loadTaskComments(task *ScrapeTask) error {
	taskData, err := cs.storage.LoadTask(task.TaskID)
	if err != nil {
		return err
	}

	// 转换评论数据
	task.Comments = cs.convertFromStorageFormat(taskData.Comments)
	// 更新进度中的评论总数
	task.Progress.TotalComments = len(task.Comments)
	return nil
}

// convertToStorageFormat 转换为存储层格式
func (cs *CommentService) convertToStorageFormat(task *ScrapeTask) *storage.TaskData {
	comments := make([]storage.CommentEntry, len(task.Comments))
	for i, c := range task.Comments {
		comments[i] = cs.convertCommentToStorage(c)
	}

	return &storage.TaskData{
		TaskID:     task.TaskID,
		VideoID:    task.VideoID,
		VideoTitle: task.VideoTitle,
		Status:     task.Status,
		Comments:   comments,
		Progress: storage.TaskProgressEntry{
			CurrentPage:   task.Progress.CurrentPage,
			TotalComments: task.Progress.TotalComments,
			PageLimit:     task.Progress.PageLimit,
		},
		StartTime:      task.StartTime,
		EndTime:        task.EndTime,
		Error:          task.Error,
		AuthType:       task.AuthType,
		Cookie:         task.Cookie,
		AppKey:         task.AppKey,
		AppSecret:      task.AppSecret,
		PageLimit:      task.PageLimit,
		DelayMs:        task.DelayMs,
		SortMode:       task.SortMode,
		IncludeReplies: task.IncludeReplies,
	}
}

// convertFromStorageFormat 从存储层格式转换
func (cs *CommentService) convertFromStorageFormat(entries []storage.CommentEntry) []bilibili.CommentData {
	comments := make([]bilibili.CommentData, len(entries))
	for i, e := range entries {
		comments[i] = cs.convertCommentFromStorage(e)
	}
	return comments
}

// convertCommentToStorage 转换单条评论到存储格式
func (cs *CommentService) convertCommentToStorage(c bilibili.CommentData) storage.CommentEntry {
	replies := make([]storage.CommentEntry, len(c.Replies))
	for i, r := range c.Replies {
		replies[i] = cs.convertCommentToStorage(r)
	}

	return storage.CommentEntry{
		RPID:      c.RPID,
		OID:       c.OID,
		Type:      c.Type,
		Mid:       c.Mid,
		Root:      c.Root,
		Parent:    c.Parent,
		Dialog:    c.Dialog,
		Count:     c.Count,
		RCount:    c.RCount,
		State:     c.State,
		FansGrade: c.FansGrade,
		Attr:      c.Attr,
		Ctime:     c.Ctime,
		Like:      c.Like,
		Content: storage.CommentContent{
			Message: c.Content.Message,
		},
		Member: storage.CommentMember{
			Mid:    c.Member.Mid,
			Name:   c.Member.Uname,
			Sex:    c.Member.Sex,
			Avatar: c.Member.Avatar,
			Sign:   c.Member.Sign,
			Rank:   c.Member.Rank,
			Level:  c.Member.LevelInfo.CurrentLevel,
		},
		Replies: replies,
	}
}

// convertCommentFromStorage 从存储格式转换单条评论
func (cs *CommentService) convertCommentFromStorage(e storage.CommentEntry) bilibili.CommentData {
	replies := make([]bilibili.CommentData, len(e.Replies))
	for i, r := range e.Replies {
		replies[i] = cs.convertCommentFromStorage(r)
	}

	return bilibili.CommentData{
		RPID:      e.RPID,
		OID:       e.OID,
		Type:      e.Type,
		Mid:       e.Mid,
		Root:      e.Root,
		Parent:    e.Parent,
		Dialog:    e.Dialog,
		Count:     e.Count,
		RCount:    e.RCount,
		State:     e.State,
		FansGrade: e.FansGrade,
		Attr:      e.Attr,
		Ctime:     e.Ctime,
		Like:      e.Like,
		Content: bilibili.CommentContent{
			Message: e.Content.Message,
		},
		Member: bilibili.CommentMember{
			Mid:    e.Member.Mid,
			Uname:  e.Member.Name,
			Sex:    e.Member.Sex,
			Avatar: e.Member.Avatar,
			Sign:   e.Member.Sign,
			Rank:   e.Member.Rank,
		},
		Replies: replies,
	}
}

// markDirty 标记任务为脏数据
func (cs *CommentService) markDirty(taskID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.dirty[taskID] = true
}

// Shutdown 优雅关闭服务
func (cs *CommentService) Shutdown(ctx context.Context) error {
	utils.LogInfo("Shutting down CommentService...")

	// 取消 context
	cs.cancel()

	// 等待所有 goroutine 结束
	done := make(chan struct{})
	go func() {
		cs.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		utils.LogInfo("CommentService shutdown complete")
		return nil
	case <-ctx.Done():
		utils.LogError("CommentService shutdown timeout")
		return ctx.Err()
	}
}
