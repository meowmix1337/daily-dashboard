package model

// TaskLabel is a user-defined tag that can be assigned to tasks.
type TaskLabel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}
