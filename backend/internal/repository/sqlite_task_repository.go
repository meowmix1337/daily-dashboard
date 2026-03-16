package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// sqliteTaskRow mirrors the tasks table with string timestamps for SQLite scanning.
type sqliteTaskRow struct {
	ID         string `db:"id"`
	UserID     string `db:"user_id"`
	Text       string `db:"text"`
	Done       int    `db:"done"`
	PriorityID string `db:"priority_id"`
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

const timeFormat = "2006-01-02T15:04:05.000Z"

func (r *sqliteTaskRow) toTaskRow() (TaskRow, error) {
	createdAt, err := time.Parse(timeFormat, r.CreatedAt)
	if err != nil {
		return TaskRow{}, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(timeFormat, r.UpdatedAt)
	if err != nil {
		return TaskRow{}, fmt.Errorf("parse updated_at: %w", err)
	}
	return TaskRow{
		ID:         r.ID,
		UserID:     r.UserID,
		Text:       r.Text,
		Done:       r.Done == 1,
		PriorityID: r.PriorityID,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// SQLiteTaskRepository implements TaskRepository backed by SQLite via sqlx.
type SQLiteTaskRepository struct {
	db *sqlx.DB
}

// NewSQLiteTaskRepository creates a new SQLiteTaskRepository.
func NewSQLiteTaskRepository(db *sqlx.DB) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{db: db}
}

func (r *SQLiteTaskRepository) List(ctx context.Context, userID string) ([]TaskRow, error) {
	var rows []sqliteTaskRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, user_id, text, done, priority_id, created_at, updated_at
		 FROM tasks
		 WHERE deleted_at IS NULL AND user_id = ?
		 ORDER BY
		   CASE priority_id
		     WHEN 'high' THEN 1
		     WHEN 'medium' THEN 2
		     WHEN 'low' THEN 3
		   END,
		   created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	result := make([]TaskRow, 0, len(rows))
	for i := range rows {
		tr, err := rows[i].toTaskRow()
		if err != nil {
			return nil, err
		}
		result = append(result, tr)
	}
	return result, nil
}

func (r *SQLiteTaskRepository) Get(ctx context.Context, id string, userID string) (TaskRow, error) {
	var row sqliteTaskRow
	err := r.db.GetContext(ctx, &row,
		`SELECT id, user_id, text, done, priority_id, created_at, updated_at
		 FROM tasks
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return TaskRow{}, ErrTaskNotFound
		}
		return TaskRow{}, fmt.Errorf("get task: %w", err)
	}
	return row.toTaskRow()
}

func (r *SQLiteTaskRepository) Create(ctx context.Context, t TaskCreate) (TaskRow, error) {
	now := time.Now().UTC().Format(timeFormat)
	doneInt := 0
	if t.Done {
		doneInt = 1
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (id, user_id, text, done, priority_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.UserID, t.Text, doneInt, t.PriorityID, now, now,
	)
	if err != nil {
		return TaskRow{}, fmt.Errorf("create task: %w", err)
	}
	return r.Get(ctx, t.ID, t.UserID)
}

func (r *SQLiteTaskRepository) Update(ctx context.Context, id string, userID string, u TaskUpdate) (TaskRow, error) {
	now := time.Now().UTC().Format(timeFormat)

	// Build a single UPDATE statement with all changed fields.
	setClauses := []string{"updated_at = ?"}
	args := []interface{}{now}

	if u.Done != nil {
		doneInt := 0
		if *u.Done {
			doneInt = 1
		}
		setClauses = append(setClauses, "done = ?")
		args = append(args, doneInt)
	}
	if u.Text != nil {
		setClauses = append(setClauses, "text = ?")
		args = append(args, *u.Text)
	}
	if u.PriorityID != nil {
		setClauses = append(setClauses, "priority_id = ?")
		args = append(args, *u.PriorityID)
	}

	query := "UPDATE tasks SET " + strings.Join(setClauses, ", ") +
		" WHERE id = ? AND user_id = ? AND deleted_at IS NULL"
	args = append(args, id, userID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return TaskRow{}, fmt.Errorf("update task: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return TaskRow{}, fmt.Errorf("update task rows affected: %w", err)
	}
	if rows == 0 {
		return TaskRow{}, ErrTaskNotFound
	}

	return r.Get(ctx, id, userID)
}

func (r *SQLiteTaskRepository) Delete(ctx context.Context, id string, userID string) error {
	now := time.Now().UTC().Format(timeFormat)
	result, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET deleted_at = ?, updated_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		now, now, id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete task rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTaskNotFound
	}
	return nil
}
