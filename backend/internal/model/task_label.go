package model

import "time"

// TaskLabel is a user-defined tag that can be assigned to tasks.
type TaskLabel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskLabelCreate holds fields for creating a new label.
type TaskLabelCreate struct {
	ID     string
	UserID string
	Name   string
	Color  string
}

// TaskLabelUpdate holds optional fields for updating a label.
type TaskLabelUpdate struct {
	Name  *string
	Color *string
}

// TaskLabelAssignmentCreate holds fields for assigning a label to a task.
type TaskLabelAssignmentCreate struct {
	ID      string
	TaskID  string
	LabelID string
}
