package storage

import "time"

type TaskIndex struct {
	Version     string     `json:"version"`
	LastUpdated time.Time  `json:"last_updated"`
	Tasks       []TaskMeta `json:"tasks"`
}

type TaskMeta struct {
	TaskID       string    `json:"task_id"`
	VideoID      string    `json:"video_id"`
	VideoTitle   string    `json:"video_title"`
	Status       string    `json:"status"`
	CommentCount int       `json:"comment_count"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	DataFile     string    `json:"data_file"`
	Error        string    `json:"error,omitempty"`
}

// TaskData 只保存可持久化业务数据。认证凭据必须仅驻留内存。
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
	PageLimit      int               `json:"page_limit"`
	DelayMs        int               `json:"delay_ms"`
	SortMode       string            `json:"sort_mode"`
	IncludeReplies bool              `json:"include_replies"`
}

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

type CommentContent struct {
	Message string `json:"message"`
}

type CommentMember struct {
	Mid    string `json:"mid"`
	Name   string `json:"name"`
	Sex    string `json:"sex"`
	Avatar string `json:"avatar"`
	Sign   string `json:"sign"`
	Rank   string `json:"rank"`
	Level  int    `json:"level"`
}

type TaskProgressEntry struct {
	CurrentPage   int `json:"current_page"`
	TotalComments int `json:"total_comments"`
	PageLimit     int `json:"page_limit"`
}
