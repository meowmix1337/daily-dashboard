package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/daily-dashboard/backend/internal/model"
)

func newTestWeatherService(srv *httptest.Server) *WeatherService {
	parsed, _ := url.Parse(srv.URL)
	transport := &hostRewriteTransport{
		rt:   srv.Client().Transport,
		host: parsed.Host,
	}
	return &WeatherService{
		httpClient: &http.Client{Transport: transport},
		cache:      &CacheService{},
		lat:        37.7749,
		lon:        -122.4194,
	}
}

func TestWeatherService_CacheHit(t *testing.T) {
	cache := &CacheService{}
	cached := model.WeatherData{Temp: 72.0, Condition: "Clear Sky"}
	cache.Set("weather", cached, time.Minute)

	svc := NewWeatherService(&http.Client{}, cache, 37.7749, -122.4194)
	got, err := svc.Fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Condition != "Clear Sky" {
		t.Errorf("got condition %q, want %q", got.Condition, "Clear Sky")
	}
}

func TestWeatherService_FetchFromAPI(t *testing.T) {
	// Build a minimal valid Open-Meteo response
	forecast := openMeteoForecast{}
	forecast.Current.Temperature2m = 68.0
	forecast.Current.RelativeHumidity2m = 65
	forecast.Current.WeatherCode = 0
	forecast.Current.WindSpeed10m = 10.0
	forecast.Daily.Temperature2mMax = []float64{75.0}
	forecast.Daily.Temperature2mMin = []float64{55.0}
	// Use a fixed far-future timestamp so the hourly slot is never filtered out
	// regardless of when the test runs (avoids flakiness at minute boundaries).
	forecast.Hourly.Time = []string{"2099-01-01T00:00"}
	forecast.Hourly.Temperature2m = []float64{70.0}
	forecast.Hourly.WeatherCode = []int{0}
	forecast.Hourly.UVIndex = []float64{3.0}

	aqiResp := openMeteoAQI{}
	aqiResp.Current.USAQI = 42

	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount.Add(1) == 1 {
			json.NewEncoder(w).Encode(forecast)
		} else {
			json.NewEncoder(w).Encode(aqiResp)
		}
	}))
	defer srv.Close()

	svc := newTestWeatherService(srv)
	got, err := svc.fetchFromAPI(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Temp != 68.0 {
		t.Errorf("got temp %f, want 68.0", got.Temp)
	}
	if got.Condition != "Clear Sky" {
		t.Errorf("got condition %q, want Clear Sky", got.Condition)
	}
	if got.High != 75.0 {
		t.Errorf("got high %f, want 75.0", got.High)
	}
	if got.Low != 55.0 {
		t.Errorf("got low %f, want 55.0", got.Low)
	}
	if got.AQI != 42 {
		t.Errorf("got AQI %d, want 42", got.AQI)
	}
	if got.AQILabel != "Good" {
		t.Errorf("got AQI label %q, want Good", got.AQILabel)
	}
	if len(got.Hourly) == 0 {
		t.Error("expected at least one hourly forecast entry")
	}
	if got.UVIndex != 3.0 {
		t.Errorf("got UV index %v, want 3.0", got.UVIndex)
	}
}

func TestWeatherService_FetchFromAPI_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Immediately close to simulate network error
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()

	svc := newTestWeatherService(srv)
	_, err := svc.fetchFromAPI(context.Background())
	if err == nil {
		t.Fatal("expected error on network failure")
	}
}

func TestAQICategory(t *testing.T) {
	cases := []struct {
		aqi      int
		expected string
	}{
		{0, "Good"},
		{50, "Good"},
		{51, "Moderate"},
		{100, "Moderate"},
		{101, "Unhealthy for Sensitive"},
		{150, "Unhealthy for Sensitive"},
		{151, "Unhealthy"},
		{200, "Unhealthy"},
		{201, "Very Unhealthy"},
		{300, "Very Unhealthy"},
		{301, "Hazardous"},
	}
	for _, tc := range cases {
		got := aqiCategory(tc.aqi)
		if got != tc.expected {
			t.Errorf("aqiCategory(%d) = %q, want %q", tc.aqi, got, tc.expected)
		}
	}
}

func TestWMOToCondition(t *testing.T) {
	cond, icon := wmoToCondition(0)
	if cond != "Clear Sky" {
		t.Errorf("got %q, want Clear Sky", cond)
	}
	if icon == "" {
		t.Error("expected non-empty icon")
	}

	// Unknown code
	cond2, icon2 := wmoToCondition(999)
	if cond2 != "Unknown" {
		t.Errorf("got %q, want Unknown", cond2)
	}
	if icon2 == "" {
		t.Error("expected fallback icon")
	}
}
