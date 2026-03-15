package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/daily-dashboard/backend/internal/model"
)

func TestQuotesService_NoAPIKey(t *testing.T) {
	cache := &CacheService{}
	svc := NewQuotesService(&http.Client{}, "", cache)

	_, err := svc.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error when API key is empty")
	}
}

func TestQuotesService_CacheHit(t *testing.T) {
	cache := &CacheService{}
	cache.Set("quote", model.Quote{Text: "cached quote", Author: "cached author"}, time.Minute)

	svc := NewQuotesService(&http.Client{}, "key", cache)
	got, err := svc.Fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "cached quote" {
		t.Errorf("got %q, want %q", got.Text, "cached quote")
	}
}

func newTestQuotesService(srv *httptest.Server, apiKey string) *QuotesService {
	cache := &CacheService{}
	parsed, _ := url.Parse(srv.URL)
	transport := &hostRewriteTransport{
		rt:   srv.Client().Transport,
		host: parsed.Host,
	}
	return &QuotesService{
		httpClient: &http.Client{Transport: transport},
		apiKey:     apiKey,
		cache:      cache,
	}
}

func TestQuotesService_FetchFromAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode([]apiNinjasQuote{
			{Quote: "Test quote", Author: "Test Author"},
		})
	}))
	defer srv.Close()

	svc := newTestQuotesService(srv, "test-key")
	got, err := svc.fetchFromAPI(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "Test quote" {
		t.Errorf("got text %q, want %q", got.Text, "Test quote")
	}
	if got.Author != "Test Author" {
		t.Errorf("got author %q, want %q", got.Author, "Test Author")
	}
}

func TestQuotesService_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	svc := newTestQuotesService(srv, "key")
	_, err := svc.fetchFromAPI(context.Background())
	if err == nil {
		t.Fatal("expected error on non-200 response")
	}
}

func TestQuotesService_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]apiNinjasQuote{})
	}))
	defer srv.Close()

	svc := newTestQuotesService(srv, "key")
	_, err := svc.fetchFromAPI(context.Background())
	if err == nil {
		t.Fatal("expected error on empty quotes response")
	}
}

// hostRewriteTransport rewrites all outbound requests to target a specific host.
// This lets tests point service code at an httptest.Server without modifying production URLs.
type hostRewriteTransport struct {
	rt   http.RoundTripper
	host string
}

func (t *hostRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	r2.URL.Scheme = "http"
	r2.URL.Host = t.host
	return t.rt.RoundTrip(r2)
}
