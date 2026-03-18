package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// sqliteUserSettingsRow mirrors the user_settings table with string timestamps and nullable SQL types.
type sqliteUserSettingsRow struct {
	ID             string          `db:"id"`
	UserID         string          `db:"user_id"`
	Latitude       sql.NullFloat64 `db:"latitude"`
	Longitude      sql.NullFloat64 `db:"longitude"`
	CalendarICSURL sql.NullString  `db:"calendar_ics_url"`
	Timezone       sql.NullString  `db:"timezone"`
	CreatedAt      string          `db:"created_at"`
	UpdatedAt      string          `db:"updated_at"`
}

func (r *sqliteUserSettingsRow) toUserSettingsRow() (UserSettingsRow, error) {
	createdAt, err := time.Parse(timeFormat, r.CreatedAt)
	if err != nil {
		return UserSettingsRow{}, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(timeFormat, r.UpdatedAt)
	if err != nil {
		return UserSettingsRow{}, fmt.Errorf("parse updated_at: %w", err)
	}

	row := UserSettingsRow{
		ID:        r.ID,
		UserID:    r.UserID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	if r.Latitude.Valid {
		v := r.Latitude.Float64
		row.Latitude = &v
	}
	if r.Longitude.Valid {
		v := r.Longitude.Float64
		row.Longitude = &v
	}
	if r.CalendarICSURL.Valid {
		v := r.CalendarICSURL.String
		row.CalendarICSURL = &v
	}
	if r.Timezone.Valid {
		v := r.Timezone.String
		row.Timezone = &v
	}
	return row, nil
}

// SQLiteUserSettingsRepository implements UserSettingsRepository backed by SQLite via sqlx.
type SQLiteUserSettingsRepository struct {
	db *sqlx.DB
}

// NewSQLiteUserSettingsRepository creates a new SQLiteUserSettingsRepository.
func NewSQLiteUserSettingsRepository(db *sqlx.DB) *SQLiteUserSettingsRepository {
	return &SQLiteUserSettingsRepository{db: db}
}

func (r *SQLiteUserSettingsRepository) Get(ctx context.Context, userID string) (UserSettingsRow, error) {
	var row sqliteUserSettingsRow
	err := r.db.GetContext(ctx, &row,
		`SELECT id, user_id, latitude, longitude, calendar_ics_url, timezone, created_at, updated_at
		 FROM user_settings
		 WHERE user_id = ? AND deleted_at IS NULL`,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserSettingsRow{}, ErrSettingsNotFound
		}
		return UserSettingsRow{}, fmt.Errorf("get user settings: %w", err)
	}
	return row.toUserSettingsRow()
}

func (r *SQLiteUserSettingsRepository) Upsert(ctx context.Context, userID string, u UserSettingsUpsert) (UserSettingsRow, error) {
	now := time.Now().UTC().Format(timeFormat)

	existing, err := r.Get(ctx, userID)
	if err != nil && !errors.Is(err, ErrSettingsNotFound) {
		return UserSettingsRow{}, fmt.Errorf("upsert user settings: %w", err)
	}

	if errors.Is(err, ErrSettingsNotFound) {
		// INSERT new row.
		id, genErr := uuid.NewV7()
		if genErr != nil {
			return UserSettingsRow{}, fmt.Errorf("generate settings id: %w", genErr)
		}
		_, execErr := r.db.ExecContext(ctx,
			`INSERT INTO user_settings (id, user_id, latitude, longitude, calendar_ics_url, timezone, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id.String(), userID, u.Latitude, u.Longitude, u.CalendarICSURL, u.Timezone, now, now,
		)
		if execErr != nil {
			return UserSettingsRow{}, fmt.Errorf("insert user settings: %w", execErr)
		}
	} else {
		// UPDATE existing row.
		_, execErr := r.db.ExecContext(ctx,
			`UPDATE user_settings
			 SET latitude = ?, longitude = ?, calendar_ics_url = ?, timezone = ?, updated_at = ?
			 WHERE id = ? AND deleted_at IS NULL`,
			u.Latitude, u.Longitude, u.CalendarICSURL, u.Timezone, now, existing.ID,
		)
		if execErr != nil {
			return UserSettingsRow{}, fmt.Errorf("update user settings: %w", execErr)
		}
	}

	return r.Get(ctx, userID)
}

func (r *SQLiteUserSettingsRepository) ListAllCategories(ctx context.Context) ([]NewsCategoryTypeRow, error) {
	var rows []NewsCategoryTypeRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, label, sort_order FROM news_category_types ORDER BY sort_order`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all news categories: %w", err)
	}
	return rows, nil
}

func (r *SQLiteUserSettingsRepository) ListSelectedCategories(ctx context.Context, userID string) ([]NewsCategoryTypeRow, error) {
	var rows []NewsCategoryTypeRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT nct.id, nct.label, nct.sort_order
		 FROM news_category_types nct
		 INNER JOIN user_news_categories unc ON unc.category_id = nct.id
		 WHERE unc.user_id = ? AND unc.deleted_at IS NULL
		 ORDER BY nct.sort_order`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list selected news categories: %w", err)
	}
	return rows, nil
}

func (r *SQLiteUserSettingsRepository) SetSelectedCategories(ctx context.Context, userID string, categoryIDs []string) error {
	now := time.Now().UTC().Format(timeFormat)

	// Soft-delete all existing selections for this user.
	_, err := r.db.ExecContext(ctx,
		`UPDATE user_news_categories SET deleted_at = ?, updated_at = ? WHERE user_id = ? AND deleted_at IS NULL`,
		now, now, userID,
	)
	if err != nil {
		return fmt.Errorf("soft-delete user news categories: %w", err)
	}

	// Insert new selections.
	for _, categoryID := range categoryIDs {
		id, genErr := uuid.NewV7()
		if genErr != nil {
			return fmt.Errorf("generate user_news_category id: %w", genErr)
		}
		_, execErr := r.db.ExecContext(ctx,
			`INSERT INTO user_news_categories (id, user_id, category_id, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?)`,
			id.String(), userID, categoryID, now, now,
		)
		if execErr != nil {
			return fmt.Errorf("insert user news category: %w", execErr)
		}
	}
	return nil
}
