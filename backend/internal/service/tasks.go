package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/daily-dashboard/backend/internal/model"
)

// TasksService manages an in-memory list of tasks.
type TasksService struct {
	mu    sync.Mutex
	tasks []model.Task
}

// NewTasksService creates a new TasksService with some default tasks.
func NewTasksService() *TasksService {
	return &TasksService{
		tasks: []model.Task{
			{ID: uuid.NewString(), Text: "Review PR #482 — auth refactor", Done: true, Priority: "high"},
			{ID: uuid.NewString(), Text: "Update Helm chart values for staging", Done: false, Priority: "high"},
			{ID: uuid.NewString(), Text: "Write ADR for event sourcing migration", Done: false, Priority: "medium"},
			{ID: uuid.NewString(), Text: "Fix flaky integration test in CI", Done: false, Priority: "medium"},
			{ID: uuid.NewString(), Text: "Prepare demo for stakeholder sync", Done: false, Priority: "low"},
		},
	}
}

// List returns all tasks.
func (s *TasksService) List(ctx context.Context) ([]model.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]model.Task, len(s.tasks))
	copy(result, s.tasks)
	return result, nil
}

// Create adds a new task.
func (s *TasksService) Create(ctx context.Context, task model.Task) (model.Task, error) {
	task.ID = uuid.NewString()
	if task.Priority == "" {
		task.Priority = "medium"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, task)
	return task, nil
}

// Update sets the done status of a task by ID.
func (s *TasksService) Update(ctx context.Context, id string, done bool) (model.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tasks {
		if t.ID == id {
			s.tasks[i].Done = done
			return s.tasks[i], nil
		}
	}
	return model.Task{}, fmt.Errorf("task %s not found", id)
}

// Delete removes a task by ID.
func (s *TasksService) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tasks {
		if t.ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task %s not found", id)
}
