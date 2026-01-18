package storage

import "time"

// TaskIndex 任务索引文件结构
type TaskIndex struct {
	Version     string     `json:"version"`      // 版本号
	LastUpdated time.Time  `json:"last_updated"` // 最后更新时间
	Tasks       []TaskMeta `json:"tasks"`        // 任务元数据列表
}

// TaskMeta 任务元数据（不含评论数据，用于索引）
type TaskMeta struct {
	TaskID       string    `json:"task_id"`
	VideoID      string    `json:"video_id"`
	VideoTitle   string    `json:"video_title"`
	Status       string    `json:"status"` // running, completed, failed
	CommentCount int       `json:"comment_count"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	DataFile     string    `json:"data_file"` // 数据文件名
	Error        string    `json:"error,omitempty"`
}

// TaskData 单个任务的完整数据（包含评论）
type TaskData struct {
	TaskID         string            `json:"task_id"`
	VideoID        string            `json:"video_id"`
	VideoTitle     string            `json:"video_title"`
	Status         string            `json:"status"`
	Comments       []CommentEntry    `json:"comments"`
	Progress       TaskProgressEntry `json:"progress"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        time.Time         `json:"end_time"`
	Error          string            `json:"error,omitempty"`
	AuthType       string            `json:"auth_type"`
	Cookie         string            `json:"cookie,omitempty"`
	AppKey         string            `json:"app_key,omitempty"`
	AppSecret      string            `json:"app_secret,omitempty"`
	PageLimit      int               `json:"page_limit"`
	DelayMs        int               `json:"delay_ms"`
	SortMode       string            `json:"sort_mode"`
	IncludeReplies bool              `json:"include_replies"`
}

// CommentEntry 评论数据（存储层专用，避免循环引用）
type CommentEntry struct {
	RPID      int64          `json:"rpid"`
	OID       int64          `json:"oid"`
	Type      int            `json:"type"`
	Mid       int64          `json:"mid"`
	Root      int64          `json:"root"`
	Parent    int64          `json:"parent"`
	Dialog    int64          `json:"dialog"`
	Count     int            `json:"count"`
	RCount    int            `json:"rcount"`
	State     int            `json:"state"`
	FansGrade int            `json:"fans_grade"`
	Attr      int            `json:"attr"`
	Ctime     int            `json:"ctime"`
	Like      int            `json:"like"`
	Content   CommentContent `json:"content"`
	Member    CommentMember  `json:"member"`
	Replies   []CommentEntry `json:"replies"`
}

// CommentContent 评论内容
type CommentContent struct {
	Message string `json:"message"`
}

// CommentMember 评论用户信息
type CommentMember struct {
	Mid    string `json:"mid"`
	Name   string `json:"name"`
	Sex    string `json:"sex"`
	Avatar string `json:"avatar"`
	Sign   string `json:"sign"`
	Rank   string `json:"rank"`
	Level  int    `json:"level"`
}

// TaskProgressEntry 任务进度
type TaskProgressEntry struct {
	CurrentPage   int `json:"current_page"`
	TotalComments int `json:"total_comments"`
	PageLimit     int `json:"page_limit"`
}
