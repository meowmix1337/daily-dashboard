package repository

import (
	"context"
	"errors"
)

// ErrSymbolNotFound is returned when a watchlist symbol does not exist.
var ErrSymbolNotFound = errors.New("symbol not found in watchlist")

// WatchlistRow represents a single row in the stocks_watchlist table.
type WatchlistRow struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	Symbol    string `db:"symbol"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

// StocksWatchlistRepository defines the data-access contract for the stocks watchlist.
type StocksWatchlistRepository interface {
	// List returns all active (non-deleted) symbols for the given user.
	List(ctx context.Context, userID string) ([]WatchlistRow, error)
	// Add inserts or re-activates a symbol for the given user (UPSERT).
	Add(ctx context.Context, userID string, symbol string) error
	// Remove soft-deletes a symbol for the given user.
	Remove(ctx context.Context, userID string, symbol string) (int64, error)
}
