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
	CommentsLoaded bool   // 评论是否已加载到内存，仅运行时使用
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
	}

	if err := storage.Initialize(); err != nil {
		utils.LogError("初始化任务存储失败: " + err.Error())
	}

	cs.loadTasksFromStorage()

	// 仅保留清理 goroutine；持久化在状态转换点同步完成，避免未跟踪 save goroutine。
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
	if sortMode == "" {
		sortMode = "time"
	}

	task := &ScrapeTask{
		TaskID:         taskID,
		VideoID:        videoID,
		Status:         "running",
		Comments:       []bilibili.CommentData{},
		CommentsLoaded: true,
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

	// 先持久化已接受的任务，再启动后台工作，避免“已返回 task_id 但任务尚未落盘”。
	if err := cs.persistTaskByID(taskID); err != nil {
		cs.mu.Lock()
		delete(cs.tasks, taskID)
		cs.mu.Unlock()
		return "", fmt.Errorf("failed to persist new task: %w", err)
	}

	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		cs.executeScrapingTask(taskID)
	}()

	return taskID, nil
}

// GetTaskProgress 返回只读元数据快照，不隐式加载大块评论数据。
func (cs *CommentService) GetTaskProgress(taskID string) (*ScrapeTask, error) {
	if err := cs.ensureTaskLoaded(taskID); err != nil {
		return nil, err
	}

	cs.mu.RLock()
	defer cs.mu.RUnlock()

	task := cs.tasks[taskID]
	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	return cloneTaskLocked(task, false), nil
}

// GetTaskWithComments 返回包含评论的只读快照。
// 评论缺失时会在不持有 service 锁的情况下从存储懒加载。
func (cs *CommentService) GetTaskWithComments(taskID string) (*ScrapeTask, error) {
	if err := cs.ensureTaskLoaded(taskID); err != nil {
		return nil, err
	}

	cs.mu.RLock()
	task := cs.tasks[taskID]
	if task == nil {
		cs.mu.RUnlock()
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if task.CommentsLoaded || task.Status != "completed" {
		snapshot := cloneTaskLocked(task, true)
		cs.mu.RUnlock()
		return snapshot, nil
	}
	cs.mu.RUnlock()

	taskData, err := cs.storage.LoadTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to load task comments: %w", err)
	}
	comments := cs.convertFromStorageFormat(taskData.Comments)

	cs.mu.Lock()
	task = cs.tasks[taskID]
	if task == nil {
		cs.mu.Unlock()
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if !task.CommentsLoaded {
		task.Comments = comments
		task.CommentsLoaded = true
		if task.Progress.TotalComments == 0 {
			task.Progress.TotalComments = len(comments)
		}
	}
	snapshot := cloneTaskLocked(task, true)
	cs.mu.Unlock()
	return snapshot, nil
}

// GetAllTasks 返回按开始时间降序排列的元数据快照。
func (cs *CommentService) GetAllTasks() []*ScrapeTask {
	cs.mu.RLock()
	tasks := make([]*ScrapeTask, 0, len(cs.tasks))
	for _, task := range cs.tasks {
		tasks = append(tasks, cloneTaskLocked(task, false))
	}
	cs.mu.RUnlock()

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})
	return tasks
}

// GetTaskResult 获取任务结果（带筛选排序）。
func (cs *CommentService) GetTaskResult(taskID, sortBy, keyword string, limit int) ([]bilibili.CommentData, int, error) {
	task, err := cs.GetTaskWithComments(taskID)
	if err != nil {
		return nil, 0, err
	}
	if task.Status != "completed" {
		return nil, 0, fmt.Errorf("task not completed yet")
	}

	comments := cloneComments(task.Comments)
	if keyword != "" {
		comments = cs.FilterComments(comments, keyword)
	}
	totalCount := len(comments)
	if sortBy != "" {
		cs.SortComments(comments, sortBy)
	}
	if limit > 0 && limit < len(comments) {
		comments = comments[:limit]
	}
	return comments, totalCount, nil
}

