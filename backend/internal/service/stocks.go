package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/repository"
)

// ErrSymbolNotFound is returned when a watchlist symbol does not exist.
var ErrSymbolNotFound = errors.New("symbol not in watchlist")

// StocksService fetches stock quotes from Finnhub and crypto from CoinGecko.
type StocksService struct {
	httpClient *http.Client
	apiKey     string
	cache      *CacheService
	repo       repository.StocksWatchlistRepository
}

// NewStocksService creates a new StocksService backed by a watchlist repository.
func NewStocksService(httpClient *http.Client, apiKey string, cache *CacheService, repo repository.StocksWatchlistRepository) *StocksService {
	return &StocksService{
		httpClient: httpClient,
		apiKey:     apiKey,
		cache:      cache,
		repo:       repo,
	}
}

const stocksCacheTTL = 10 * time.Second

// Fetch retrieves stock quotes for the current watchlist.
func (s *StocksService) Fetch(ctx context.Context, userID string) ([]model.StockQuote, error) {
	cacheKey := "stocks:" + userID
	if v, ok := s.cache.Get(cacheKey); ok {
		return v.([]model.StockQuote), nil
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("FINNHUB_API_KEY not configured")
	}

	quotes, err := s.fetchFromAPIs(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, quotes, stocksCacheTTL)
	return quotes, nil
}

// GetSymbols returns the current watchlist symbols for a user.
func (s *StocksService) GetSymbols(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get symbols: %w", err)
	}
	symbols := make([]string, len(rows))
	for i, row := range rows {
		symbols[i] = row.Symbol
	}
	return symbols, nil
}

// AddSymbol adds a symbol to the user's watchlist (UPSERT — re-activates soft-deleted rows).
func (s *StocksService) AddSymbol(ctx context.Context, userID string, sym string) error {
	sym = strings.ToUpper(strings.TrimSpace(sym))
	if sym == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if err := s.repo.Add(ctx, userID, sym); err != nil {
		return err
	}
	s.cache.Delete("stocks:" + userID)
	return nil
}

// RemoveSymbol removes a symbol from the user's watchlist (soft-delete).
func (s *StocksService) RemoveSymbol(ctx context.Context, userID string, sym string) error {
	sym = strings.ToUpper(strings.TrimSpace(sym))

	// 1. Verify the symbol exists in the watchlist.
	if _, err := s.repo.Get(ctx, userID, sym); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSymbolNotFound
		}
		return fmt.Errorf("get symbol: %w", err)
	}

	// 2. Soft-delete it.
	if err := s.repo.Remove(ctx, userID, sym); err != nil {
		return fmt.Errorf("remove symbol: %w", err)
	}
	s.cache.Delete("stocks:" + userID)
	return nil
}

// SearchSymbols queries Finnhub for symbols matching the given query string.
func (s *StocksService) SearchSymbols(ctx context.Context, query string) ([]model.SymbolSearchResult, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("FINNHUB_API_KEY not configured")
	}

	u := fmt.Sprintf("https://finnhub.io/api/v1/search?q=%s",
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Finnhub-Token", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB max
	if err != nil {
		return nil, err
	}

	var result finnhubSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	max := len(result.Result)
	if max > 10 {
		max = 10
	}
	out := make([]model.SymbolSearchResult, 0, max)
	for _, item := range result.Result[:max] {
		out = append(out, model.SymbolSearchResult{
			Symbol:      item.Symbol,
			Description: item.Description,
			Type:        item.Type,
		})
	}
	return out, nil
}

func (s *StocksService) fetchFromAPIs(ctx context.Context, userID string) ([]model.StockQuote, error) {
	symbols, err := s.GetSymbols(ctx, userID)
	if err != nil {
		return nil, err
	}

	results := make([]model.StockQuote, len(symbols))
	g, gctx := errgroup.WithContext(ctx)

	for i, sym := range symbols {
		i, sym := i, sym
		g.Go(func() error {
			var q model.StockQuote
			var err error
			if sym == "BTC" {
				q, err = s.fetchBTC(gctx)
				if err != nil {
					return nil // BTC failure is non-fatal; slot stays zero value
				}
			} else {
				q, err = s.fetchFinnhub(gctx, sym)
				if err != nil {
					return err
				}
			}
			results[i] = q
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Filter out zero-value entries (e.g. BTC slot if fetch failed)
	filtered := make([]model.StockQuote, 0, len(results))
	for _, q := range results {
		if q.Symbol != "" {
			filtered = append(filtered, q)
		}
	}
	return filtered, nil
}

func (s *StocksService) fetchFinnhub(ctx context.Context, symbol string) (model.StockQuote, error) {
	u := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s", url.QueryEscape(symbol))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return model.StockQuote{}, err
	}
	req.Header.Set("X-Finnhub-Token", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return model.StockQuote{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB max
	if err != nil {
		return model.StockQuote{}, err
	}

	var q finnhubQuote
	if err := json.Unmarshal(body, &q); err != nil {
		return model.StockQuote{}, err
	}

	return model.StockQuote{
		Symbol: symbol,
		Price:  q.C,
		Change: q.D,
		Pct:    q.DP,
	}, nil
}

func (s *StocksService) fetchBTC(ctx context.Context) (model.StockQuote, error) {
	u := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd&include_24hr_change=true"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return model.StockQuote{}, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return model.StockQuote{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB max
	if err != nil {
		return model.StockQuote{}, err
	}

	var result map[string]map[string]float64
	if err := json.Unmarshal(body, &result); err != nil {
		return model.StockQuote{}, err
	}

	btc, ok := result["bitcoin"]
	if !ok {
		return model.StockQuote{}, fmt.Errorf("bitcoin not in response")
	}

	price := btc["usd"]
	pct := btc["usd_24h_change"]

	return model.StockQuote{
		Symbol: "BTC",
		Price:  price,
		Change: price * pct / 100,
		Pct:    pct,
	}, nil
}

// Finnhub API types
type finnhubQuote struct {
	C  float64 `json:"c"`  // current price
	D  float64 `json:"d"`  // change
	DP float64 `json:"dp"` // percent change
}

type finnhubSearchResponse struct {
	Count  int                 `json:"count"`
	Result []finnhubSearchItem `json:"result"`
}

type finnhubSearchItem struct {
	Description string `json:"description"`
	Symbol      string `json:"symbol"`
	Type        string `json:"type"`
}
