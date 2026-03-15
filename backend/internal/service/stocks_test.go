package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/daily-dashboard/backend/internal/model"
)

func newTestStocksService(srv *httptest.Server, apiKey string) *StocksService {
	parsed, _ := url.Parse(srv.URL)
	transport := &hostRewriteTransport{
		rt:   srv.Client().Transport,
		host: parsed.Host,
	}
	return &StocksService{
		httpClient: &http.Client{Transport: transport},
		apiKey:     apiKey,
		cache:      &CacheService{},
		symbols:    []string{"AAPL"},
	}
}

func TestStocksService_NoAPIKey(t *testing.T) {
	svc := NewStocksService(&http.Client{}, "", &CacheService{})
	_, err := svc.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error when API key is empty")
	}
}

func TestStocksService_CacheHit(t *testing.T) {
	cache := &CacheService{}
	cached := []model.StockQuote{{Symbol: "AAPL", Price: 100}}
	cache.Set("stocks", cached, time.Minute)

	svc := NewStocksService(&http.Client{}, "key", cache)
	got, err := svc.Fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Symbol != "AAPL" {
		t.Errorf("unexpected cached result: %+v", got)
	}
}

func TestStocksService_AddSymbol(t *testing.T) {
	svc := NewStocksService(&http.Client{}, "", &CacheService{})
	initial := len(svc.GetSymbols())

	if err := svc.AddSymbol("tsla"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	syms := svc.GetSymbols()
	if len(syms) != initial+1 {
		t.Errorf("expected %d symbols, got %d", initial+1, len(syms))
	}
	// symbol should be uppercased
	found := false
	for _, s := range syms {
		if s == "TSLA" {
			found = true
		}
	}
	if !found {
		t.Error("expected TSLA in watchlist")
	}
}

func TestStocksService_AddSymbol_Duplicate(t *testing.T) {
	svc := NewStocksService(&http.Client{}, "", &CacheService{})
	if err := svc.AddSymbol("AAPL"); err != ErrSymbolExists {
		t.Fatalf("expected ErrSymbolExists, got %v", err)
	}
}

func TestStocksService_AddSymbol_Empty(t *testing.T) {
	svc := NewStocksService(&http.Client{}, "", &CacheService{})
	if err := svc.AddSymbol(""); err == nil {
		t.Fatal("expected error for empty symbol")
	}
}

func TestStocksService_RemoveSymbol(t *testing.T) {
	svc := NewStocksService(&http.Client{}, "", &CacheService{})
	initial := len(svc.GetSymbols())

	if err := svc.RemoveSymbol("AAPL"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(svc.GetSymbols()) != initial-1 {
		t.Errorf("expected %d symbols after remove", initial-1)
	}
}

func TestStocksService_RemoveSymbol_NotFound(t *testing.T) {
	svc := NewStocksService(&http.Client{}, "", &CacheService{})
	if err := svc.RemoveSymbol("ZZZZ"); err != ErrSymbolNotFound {
		t.Fatalf("expected ErrSymbolNotFound, got %v", err)
	}
}

func TestStocksService_FetchFinnhub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(finnhubQuote{C: 150.5, D: 1.2, DP: 0.8})
	}))
	defer srv.Close()

	svc := newTestStocksService(srv, "test-key")
	got, err := svc.fetchFinnhub(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Symbol != "AAPL" {
		t.Errorf("got symbol %q, want AAPL", got.Symbol)
	}
	if got.Price != 150.5 {
		t.Errorf("got price %f, want 150.5", got.Price)
	}
	if got.Change != 1.2 {
		t.Errorf("got change %f, want 1.2", got.Change)
	}
	if got.Pct != 0.8 {
		t.Errorf("got pct %f, want 0.8", got.Pct)
	}
}

func TestStocksService_FetchBTC(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]map[string]float64{
			"bitcoin": {"usd": 50000, "usd_24h_change": 2.5},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	svc := newTestStocksService(srv, "key")
	got, err := svc.fetchBTC(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Symbol != "BTC" {
		t.Errorf("got symbol %q, want BTC", got.Symbol)
	}
	if got.Price != 50000 {
		t.Errorf("got price %f, want 50000", got.Price)
	}
	// change = price * pct / 100 = 50000 * 2.5 / 100 = 1250
	if got.Change != 1250.0 {
		t.Errorf("got change %f, want 1250.0", got.Change)
	}
	if got.Pct != 2.5 {
		t.Errorf("got pct %f, want 2.5", got.Pct)
	}
}

func TestStocksService_SearchSymbols_NoAPIKey(t *testing.T) {
	svc := NewStocksService(&http.Client{}, "", &CacheService{})
	_, err := svc.SearchSymbols(context.Background(), "AAPL")
	if err == nil {
		t.Fatal("expected error when API key is empty")
	}
}

func TestStocksService_SearchSymbols(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(finnhubSearchResponse{
			Count: 2,
			Result: []finnhubSearchItem{
				{Symbol: "AAPL", Description: "Apple Inc", Type: "Common Stock"},
				{Symbol: "AAPLX", Description: "Apple Extended Fund", Type: "Common Stock"},
			},
		})
	}))
	defer srv.Close()

	svc := newTestStocksService(srv, "test-key")
	results, err := svc.SearchSymbols(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Symbol != "AAPL" {
		t.Errorf("got symbol %q, want AAPL", results[0].Symbol)
	}
	if results[0].Description != "Apple Inc" {
		t.Errorf("got description %q, want Apple Inc", results[0].Description)
	}
}

func TestStocksService_SearchSymbols_Truncation(t *testing.T) {
	items := make([]finnhubSearchItem, 15)
	for i := range items {
		items[i] = finnhubSearchItem{Symbol: fmt.Sprintf("SYM%d", i), Description: "test"}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(finnhubSearchResponse{Count: 15, Result: items})
	}))
	defer srv.Close()

	svc := newTestStocksService(srv, "test-key")
	results, err := svc.SearchSymbols(context.Background(), "SYM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 10 {
		t.Errorf("got %d results, want 10 (truncated to max)", len(results))
	}
}

func TestStocksService_GetSymbols_ReturnsCopy(t *testing.T) {
	svc := NewStocksService(&http.Client{}, "", &CacheService{})
	syms := svc.GetSymbols()
	syms[0] = "MUTATED"

	// Original should be unaffected
	original := svc.GetSymbols()
	if original[0] == "MUTATED" {
		t.Error("GetSymbols should return a copy, not a reference")
	}
}