// executeScrapingTask 执行爬取任务（后台goroutine）
func (cs *CommentService) executeScrapingTask(taskID string) {
	cs.mu.RLock()
	taskConfig := cloneTaskLocked(cs.tasks[taskID], false)
	cs.mu.RUnlock()
	if taskConfig == nil {
		return
	}

	videoResp, err := bilibili.GetVideoByBVIDContext(cs.ctx, taskConfig.VideoID)
	if err != nil {
		if cs.ctx.Err() != nil {
			cs.cancelTask(taskID, "Task cancelled by shutdown")
		} else {
			cs.updateTaskError(taskID, fmt.Sprintf("failed to get video info: %v", err))
		}
		return
	}

	cs.mu.Lock()
	if task := cs.tasks[taskID]; task != nil {
		task.VideoTitle = videoResp.Data.Title
	}
	cs.mu.Unlock()

	var opts []bilibili.CommentOption
	switch taskConfig.AuthType {
	case "cookie":
		if taskConfig.Cookie != "" {
			opts = append(opts, bilibili.WithCookie(taskConfig.Cookie))
		}
	case "app":
		if taskConfig.AppKey != "" && taskConfig.AppSecret != "" {
			opts = append(opts, bilibili.WithAppAuth(taskConfig.AppKey, taskConfig.AppSecret))
		}
	}
	if taskConfig.SortMode != "" {
		opts = append(opts, bilibili.WithSortMode(taskConfig.SortMode))
	}

	oid := videoResp.Data.AID
	paginator := bilibili.NewCommentPaginator(oid, 20, taskConfig.PageLimit, opts...)
	commentMap := make(map[int64]bilibili.CommentData)

	for !paginator.Done() {
		commentsResp, err := paginator.Next(cs.ctx)
		if err != nil {
			if cs.ctx.Err() != nil {
				cs.cancelTask(taskID, "Task cancelled by shutdown")
			} else {
				cs.updateTaskError(taskID, fmt.Sprintf("failed to get comments on page %d: %v", paginator.Page()+1, err))
			}
			return
		}
		page := paginator.Page()

		for _, comment := range commentsResp.Data.Replies {
			if taskConfig.IncludeReplies && comment.RCount > 0 {
				timer := time.NewTimer(200 * time.Millisecond)
				select {
				case <-timer.C:
				case <-cs.ctx.Done():
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					cs.cancelTask(taskID, "Task cancelled by shutdown")
					return
				}

				subComments, err := bilibili.GetSubCommentsContext(cs.ctx, oid, comment.RPID, opts...)
				if err == nil && len(subComments) > 0 {
					if len(subComments) > 3 {
						subComments = subComments[:3]
					}
					comment.Replies = subComments
				} else if cs.ctx.Err() != nil {
					cs.cancelTask(taskID, "Task cancelled by shutdown")
					return
				}
			}
			commentMap[comment.RPID] = comment
		}

		cs.mu.Lock()
		if task := cs.tasks[taskID]; task != nil {
			task.Progress.CurrentPage = page
			task.Progress.TotalComments = len(commentMap)
		}
		cs.mu.Unlock()

		if paginator.Done() {
			break
		}
		if taskConfig.DelayMs > 0 {
			timer := time.NewTimer(time.Duration(taskConfig.DelayMs) * time.Millisecond)
			select {
			case <-timer.C:
			case <-cs.ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				cs.cancelTask(taskID, "Task cancelled by shutdown")
				return
			}
		}
	}

	comments := make([]bilibili.CommentData, 0, len(commentMap))
	for _, comment := range commentMap {
		comments = append(comments, comment)
	}

	cs.mu.Lock()
	if task := cs.tasks[taskID]; task != nil {
		task.Status = "completed"
		task.Comments = comments
		task.CommentsLoaded = true
		task.Progress.TotalComments = len(comments)
		task.EndTime = time.Now()
		task.Error = ""
	}
	cs.mu.Unlock()

	if err := cs.persistTaskByID(taskID); err != nil {
		utils.LogError("持久化完成任务失败: " + err.Error())
		cs.mu.Lock()
		if task := cs.tasks[taskID]; task != nil {
			task.Error = "persistence failed: " + err.Error()
		}
		cs.mu.Unlock()
		return
	}

	cs.mu.Lock()
	if task := cs.tasks[taskID]; task != nil && task.Status == "completed" {
		task.Comments = nil
		task.CommentsLoaded = false
	}
	cs.mu.Unlock()
}

func (cs *CommentService) cancelTask(taskID, message string) {
	cs.mu.Lock()
	task := cs.tasks[taskID]
	if task != nil {
		task.Status = "cancelled"
		task.Error = message
		task.EndTime = time.Now()
	}
	cs.mu.Unlock()

	if task != nil {
		if err := cs.persistTaskByID(taskID); err != nil {
			utils.LogError("持久化已取消任务失败: " + err.Error())
		}
	}
}

