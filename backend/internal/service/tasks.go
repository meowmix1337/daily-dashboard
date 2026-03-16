package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/repository"
)

// ErrTaskNotFound is returned when a task does not exist or does not belong to the user.
var ErrTaskNotFound = errors.New("task not found")

// ErrTaskValidation is returned when task input fails validation.
var ErrTaskValidation = errors.New("task validation failed")

// TasksService manages tasks via a repository.
type TasksService struct {
	repo repository.TaskRepository
}

// NewTasksService creates a new TasksService backed by the given repository.
func NewTasksService(repo repository.TaskRepository) *TasksService {
	return &TasksService{repo: repo}
}

// List returns all tasks for the given user.
func (s *TasksService) List(ctx context.Context, userID string) ([]model.Task, error) {
	rows, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	tasks := make([]model.Task, 0, len(rows))
	for _, r := range rows {
		tasks = append(tasks, rowToModel(r))
	}
	return tasks, nil
}

// Create adds a new task for the given user.
func (s *TasksService) Create(ctx context.Context, userID string, text string, priority string) (model.Task, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return model.Task{}, fmt.Errorf("%w: task text cannot be empty", ErrTaskValidation)
	}

	if priority == "" {
		priority = "medium"
	}
	if priority != "high" && priority != "medium" && priority != "low" {
		return model.Task{}, fmt.Errorf("%w: invalid priority %q", ErrTaskValidation, priority)
	}

	id, err := uuid.NewV7()
	if err != nil {
		slog.Error("failed to generate UUID v7", "error", err)
		return model.Task{}, fmt.Errorf("generate task id: %w", err)
	}

	row, err := s.repo.Create(ctx, repository.TaskCreate{
		ID:         id.String(),
		UserID:     userID,
		Text:       text,
		Done:       false,
		PriorityID: priority,
	})
	if err != nil {
		return model.Task{}, fmt.Errorf("create task: %w", err)
	}
	return rowToModel(row), nil
}

// Update modifies a task by ID, scoped to the given user.
func (s *TasksService) Update(ctx context.Context, id string, userID string, done *bool, text *string, priority *string) (model.Task, error) {
	if text != nil {
		trimmed := strings.TrimSpace(*text)
		if trimmed == "" {
			return model.Task{}, fmt.Errorf("%w: task text cannot be empty", ErrTaskValidation)
		}
		text = &trimmed
	}
	if priority != nil && *priority != "high" && *priority != "medium" && *priority != "low" {
		return model.Task{}, fmt.Errorf("%w: invalid priority %q", ErrTaskValidation, *priority)
	}

	// 1. Verify the task exists and belongs to this user.
	if _, err := s.repo.Get(ctx, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repository.ErrTaskNotFound) {
			return model.Task{}, ErrTaskNotFound
		}
		return model.Task{}, fmt.Errorf("get task: %w", err)
	}

	// 2. Apply the update.
	if err := s.repo.Update(ctx, id, userID, repository.TaskUpdate{
		Done:       done,
		Text:       text,
		PriorityID: priority,
	}); err != nil {
		return model.Task{}, fmt.Errorf("update task: %w", err)
	}

	// 3. Re-fetch to return the current state.
	row, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		return model.Task{}, fmt.Errorf("fetch updated task: %w", err)
	}
	return rowToModel(row), nil
}

// Delete soft-deletes a task by ID, scoped to the given user.
func (s *TasksService) Delete(ctx context.Context, id string, userID string) error {
	// 1. Verify the task exists and belongs to this user.
	if _, err := s.repo.Get(ctx, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repository.ErrTaskNotFound) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("get task: %w", err)
	}

	// 2. Soft-delete it.
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func rowToModel(r repository.TaskRow) model.Task {
	return model.Task{
		ID:       r.ID,
		Text:     r.Text,
		Done:     r.Done,
		Priority: r.PriorityID,
	}
}
