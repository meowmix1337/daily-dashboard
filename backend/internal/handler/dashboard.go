package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/service"
)

// DashboardHandler aggregates all widget data into a single response.
type DashboardHandler struct {
	weather  *service.WeatherService
	stocks   *service.StocksService
	calendar *service.CalendarService
	tasks    *service.TasksService
	sunrise  *service.SunriseService
	quotes   *service.QuotesService
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(
	weather *service.WeatherService,
	stocks *service.StocksService,
	calendar *service.CalendarService,
	tasks *service.TasksService,
	sunrise *service.SunriseService,
	quotes *service.QuotesService,
) *DashboardHandler {
	return &DashboardHandler{
		weather:  weather,
		stocks:   stocks,
		calendar: calendar,
		tasks:    tasks,
		sunrise:  sunrise,
		quotes:   quotes,
	}
}

// Get fans out to all services concurrently and returns a unified response.
// If any individual service fails, that field is left nil/empty.
func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var resp model.DashboardResponse

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		data, err := h.weather.Fetch(gctx)
		if err != nil {
			slog.Warn("weather unavailable", "error", err)
			return nil
		}
		resp.Weather = &data
		return nil
	})

	g.Go(func() error {
		data, err := h.stocks.Fetch(gctx)
		if err != nil {
			slog.Warn("stocks unavailable", "error", err)
			return nil
		}
		resp.Stocks = data
		return nil
	})

	g.Go(func() error {
		data, err := h.calendar.Fetch(gctx)
		if err != nil {
			slog.Warn("calendar fetch failed", "error", err)
			return nil
		}
		resp.Calendar = data
		return nil
	})

	g.Go(func() error {
		data, err := h.tasks.List(gctx)
		if err != nil {
			slog.Warn("tasks fetch failed", "error", err)
			return nil
		}
		resp.Tasks = data
		return nil
	})

	g.Go(func() error {
		sunriseTime, sunsetTime, daylight, err := h.sunrise.Fetch(gctx)
		if err != nil {
			slog.Warn("sunrise unavailable", "error", err)
		}
		quote, err := h.quotes.Fetch(gctx)
		if err != nil {
			slog.Warn("quotes unavailable", "error", err)
		}
		meta := model.MetaData{
			Sunrise:  sunriseTime,
			Sunset:   sunsetTime,
			Daylight: daylight,
			Quote:    quote,
		}
		resp.Meta = &meta
		return nil
	})

	// errgroup.Wait() will not return a non-nil error since we always return nil
	_ = g.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