// updateTaskError 更新任务错误状态
func (cs *CommentService) updateTaskError(taskID, errMsg string) {
	cs.mu.Lock()
	task := cs.tasks[taskID]
	if task != nil {
		task.Status = "failed"
		task.Error = errMsg
		task.EndTime = time.Now()
	}
	cs.mu.Unlock()

	if task != nil {
		if err := cs.persistTaskByID(taskID); err != nil {
			utils.LogError("持久化失败任务失败: " + err.Error())
		}
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

// CleanOldTasks 清理旧任务。service 锁内只更新内存，磁盘 I/O 全部在锁外完成。
func (cs *CommentService) CleanOldTasks() {
	cutoff := time.Now().Add(-1 * time.Hour)
	removed := make(map[string]*ScrapeTask)

	cs.mu.Lock()
	for taskID, task := range cs.tasks {
		if !task.EndTime.IsZero() && task.EndTime.Before(cutoff) {
			removed[taskID] = task
			delete(cs.tasks, taskID)
		}
	}
	cs.mu.Unlock()

	if len(removed) == 0 {
		return
	}

	if err := cs.persistIndex(); err != nil {
		// 索引更新失败时恢复内存，避免 index 指向被删除的数据。
		cs.mu.Lock()
		for taskID, task := range removed {
			if _, exists := cs.tasks[taskID]; !exists {
				cs.tasks[taskID] = task
			}
		}
		cs.mu.Unlock()
		utils.LogError("更新任务索引失败，取消清理: " + err.Error())
		return
	}

	for taskID := range removed {
		if err := cs.storage.DeleteTask(taskID); err != nil {
			utils.LogError(fmt.Sprintf("删除任务 %s 失败: %v", taskID, err))
		}
	}
}

// loadTasksFromStorage 从存储加载任务元数据，评论按需加载。
func (cs *CommentService) loadTasksFromStorage() {
	metas, err := cs.storage.ListTasks()
	if err != nil {
		utils.LogError("加载任务失败: " + err.Error())
		return
	}

	utils.LogInfo(fmt.Sprintf("从存储加载 %d 个任务", len(metas)))
	statusChanged := false

	cs.mu.Lock()
	for _, meta := range metas {
		if meta.Status == "running" {
			meta.Status = "failed"
			meta.Error = "任务被中断（服务器重启）"
			statusChanged = true
		}
		cs.tasks[meta.TaskID] = taskFromMeta(meta)
	}
	cs.mu.Unlock()

	if statusChanged {
		if err := cs.persistIndex(); err != nil {
			utils.LogError("持久化重启后的任务状态失败: " + err.Error())
		}
	}
}

// ensureTaskLoaded 确保指定任务的元数据存在于内存。
func (cs *CommentService) ensureTaskLoaded(taskID string) error {
	cs.mu.RLock()
	_, exists := cs.tasks[taskID]
	cs.mu.RUnlock()
	if exists {
		return nil
	}

	index, err := cs.storage.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	var meta *storage.TaskMeta
	for i := range index.Tasks {
		if index.Tasks[i].TaskID == taskID {
			copyMeta := index.Tasks[i]
			meta = &copyMeta
			break
		}
	}
	if meta == nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	loaded := taskFromMeta(*meta)
	cs.mu.Lock()
	if _, exists := cs.tasks[taskID]; !exists {
		cs.tasks[taskID] = loaded
	}
	cs.mu.Unlock()
	return nil
}

func taskFromMeta(meta storage.TaskMeta) *ScrapeTask {
	pageLimit := meta.PageLimit
	if pageLimit <= 0 {
		pageLimit = 2
	}
	sortMode := meta.SortMode
	if sortMode == "" {
		sortMode = "time"
	}

	return &ScrapeTask{
		TaskID:         meta.TaskID,
		VideoID:        meta.VideoID,
		VideoTitle:     meta.VideoTitle,
		Status:         meta.Status,
		Comments:       nil,
		CommentsLoaded: false,
		Progress: TaskProgress{
			CurrentPage:   meta.CurrentPage,
			TotalComments: meta.CommentCount,
			PageLimit:     pageLimit,
		},
		StartTime:      meta.StartTime,
		EndTime:        meta.EndTime,
		Error:          meta.Error,
		AuthType:       meta.AuthType,
		PageLimit:      pageLimit,
		DelayMs:        meta.DelayMs,
		SortMode:       sortMode,
		IncludeReplies: meta.IncludeReplies,
	}
}

func cloneTaskLocked(task *ScrapeTask, includeComments bool) *ScrapeTask {
	if task == nil {
		return nil
	}
	copyTask := *task
	if includeComments {
		copyTask.Comments = cloneComments(task.Comments)
	} else {
		copyTask.Comments = nil
	}
	return &copyTask
}

func cloneComments(comments []bilibili.CommentData) []bilibili.CommentData {
	if comments == nil {
		return nil
	}
	result := make([]bilibili.CommentData, len(comments))
	for i, comment := range comments {
		result[i] = comment
		if comment.Replies != nil {
			result[i].Replies = cloneComments(comment.Replies)
		}
		if comment.Content.Emote != nil {
			result[i].Content.Emote = make(map[string]bilibili.Emote, len(comment.Content.Emote))
			for key, value := range comment.Content.Emote {
				result[i].Content.Emote[key] = value
			}
		}
		if comment.Content.JumpUrl != nil {
			result[i].Content.JumpUrl = make(map[string]bilibili.JumpUrl, len(comment.Content.JumpUrl))
			for key, value := range comment.Content.JumpUrl {
				result[i].Content.JumpUrl[key] = value
			}
		}
	}
	return result
}

// persistenceSnapshot 在一把读锁下同时生成任务快照和索引快照。
func (cs *CommentService) persistenceSnapshot(taskID string) (*ScrapeTask, *storage.TaskIndex, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	task := cs.tasks[taskID]
	if task == nil {
		return nil, nil, fmt.Errorf("task not found: %s", taskID)
	}
	taskSnapshot := cloneTaskLocked(task, task.CommentsLoaded)
	return taskSnapshot, cs.buildIndexLocked(), nil
}

func (cs *CommentService) persistTaskByID(taskID string) error {
	taskSnapshot, index, err := cs.persistenceSnapshot(taskID)
	if err != nil {
		return err
	}

	// 已卸载的 completed 评论已经安全存在于磁盘，避免用 nil 覆盖已有结果。
	if taskSnapshot.CommentsLoaded {
		if err := cs.storage.SaveTask(cs.convertToStorageFormat(taskSnapshot)); err != nil {
			return err
		}
	}
	return cs.storage.SaveIndex(index)
}

func (cs *CommentService) persistIndex() error {
	cs.mu.RLock()
	index := cs.buildIndexLocked()
	cs.mu.RUnlock()
	return cs.storage.SaveIndex(index)
}

func (cs *CommentService) buildIndexLocked() *storage.TaskIndex {
	metas := make([]storage.TaskMeta, 0, len(cs.tasks))
	for _, task := range cs.tasks {
		commentCount := task.Progress.TotalComments
		if commentCount == 0 && task.CommentsLoaded {
			commentCount = len(task.Comments)
		}
		metas = append(metas, storage.TaskMeta{
			TaskID:         task.TaskID,
			VideoID:        task.VideoID,
			VideoTitle:     task.VideoTitle,
			Status:         task.Status,
			CommentCount:   commentCount,
			CurrentPage:    task.Progress.CurrentPage,
			PageLimit:      task.PageLimit,
			DelayMs:        task.DelayMs,
			SortMode:       task.SortMode,
			IncludeReplies: task.IncludeReplies,
			AuthType:       task.AuthType,
			StartTime:      task.StartTime,
			EndTime:        task.EndTime,
			DataFile:       task.TaskID + ".json",
			Error:          task.Error,
		})
	}
	return &storage.TaskIndex{Tasks: metas}
}

// flushAllTasks 在 shutdown 最后写入所有仍驻留内存的快照，并一次性更新索引。
func (cs *CommentService) flushAllTasks() error {
	cs.mu.RLock()
	snapshots := make([]*ScrapeTask, 0, len(cs.tasks))
	for _, task := range cs.tasks {
		snapshots = append(snapshots, cloneTaskLocked(task, task.CommentsLoaded))
	}
	index := cs.buildIndexLocked()
	cs.mu.RUnlock()

	for _, task := range snapshots {
		if !task.CommentsLoaded {
			continue
		}
		if err := cs.storage.SaveTask(cs.convertToStorageFormat(task)); err != nil {
			return fmt.Errorf("flush task %s: %w", task.TaskID, err)
		}
	}
	return cs.storage.SaveIndex(index)
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
		// Cookie/AppKey/AppSecret are runtime-only and must never be persisted.
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

	comment := bilibili.CommentData{
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
	comment.Member.LevelInfo.CurrentLevel = e.Member.Level
	return comment
}

// Shutdown 优雅关闭服务
func (cs *CommentService) Shutdown(ctx context.Context) error {
	utils.LogInfo("Shutting down CommentService...")
	cs.cancel()

	done := make(chan struct{})
	go func() {
		cs.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if err := cs.flushAllTasks(); err != nil {
			utils.LogError("CommentService final flush failed: " + err.Error())
			return err
		}
		utils.LogInfo("CommentService shutdown complete")
		return nil
	case <-ctx.Done():
		utils.LogError("CommentService shutdown timeout")
		return ctx.Err()
	}
}
