package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// SQLiteStocksWatchlistRepository implements StocksWatchlistRepository backed by SQLite.
type SQLiteStocksWatchlistRepository struct {
	db *sqlx.DB
}

// NewSQLiteStocksWatchlistRepository creates a new SQLiteStocksWatchlistRepository.
func NewSQLiteStocksWatchlistRepository(db *sqlx.DB) *SQLiteStocksWatchlistRepository {
	return &SQLiteStocksWatchlistRepository{db: db}
}

func (r *SQLiteStocksWatchlistRepository) List(ctx context.Context, userID string) ([]WatchlistRow, error) {
	var rows []WatchlistRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, user_id, symbol, created_at, updated_at
		 FROM stocks_watchlist
		 WHERE deleted_at IS NULL AND user_id = ?
		 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list watchlist: %w", err)
	}
	return rows, nil
}

func (r *SQLiteStocksWatchlistRepository) Add(ctx context.Context, userID string, symbol string) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate uuid: %w", err)
	}
	now := time.Now().UTC().Format(timeFormat)

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO stocks_watchlist (id, user_id, symbol, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(user_id, symbol) DO UPDATE
		   SET deleted_at = NULL,
		       updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')`,
		id.String(), userID, symbol, now, now,
	)
	if err != nil {
		return fmt.Errorf("add to watchlist: %w", err)
	}
	return nil
}

func (r *SQLiteStocksWatchlistRepository) Remove(ctx context.Context, userID string, symbol string) (int64, error) {
	now := time.Now().UTC().Format(timeFormat)
	result, err := r.db.ExecContext(ctx,
		`UPDATE stocks_watchlist
		 SET deleted_at = ?, updated_at = ?
		 WHERE user_id = ? AND symbol = ? AND deleted_at IS NULL`,
		now, now, userID, symbol,
	)
	if err != nil {
		return 0, fmt.Errorf("remove from watchlist: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("remove from watchlist rows affected: %w", err)
	}
	return rows, nil
}
