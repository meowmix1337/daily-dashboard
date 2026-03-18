package repository

import (
	"context"
	"errors"
	"time"
)

// ErrLabelNotFound is returned when a label does not exist or does not belong to the user.
var ErrLabelNotFound = errors.New("label not found")

// ErrLabelAssignmentNotFound is returned when a label assignment does not exist.
var ErrLabelAssignmentNotFound = errors.New("label assignment not found")

// TaskLabelRow represents a single row in the task_labels table.
type TaskLabelRow struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Name      string    `db:"name"`
	Color     string    `db:"color"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TaskLabelCreate holds the fields needed to insert a new task label.
type TaskLabelCreate struct {
	ID     string
	UserID string
	Name   string
	Color  string
}

// TaskLabelUpdate holds the mutable fields for updating a task label.
type TaskLabelUpdate struct {
	Name  *string
	Color *string
}

// TaskLabelAssignmentRow represents a single row in the task_label_assignments table.
type TaskLabelAssignmentRow struct {
	ID        string    `db:"id"`
	TaskID    string    `db:"task_id"`
	LabelID   string    `db:"label_id"`
	CreatedAt time.Time `db:"created_at"`
}

// TaskLabelAssignmentCreate holds the fields needed to insert a new label assignment.
type TaskLabelAssignmentCreate struct {
	ID      string
	TaskID  string
	LabelID string
}

// TaskLabelRepository defines the data-access contract for task labels.
type TaskLabelRepository interface {
	// List returns all active labels for the given user.
	List(ctx context.Context, userID string) ([]TaskLabelRow, error)
	// Get returns a single active label by ID for the given user.
	Get(ctx context.Context, id string, userID string) (TaskLabelRow, error)
	// Create inserts a new label and returns the created row.
	Create(ctx context.Context, l TaskLabelCreate) (TaskLabelRow, error)
	// Update applies partial updates to a label.
	Update(ctx context.Context, id string, userID string, u TaskLabelUpdate) error
	// Delete soft-deletes a label.
	Delete(ctx context.Context, id string, userID string) error
	// ListForTask returns all active labels assigned to the given task, scoped to the user.
	ListForTask(ctx context.Context, taskID string, userID string) ([]TaskLabelRow, error)
	// AssignLabel creates an assignment of a label to a task.
	AssignLabel(ctx context.Context, a TaskLabelAssignmentCreate) error
	// RemoveLabel soft-deletes the assignment of a label from a task, scoped to the user.
	RemoveLabel(ctx context.Context, taskID string, labelID string, userID string) error
}
