package repository

import (
	"context"

	"github.com/meowmix1337/argus/backend/internal/model"
)

// TaskRepository defines the data-access contract for tasks.
type TaskRepository interface {
	List(ctx context.Context, userID string) ([]model.Task, error)
	Get(ctx context.Context, id string, userID string) (model.Task, error)
	Create(ctx context.Context, t model.TaskCreate) (model.Task, error)
	Update(ctx context.Context, id string, userID string, u model.TaskUpdate) error
	Delete(ctx context.Context, id string, userID string) error
}
