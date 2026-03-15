package service

import (
	"context"
	"testing"

	"github.com/daily-dashboard/backend/internal/model"
)

func TestTasksService_List(t *testing.T) {
	svc := NewTasksService()
	tasks, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("expected default tasks to be seeded")
	}
}

func TestTasksService_Create(t *testing.T) {
	svc := NewTasksService()
	initial, _ := svc.List(context.Background())

	created, err := svc.Create(context.Background(), model.Task{Text: "new task", Priority: "high"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected ID to be assigned")
	}
	if created.Text != "new task" {
		t.Errorf("got text %q, want %q", created.Text, "new task")
	}

	after, _ := svc.List(context.Background())
	if len(after) != len(initial)+1 {
		t.Errorf("expected %d tasks, got %d", len(initial)+1, len(after))
	}
}

func TestTasksService_Create_DefaultPriority(t *testing.T) {
	svc := NewTasksService()
	created, err := svc.Create(context.Background(), model.Task{Text: "no priority"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Priority != "medium" {
		t.Errorf("got priority %q, want medium", created.Priority)
	}
}

func TestTasksService_Update(t *testing.T) {
	svc := NewTasksService()
	tasks, _ := svc.List(context.Background())
	id := tasks[0].ID

	updated, err := svc.Update(context.Background(), id, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated.Done {
		t.Error("expected task to be marked done")
	}
}

func TestTasksService_Update_NotFound(t *testing.T) {
	svc := NewTasksService()
	_, err := svc.Update(context.Background(), "nonexistent-id", true)
	if err == nil {
		t.Fatal("expected error for missing task")
	}
}

func TestTasksService_Delete(t *testing.T) {
	svc := NewTasksService()
	tasks, _ := svc.List(context.Background())
	id := tasks[0].ID
	initial := len(tasks)

	if err := svc.Delete(context.Background(), id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after, _ := svc.List(context.Background())
	if len(after) != initial-1 {
		t.Errorf("expected %d tasks after delete, got %d", initial-1, len(after))
	}
}

func TestTasksService_Delete_NotFound(t *testing.T) {
	svc := NewTasksService()
	if err := svc.Delete(context.Background(), "nonexistent-id"); err == nil {
		t.Fatal("expected error for missing task")
	}
}

func TestTasksService_List_ReturnsCopy(t *testing.T) {
	svc := NewTasksService()
	tasks, _ := svc.List(context.Background())
	tasks[0].Text = "mutated"

	again, _ := svc.List(context.Background())
	if again[0].Text == "mutated" {
		t.Error("List should return a copy, not a reference to the internal slice")
	}
}
