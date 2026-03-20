package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/repository"
)

// ErrLabelNotFound is returned when a label does not exist or does not belong to the user.
var ErrLabelNotFound = errors.New("label not found")

// ErrLabelValidation is returned when label input fails validation.
var ErrLabelValidation = errors.New("label validation failed")

// ErrLabelAlreadyAssigned is returned when a label is already assigned to a task.
var ErrLabelAlreadyAssigned = errors.New("label already assigned to task")

// TaskLabelsService manages task labels via a repository.
type TaskLabelsService struct {
	repo repository.TaskLabelRepository
}

// NewTaskLabelsService creates a new TaskLabelsService backed by the given repository.
func NewTaskLabelsService(repo repository.TaskLabelRepository) *TaskLabelsService {
	return &TaskLabelsService{repo: repo}
}

// List returns all active labels for the given user.
func (s *TaskLabelsService) List(ctx context.Context, userID string) ([]model.TaskLabel, error) {
	rows, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list labels: %w", err)
	}
	labels := make([]model.TaskLabel, 0, len(rows))
	for _, r := range rows {
		labels = append(labels, labelRowToModel(r))
	}
	return labels, nil
}

// Create adds a new label for the given user.
func (s *TaskLabelsService) Create(ctx context.Context, userID string, name string, color string) (model.TaskLabel, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.TaskLabel{}, fmt.Errorf("%w: label name cannot be empty", ErrLabelValidation)
	}
	if len(name) > 16 {
		return model.TaskLabel{}, fmt.Errorf("%w: label name exceeds 16 characters", ErrLabelValidation)
	}

	if color == "" {
		color = "#6366f1"
	}

	id, err := uuid.NewV7()
	if err != nil {
		slog.Error("failed to generate UUID v7", "error", err)
		return model.TaskLabel{}, fmt.Errorf("generate label id: %w", err)
	}

	row, err := s.repo.Create(ctx, repository.TaskLabelCreate{
		ID:     id.String(),
		UserID: userID,
		Name:   name,
		Color:  color,
	})
	if err != nil {
		return model.TaskLabel{}, fmt.Errorf("create label: %w", err)
	}
	return labelRowToModel(row), nil
}

// Update modifies a label by ID, scoped to the given user.
func (s *TaskLabelsService) Update(ctx context.Context, id string, userID string, name *string, color *string) (model.TaskLabel, error) {
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return model.TaskLabel{}, fmt.Errorf("%w: label name cannot be empty", ErrLabelValidation)
		}
		if len(trimmed) > 16 {
			return model.TaskLabel{}, fmt.Errorf("%w: label name exceeds 16 characters", ErrLabelValidation)
		}
		name = &trimmed
	}

	// Verify the label exists and belongs to this user.
	if _, err := s.repo.Get(ctx, id, userID); err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return model.TaskLabel{}, ErrLabelNotFound
		}
		return model.TaskLabel{}, fmt.Errorf("get label: %w", err)
	}

	if err := s.repo.Update(ctx, id, userID, repository.TaskLabelUpdate{
		Name:  name,
		Color: color,
	}); err != nil {
		return model.TaskLabel{}, fmt.Errorf("update label: %w", err)
	}

	row, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		return model.TaskLabel{}, fmt.Errorf("fetch updated label: %w", err)
	}
	return labelRowToModel(row), nil
}

// Delete soft-deletes a label by ID, scoped to the given user.
func (s *TaskLabelsService) Delete(ctx context.Context, id string, userID string) error {
	// Verify the label exists and belongs to this user.
	if _, err := s.repo.Get(ctx, id, userID); err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return ErrLabelNotFound
		}
		return fmt.Errorf("get label: %w", err)
	}

	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete label: %w", err)
	}
	return nil
}

// ListForTask returns all active labels assigned to the given task, scoped to the user.
func (s *TaskLabelsService) ListForTask(ctx context.Context, taskID string, userID string) ([]model.TaskLabel, error) {
	rows, err := s.repo.ListForTask(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("list labels for task: %w", err)
	}
	labels := make([]model.TaskLabel, 0, len(rows))
	for _, r := range rows {
		labels = append(labels, labelRowToModel(r))
	}
	return labels, nil
}

// AssignLabel assigns a label to a task.
func (s *TaskLabelsService) AssignLabel(ctx context.Context, taskID string, labelID string, userID string) error {
	// Verify the label exists and belongs to this user.
	if _, err := s.repo.Get(ctx, labelID, userID); err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return ErrLabelNotFound
		}
		return fmt.Errorf("get label: %w", err)
	}

	// Check if already assigned.
	existing, err := s.repo.ListForTask(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("list labels for task: %w", err)
	}
	for _, l := range existing {
		if l.ID == labelID {
			return ErrLabelAlreadyAssigned
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		slog.Error("failed to generate UUID v7", "error", err)
		return fmt.Errorf("generate assignment id: %w", err)
	}

	if err := s.repo.AssignLabel(ctx, repository.TaskLabelAssignmentCreate{
		ID:      id.String(),
		TaskID:  taskID,
		LabelID: labelID,
	}); err != nil {
		if errors.Is(err, repository.ErrLabelAlreadyAssigned) {
			return ErrLabelAlreadyAssigned
		}
		return fmt.Errorf("assign label: %w", err)
	}
	return nil
}

// RemoveLabel removes a label assignment from a task.
func (s *TaskLabelsService) RemoveLabel(ctx context.Context, taskID string, labelID string, userID string) error {
	if err := s.repo.RemoveLabel(ctx, taskID, labelID, userID); err != nil {
		return fmt.Errorf("remove label: %w", err)
	}
	return nil
}

func labelRowToModel(r repository.TaskLabelRow) model.TaskLabel {
	return model.TaskLabel{
		ID:        r.ID,
		Name:      r.Name,
		Color:     r.Color,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	}
}
