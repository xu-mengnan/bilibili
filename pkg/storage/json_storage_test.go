package storage

import (
	"testing"
	"time"
)

func TestCleanOldTasksDoesNotDeadlock(t *testing.T) {
	store := NewJSONStorage(t.TempDir())
	if err := store.Initialize(); err != nil {
		t.Fatal(err)
	}

	old := time.Now().Add(-2 * time.Hour)
	task := &TaskData{
		TaskID:      "old-task",
		Status:      "completed",
		Comments:    []CommentEntry{{RPID: 1}},
		StartTime:   old.Add(-time.Minute),
		EndTime:     old,
		PageLimit:   2,
		Progress:    TaskProgressEntry{TotalComments: 1, PageLimit: 2},
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveIndex(&TaskIndex{Tasks: []TaskMeta{{
		TaskID:       task.TaskID,
		Status:       task.Status,
		CommentCount: 1,
		PageLimit:    2,
		StartTime:    task.StartTime,
		EndTime:      task.EndTime,
		DataFile:     task.TaskID + ".json",
	}}}); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- store.CleanOldTasks(time.Now().Add(-time.Hour))
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("CleanOldTasks timed out; possible lock re-entry deadlock")
	}

	tasks, err := store.ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected cleaned index, got %d tasks", len(tasks))
	}
}

func TestCleanOldTasksKeepsRunningTaskWithZeroEndTime(t *testing.T) {
	store := NewJSONStorage(t.TempDir())
	if err := store.Initialize(); err != nil {
		t.Fatal(err)
	}

	task := &TaskData{
		TaskID:    "running-task",
		Status:    "running",
		StartTime: time.Now(),
		PageLimit: 2,
		Progress:  TaskProgressEntry{PageLimit: 2},
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveIndex(&TaskIndex{Tasks: []TaskMeta{{
		TaskID:    task.TaskID,
		Status:    task.Status,
		PageLimit: 2,
		StartTime: task.StartTime,
		DataFile:  task.TaskID + ".json",
	}}}); err != nil {
		t.Fatal(err)
	}

	if err := store.CleanOldTasks(time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	tasks, err := store.ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].TaskID != task.TaskID {
		t.Fatalf("running task with zero end time must be retained: %#v", tasks)
	}
}

func TestSaveTaskRoundTripAfterReplacement(t *testing.T) {
	store := NewJSONStorage(t.TempDir())
	if err := store.Initialize(); err != nil {
		t.Fatal(err)
	}

	first := &TaskData{TaskID: "task", Status: "running", VideoTitle: "before"}
	if err := store.SaveTask(first); err != nil {
		t.Fatal(err)
	}

	second := &TaskData{TaskID: "task", Status: "completed", VideoTitle: "after"}
	if err := store.SaveTask(second); err != nil {
		t.Fatal(err)
	}

	got, err := store.LoadTask("task")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "completed" || got.VideoTitle != "after" {
		t.Fatalf("unexpected round trip: %#v", got)
	}
}
