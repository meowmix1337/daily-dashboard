package repository

import (
	"context"
	"errors"
	"time"
)

// ErrSettingsNotFound is returned when user settings do not exist for the user.
var ErrSettingsNotFound = errors.New("user settings not found")

// ErrCategoryNotFound is returned when a news category does not exist.
var ErrCategoryNotFound = errors.New("news category not found")

// UserSettingsRow represents a single row in the user_settings table.
type UserSettingsRow struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	Latitude       *float64  `db:"latitude"`
	Longitude      *float64  `db:"longitude"`
	CalendarICSURL *string   `db:"calendar_ics_url"`
	Timezone       *string   `db:"timezone"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// NewsCategoryTypeRow represents a single row in the news_category_types table.
type NewsCategoryTypeRow struct {
	ID        string `db:"id"`
	Label     string `db:"label"`
	SortOrder int    `db:"sort_order"`
}

// UserNewsCategoryRow represents a single row in the user_news_categories table.
type UserNewsCategoryRow struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	CategoryID string    `db:"category_id"`
	CreatedAt  time.Time `db:"created_at"`
}

// UserSettingsUpsert holds the mutable fields for creating or updating user settings.
type UserSettingsUpsert struct {
	Latitude       *float64
	Longitude      *float64
	CalendarICSURL *string
	Timezone       *string
}

// UserSettingsRepository defines the data-access contract for user settings.
type UserSettingsRepository interface {
	// Get returns the settings for the given user, or ErrSettingsNotFound if none exist.
	Get(ctx context.Context, userID string) (UserSettingsRow, error)
	// Upsert creates or updates settings for the given user, returning the final row.
	Upsert(ctx context.Context, userID string, u UserSettingsUpsert) (UserSettingsRow, error)
	// ListAllCategories returns all news category types ordered by sort_order.
	ListAllCategories(ctx context.Context) ([]NewsCategoryTypeRow, error)
	// ListSelectedCategories returns the news categories the user has selected.
	ListSelectedCategories(ctx context.Context, userID string) ([]NewsCategoryTypeRow, error)
	// SetSelectedCategories replaces the user's selected categories (soft-deletes old, inserts new).
	SetSelectedCategories(ctx context.Context, userID string, categoryIDs []string) error
}
