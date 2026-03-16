package repository

import (
	"context"
	"errors"
	"time"
)

// ErrTaskNotFound is returned when a task does not exist or does not belong to the user.
var ErrTaskNotFound = errors.New("task not found")

// TaskRow represents a single row in the tasks table.
type TaskRow struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	Text       string    `db:"text"`
	Done       bool      `db:"done"`
	PriorityID string    `db:"priority_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// TaskCreate holds the fields needed to insert a new task.
type TaskCreate struct {
	ID         string
	UserID     string
	Text       string
	Done       bool
	PriorityID string
}

// TaskUpdate holds the mutable fields for updating a task.
type TaskUpdate struct {
	Done       *bool
	Text       *string
	PriorityID *string
}

// TaskRepository defines the data-access contract for tasks.
type TaskRepository interface {
	List(ctx context.Context, userID string) ([]TaskRow, error)
	Get(ctx context.Context, id string, userID string) (TaskRow, error)
	Create(ctx context.Context, t TaskCreate) (TaskRow, error)
	Update(ctx context.Context, id string, userID string, u TaskUpdate) (TaskRow, error)
	Delete(ctx context.Context, id string, userID string) error
}
