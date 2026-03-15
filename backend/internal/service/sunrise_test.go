package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func newTestSunriseService(srv *httptest.Server) *SunriseService {
	parsed, _ := url.Parse(srv.URL)
	transport := &hostRewriteTransport{
		rt:   srv.Client().Transport,
		host: parsed.Host,
	}
	return &SunriseService{
		httpClient: &http.Client{Transport: transport},
		cache:      &CacheService{},
		lat:        37.7749,
		lon:        -122.4194,
	}
}

func TestSunriseService_CacheHit(t *testing.T) {
	cache := &CacheService{}
	cache.Set("sunrise", sunriseResult{
		Sunrise:  "6:00 AM",
		Sunset:   "7:00 PM",
		Daylight: "13h 0m",
	}, time.Minute)

	svc := NewSunriseService(&http.Client{}, cache, 37.7749, -122.4194)
	rise, set, daylight, err := svc.Fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rise != "6:00 AM" {
		t.Errorf("got sunrise %q, want %q", rise, "6:00 AM")
	}
	if set != "7:00 PM" {
		t.Errorf("got sunset %q, want %q", set, "7:00 PM")
	}
	if daylight != "13h 0m" {
		t.Errorf("got daylight %q, want %q", daylight, "13h 0m")
	}
}

func TestSunriseService_FetchFromAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := sunriseSunsetResponse{
			Status: "OK",
		}
		resp.Results.Sunrise = "2026-03-15T13:00:00+00:00"
		resp.Results.Sunset = "2026-03-16T02:00:00+00:00"
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	svc := newTestSunriseService(srv)
	rise, set, daylight, err := svc.fetchFromAPI(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rise == "" {
		t.Error("expected non-empty sunrise")
	}
	if set == "" {
		t.Error("expected non-empty sunset")
	}
	if daylight == "" {
		t.Error("expected non-empty daylight")
	}
}

func TestSunriseService_APIStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := sunriseSunsetResponse{Status: "INVALID_REQUEST"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	svc := newTestSunriseService(srv)
	_, _, _, err := svc.fetchFromAPI(context.Background())
	if err == nil {
		t.Fatal("expected error when API status is not OK")
	}
}
