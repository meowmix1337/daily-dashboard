package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/jmoiron/sqlx"

	"github.com/daily-dashboard/backend/internal/config"
	"github.com/daily-dashboard/backend/internal/handler"
	"github.com/daily-dashboard/backend/internal/middleware"
	"github.com/daily-dashboard/backend/internal/repository"
	"github.com/daily-dashboard/backend/internal/service"
)

// Server holds the HTTP router and all dependencies.
type Server struct {
	router *chi.Mux
	cfg    *config.Config
	db     *sqlx.DB // used by service constructors added in subsequent PRs
}

// New creates a new Server with all services, handlers, and routes registered.
func New(cfg *config.Config, db *sqlx.DB) *Server {
	s := &Server{
		router: chi.NewRouter(),
		cfg:    cfg,
		db:     db,
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
	watchlistRepo := repository.NewSQLiteStocksWatchlistRepository(s.db)
	stocksSvc := service.NewStocksService(httpClient, s.cfg.FinnhubAPIKey, cache, watchlistRepo)
	calendarSvc := service.NewCalendarService(httpClient, s.cfg.ICSCalendarURL, cache, s.cfg.Timezone)
	taskRepo := repository.NewSQLiteTaskRepository(s.db)
	tasksSvc := service.NewTasksService(taskRepo)
	sunriseSvc := service.NewSunriseService(httpClient, cache, s.cfg.Latitude, s.cfg.Longitude)
	quotesSvc := service.NewQuotesService(httpClient, s.cfg.APINinjasAPIKey, cache)

	// Auth
	authSvc := service.NewAuthService(s.db, s.cfg.GoogleClientID, s.cfg.GoogleClientSecret, s.cfg.GoogleCallbackURL)
	authH := handler.NewAuthHandler(authSvc, s.cfg.SessionKey, s.cfg.FrontendURL, s.cfg.SecureCookies)
	requireAuth := middleware.RequireAuth(s.cfg.SessionKey)
	meH := handler.NewMeHandler()

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

	// Public routes — no session required
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	authH.AddRoutes(r)

	// Protected routes — valid session cookie required
	r.Group(func(r chi.Router) {
		r.Use(requireAuth)

		meH.AddRoutes(r)
		dashboardH.AddRoutes(r)
		weatherH.AddRoutes(r)
		newsH.AddRoutes(r)
		stocksH.AddRoutes(r)
		calendarH.AddRoutes(r)
		metaH.AddRoutes(r)
		tasksH.AddRoutes(r)
	})
}
