package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/daily-dashboard/backend/internal/config"
	"github.com/daily-dashboard/backend/internal/handler"
	"github.com/daily-dashboard/backend/internal/middleware"
	"github.com/daily-dashboard/backend/internal/service"
)

// Server holds the HTTP router and all dependencies.
type Server struct {
	router *chi.Mux
	cfg    *config.Config
}

// New creates a new Server with all services, handlers, and routes registered.
func New(cfg *config.Config) *Server {
	s := &Server{
		router: chi.NewRouter(),
		cfg:    cfg,
	}
	s.setupRoutes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) setupRoutes() {
	r := s.router

	// Global middleware
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS)
	r.Use(middleware.Logging)

	// Shared dependencies
	httpClient := &http.Client{Timeout: 30 * time.Second}
	cache := service.NewCacheService()

	// Services
	weatherSvc := service.NewWeatherService(httpClient, cache, s.cfg.Latitude, s.cfg.Longitude)
	newsSvc := service.NewNewsService(httpClient, s.cfg.GNewsAPIKey, cache)
	stocksSvc := service.NewStocksService(httpClient, s.cfg.FinnhubAPIKey, cache)
	calendarSvc := service.NewCalendarService(httpClient, s.cfg.ICSCalendarURL, cache, s.cfg.Timezone)
	tasksSvc := service.NewTasksService()
	sunriseSvc := service.NewSunriseService(httpClient, cache, s.cfg.Latitude, s.cfg.Longitude)
	quotesSvc := service.NewQuotesService(httpClient, s.cfg.APINinjasAPIKey, cache)

	// Handlers
	weatherH := handler.NewWeatherHandler(weatherSvc)
	newsH := handler.NewNewsHandler(newsSvc)
	stocksH := handler.NewStocksHandler(stocksSvc)
	calendarH := handler.NewCalendarHandler(calendarSvc)
	tasksH := handler.NewTasksHandler(tasksSvc)
	metaH := handler.NewMetaHandler(sunriseSvc, quotesSvc)
	dashboardH := handler.NewDashboardHandler(
		weatherSvc,
		stocksSvc,
		calendarSvc,
		tasksSvc,
		sunriseSvc,
		quotesSvc,
	)

	// Routes
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Get("/api/dashboard", dashboardH.Get)
	r.Get("/api/weather", weatherH.Get)
	r.Get("/api/news", newsH.Get)
	r.Get("/api/stocks", stocksH.Get)
	r.Get("/api/stocks/watchlist", stocksH.GetWatchlist)
	r.Post("/api/stocks/watchlist", stocksH.AddSymbol)
	r.Delete("/api/stocks/watchlist/{symbol}", stocksH.RemoveSymbol)
	r.Get("/api/stocks/search", stocksH.SearchSymbols)
	r.Get("/api/calendar", calendarH.Get)
	r.Get("/api/meta", metaH.Get)

	r.Get("/api/tasks", tasksH.List)
	r.Post("/api/tasks", tasksH.Create)
	r.Patch("/api/tasks/{id}", tasksH.Update)
	r.Delete("/api/tasks/{id}", tasksH.Delete)
}
