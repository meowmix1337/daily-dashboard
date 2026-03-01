package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/daily-dashboard/backend/internal/model"
)

// QuotesService fetches a daily quote from API Ninjas.
type QuotesService struct {
	httpClient *http.Client
	apiKey     string
	cache      *CacheService
}

// NewQuotesService creates a new QuotesService.
func NewQuotesService(httpClient *http.Client, apiKey string, cache *CacheService) *QuotesService {
	return &QuotesService{
		httpClient: httpClient,
		apiKey:     apiKey,
		cache:      cache,
	}
}

const quotesCacheTTL = 24 * time.Hour

// Fetch retrieves a random daily quote.
func (s *QuotesService) Fetch(ctx context.Context) (model.Quote, error) {
	const cacheKey = "quote"
	if v, ok := s.cache.Get(cacheKey); ok {
		return v.(model.Quote), nil
	}

	if s.apiKey == "" {
		return model.Quote{}, fmt.Errorf("API_NINJAS_API_KEY not configured")
	}

	quote, err := s.fetchFromAPI(ctx)
	if err != nil {
		return model.Quote{}, err
	}

	s.cache.Set(cacheKey, quote, quotesCacheTTL)
	return quote, nil
}

func (s *QuotesService) fetchFromAPI(ctx context.Context) (model.Quote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.api-ninjas.com/v2/quotes", nil)
	if err != nil {
		return model.Quote{}, err
	}
	req.Header.Set("X-Api-Key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return model.Quote{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.Quote{}, fmt.Errorf("API Ninjas quotes returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Quote{}, err
	}

	var quotes []apiNinjasQuote
	if err := json.Unmarshal(body, &quotes); err != nil {
		return model.Quote{}, err
	}

	if len(quotes) == 0 {
		return model.Quote{}, fmt.Errorf("API Ninjas quotes returned empty response")
	}

	return model.Quote{
		Text:   quotes[0].Quote,
		Author: quotes[0].Author,
	}, nil
}

type apiNinjasQuote struct {
	Quote    string `json:"quote"`
	Author   string `json:"author"`
	Category string `json:"category"`
}
