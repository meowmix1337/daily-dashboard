package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/daily-dashboard/backend/internal/model"
)

// Sentinel errors for watchlist mutations.
var (
	ErrSymbolExists   = errors.New("symbol already in watchlist")
	ErrSymbolNotFound = errors.New("symbol not in watchlist")
)

// StocksService fetches stock quotes from Finnhub and crypto from CoinGecko.
type StocksService struct {
	httpClient *http.Client
	apiKey     string
	cache      *CacheService
	mu         sync.RWMutex
	symbols    []string
}

// NewStocksService creates a new StocksService with the default watchlist.
func NewStocksService(httpClient *http.Client, apiKey string, cache *CacheService) *StocksService {
	return &StocksService{
		httpClient: httpClient,
		apiKey:     apiKey,
		cache:      cache,
		symbols:    []string{"AAPL", "GOOGL", "MSFT", "BTC"},
	}
}

const stocksCacheTTL = 10 * time.Second

// Fetch retrieves stock quotes for the current watchlist.
func (s *StocksService) Fetch(ctx context.Context) ([]model.StockQuote, error) {
	const cacheKey = "stocks"
	if v, ok := s.cache.Get(cacheKey); ok {
		return v.([]model.StockQuote), nil
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("FINNHUB_API_KEY not configured")
	}

	quotes, err := s.fetchFromAPIs(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, quotes, stocksCacheTTL)
	return quotes, nil
}

// GetSymbols returns a copy of the current watchlist symbols.
func (s *StocksService) GetSymbols() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.symbols))
	copy(out, s.symbols)
	return out
}

// AddSymbol appends a symbol to the watchlist if not already present.
func (s *StocksService) AddSymbol(sym string) error {
	sym = strings.ToUpper(strings.TrimSpace(sym))
	if sym == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.symbols {
		if existing == sym {
			return ErrSymbolExists
		}
	}
	s.symbols = append(s.symbols, sym)
	s.cache.Delete("stocks")
	return nil
}

// RemoveSymbol removes a symbol from the watchlist.
func (s *StocksService) RemoveSymbol(sym string) error {
	sym = strings.ToUpper(strings.TrimSpace(sym))
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.symbols {
		if existing == sym {
			s.symbols = append(s.symbols[:i], s.symbols[i+1:]...)
			s.cache.Delete("stocks")
			return nil
		}
	}
	return ErrSymbolNotFound
}

// SearchSymbols queries Finnhub for symbols matching the given query string.
func (s *StocksService) SearchSymbols(ctx context.Context, query string) ([]model.SymbolSearchResult, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("FINNHUB_API_KEY not configured")
	}

	u := fmt.Sprintf("https://finnhub.io/api/v1/search?q=%s&token=%s",
		url.QueryEscape(query), s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

func (s *StocksService) fetchFromAPIs(ctx context.Context) ([]model.StockQuote, error) {
	s.mu.RLock()
	syms := make([]string, len(s.symbols))
	copy(syms, s.symbols)
	s.mu.RUnlock()

	results := make([]model.StockQuote, len(syms))
	g, gctx := errgroup.WithContext(ctx)

	for i, sym := range syms {
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
	u := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s&token=%s", symbol, s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return model.StockQuote{}, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return model.StockQuote{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

	body, err := io.ReadAll(resp.Body)
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
