package services

import (
	"context"
	"testing"
	"time"

	"bilibili/pkg/bilibili"
	"bilibili/pkg/storage"
)

func TestGetTaskProgressReturnsDetachedSnapshot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := storage.NewJSONStorage(t.TempDir())
	service := NewCommentService(ctx, store)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		if err := service.Shutdown(shutdownCtx); err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	}()

	service.mu.Lock()
	service.tasks["task"] = &ScrapeTask{
		TaskID:         "task",
		Status:         "running",
		VideoTitle:     "original",
		PageLimit:      3,
		CommentsLoaded: true,
		Progress:       TaskProgress{CurrentPage: 1, PageLimit: 3},
	}
	service.mu.Unlock()

	snapshot, err := service.GetTaskProgress("task")
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Status = "tampered"
	snapshot.VideoTitle = "tampered"
	snapshot.Progress.CurrentPage = 99

	again, err := service.GetTaskProgress("task")
	if err != nil {
		t.Fatal(err)
	}
	if again.Status != "running" || again.VideoTitle != "original" || again.Progress.CurrentPage != 1 {
		t.Fatalf("internal state leaked through snapshot: %#v", again)
	}
	if again.Comments != nil {
		t.Fatalf("progress snapshot should not load comments: %#v", again.Comments)
	}
}

func TestRestartMetadataAndLazyCommentLoad(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewJSONStorage(dir)
	if err := store.Initialize(); err != nil {
		t.Fatal(err)
	}

	taskID := "persisted-task"
	entry := storage.CommentEntry{
		RPID:    42,
		Ctime:   int(time.Now().Unix()),
		Like:    9,
		Content: storage.CommentContent{Message: "hello"},
		Member:  storage.CommentMember{Name: "tester", Level: 6},
	}
	taskData := &storage.TaskData{
		TaskID:         taskID,
		VideoID:        "BV1test",
		VideoTitle:     "video",
		Status:         "completed",
		Comments:       []storage.CommentEntry{entry},
		Progress:       storage.TaskProgressEntry{CurrentPage: 5, TotalComments: 1, PageLimit: 7},
		StartTime:      time.Now().Add(-time.Minute),
		EndTime:        time.Now(),
		PageLimit:      7,
		DelayMs:        450,
		SortMode:       "hot",
		IncludeReplies: true,
	}
	if err := store.SaveTask(taskData); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveIndex(&storage.TaskIndex{Tasks: []storage.TaskMeta{{
		TaskID:         taskID,
		VideoID:        taskData.VideoID,
		VideoTitle:     taskData.VideoTitle,
		Status:         taskData.Status,
		CommentCount:   1,
		CurrentPage:    5,
		PageLimit:      7,
		DelayMs:        450,
		SortMode:       "hot",
		IncludeReplies: true,
		StartTime:      taskData.StartTime,
		EndTime:        taskData.EndTime,
		DataFile:       taskID + ".json",
	}}}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service := NewCommentService(ctx, storage.NewJSONStorage(dir))
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		if err := service.Shutdown(shutdownCtx); err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	}()

	meta, err := service.GetTaskProgress(taskID)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Progress.TotalComments != 1 || meta.Progress.CurrentPage != 5 || meta.Progress.PageLimit != 7 {
		t.Fatalf("metadata not restored: %#v", meta.Progress)
	}
	if meta.PageLimit != 7 || meta.DelayMs != 450 || meta.SortMode != "hot" || !meta.IncludeReplies {
		t.Fatalf("task options not restored: %#v", meta)
	}
	if meta.Comments != nil {
		t.Fatal("metadata read must not load comments")
	}

	detail, err := service.GetTaskWithComments(taskID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Comments) != 1 {
		t.Fatalf("expected one lazily loaded comment, got %d", len(detail.Comments))
	}
	if detail.Comments[0].Member.LevelInfo.CurrentLevel != 6 {
		t.Fatalf("member level not restored: %#v", detail.Comments[0].Member.LevelInfo)
	}

	// Returned comment data is detached from service state.
	detail.Comments[0].Content.Message = "changed outside"
	detailAgain, err := service.GetTaskWithComments(taskID)
	if err != nil {
		t.Fatal(err)
	}
	if detailAgain.Comments[0].Content.Message != "hello" {
		t.Fatal("comment snapshot mutation leaked back into service state")
	}
}

func TestCloneCommentsCopiesNestedMutableFields(t *testing.T) {
	comment := bilibili.CommentData{}
	comment.Content.Message = "root"
	comment.Content.Emote = map[string]bilibili.Emote{"x": {Text: "x"}}
	comment.Content.JumpUrl = map[string]bilibili.JumpUrl{"u": {Title: "u"}}
	comment.Replies = []bilibili.CommentData{{}}
	comment.Replies[0].Content.Message = "reply"

	cloned := cloneComments([]bilibili.CommentData{comment})
	cloned[0].Content.Emote["x"] = bilibili.Emote{Text: "changed"}
	cloned[0].Content.JumpUrl["u"] = bilibili.JumpUrl{Title: "changed"}
	cloned[0].Replies[0].Content.Message = "changed"

	if comment.Content.Emote["x"].Text != "x" ||
		comment.Content.JumpUrl["u"].Title != "u" ||
		comment.Replies[0].Content.Message != "reply" {
		t.Fatal("cloneComments did not detach nested mutable state")
	}
}
