package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/daily-dashboard/backend/internal/model"
)

// NewsService fetches top headlines from GNews.
type NewsService struct {
	httpClient *http.Client
	apiKey     string
	cache      *CacheService
}

// NewNewsService creates a new NewsService.
func NewNewsService(httpClient *http.Client, apiKey string, cache *CacheService) *NewsService {
	return &NewsService{
		httpClient: httpClient,
		apiKey:     apiKey,
		cache:      cache,
	}
}

const newsCacheTTL = 30 * time.Minute

// Fetch retrieves top news headlines.
func (s *NewsService) Fetch(ctx context.Context) ([]model.NewsItem, error) {
	const cacheKey = "news"
	if v, ok := s.cache.Get(cacheKey); ok {
		return v.([]model.NewsItem), nil
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("GNEWS_API_KEY not configured")
	}

	items, err := s.fetchFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, items, newsCacheTTL)
	return items, nil
}

func (s *NewsService) fetchFromAPI(ctx context.Context) ([]model.NewsItem, error) {
	url := fmt.Sprintf(
		"https://gnews.io/api/v4/top-headlines?category=general&country=us&lang=en&max=8&apikey=%s",
		s.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var gnewsResp gNewsResponse
	if err := json.Unmarshal(body, &gnewsResp); err != nil {
		return nil, err
	}

	items := make([]model.NewsItem, 0, len(gnewsResp.Articles))
	for _, a := range gnewsResp.Articles {
		pub, err := time.Parse(time.RFC3339, a.PublishedAt)
		if err != nil {
			pub, err = time.Parse(time.RFC3339Nano, a.PublishedAt)
			if err != nil {
				slog.Warn("news: failed to parse publishedAt", "raw", a.PublishedAt, "error", err)
			}
		}
		items = append(items, model.NewsItem{
			Title:  a.Title,
			Source: a.Source.Name,
			Time:   relativeTime(pub),
			URL:    a.URL,
		})
	}
	return items, nil
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "recently"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// GNews API response types
type gNewsResponse struct {
	Articles []struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		PublishedAt string `json:"publishedAt"`
		Source      struct {
			Name string `json:"name"`
		} `json:"source"`
	} `json:"articles"`
}
