package repository

import "context"

// StocksWatchlistRepository defines the data-access contract for the stocks watchlist.
type StocksWatchlistRepository interface {
	// ListSymbols returns all active symbol strings for the given user.
	ListSymbols(ctx context.Context, userID string) ([]string, error)
	// Exists checks whether the given symbol is in the user's active watchlist.
	Exists(ctx context.Context, userID string, symbol string) (bool, error)
	// Add inserts or re-activates a symbol for the given user.
	Add(ctx context.Context, userID string, symbol string) error
	// Remove soft-deletes a symbol for the given user.
	Remove(ctx context.Context, userID string, symbol string) error
}
