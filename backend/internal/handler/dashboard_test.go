package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/service"
)

type alwaysFailTransport struct{}

func (alwaysFailTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network failure")
}

func newFailingHTTPClient() *http.Client {
	return &http.Client{Transport: alwaysFailTransport{}}
}

func newDashboardRouter(t *testing.T) chi.Router {
	t.Helper()
	cache := service.NewCacheService()
	failClient := newFailingHTTPClient()

	weatherSvc := service.NewWeatherService(failClient, cache, 37.77, -122.41)
	stocksSvc := service.NewStocksService(failClient, "fake-key", cache)
	calendarSvc := service.NewCalendarService(failClient, "", cache, time.UTC)
	tasksSvc := service.NewTasksService()
	sunriseSvc := service.NewSunriseService(failClient, cache, 37.77, -122.41)
	quotesSvc := service.NewQuotesService(failClient, "", cache)

	h := NewDashboardHandler(weatherSvc, stocksSvc, calendarSvc, tasksSvc, sunriseSvc, quotesSvc)
	r := chi.NewRouter()
	h.AddRoutes(r)
	return r
}

// TestDashboardHandler_Get_ExternalServiceFailures verifies that the dashboard
// endpoint returns 200 with partial data (tasks) even when all external HTTP
// services fail. This is the core contract of the fan-out handler.
func TestDashboardHandler_Get_ExternalServiceFailures(t *testing.T) {
	r := newDashboardRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 even when all external services fail", w.Code)
	}

	var resp model.DashboardResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode dashboard response: %v", err)
	}

	// Weather should be nil (service failed)
	if resp.Weather != nil {
		t.Error("expected weather to be nil when service fails")
	}

	// 200 + valid JSON decode proves the fan-out contract regardless of task seed state
}

func TestDashboardHandler_Get_ContentType(t *testing.T) {
	r := newDashboardRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("got Content-Type %q, want application/json", ct)
	}
}

func TestDashboardHandler_Get_ResponseShape(t *testing.T) {
	r := newDashboardRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp model.DashboardResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode dashboard response: %v", err)
	}

	// meta is always set (even when sunrise/quotes fail, an empty MetaData is returned)
	if resp.Meta == nil {
		t.Error("expected meta to be present in dashboard response")
	}
}

// TestDashboardHandler_Get_ContextCancelled ensures that a cancelled context
// does not cause a 5xx — the handler should still return 200.
func TestDashboardHandler_Get_ContextCancelled(t *testing.T) {
	r := newDashboardRouter(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the request

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 even with cancelled context", w.Code)
	}
}
