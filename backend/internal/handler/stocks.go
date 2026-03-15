package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/time/rate"

	"github.com/daily-dashboard/backend/internal/service"
)

var searchLimiter = rate.NewLimiter(rate.Every(time.Second), 2) // 2 req/s burst

type StocksHandler struct {
	service *service.StocksService
}

func NewStocksHandler(svc *service.StocksService) *StocksHandler {
	return &StocksHandler{service: svc}
}

func (h *StocksHandler) AddRoutes(r chi.Router) {
	r.Get("/api/stocks", h.Get)
	r.Get("/api/stocks/watchlist", h.GetWatchlist)
	r.Post("/api/stocks/watchlist", h.AddSymbol)
	r.Delete("/api/stocks/watchlist/{symbol}", h.RemoveSymbol)
	r.Get("/api/stocks/search", h.SearchSymbols)
}

// Get returns quotes for the current watchlist.
func (h *StocksHandler) Get(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.Fetch(r.Context())
	if err != nil {
		log.Printf("stocks fetch error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetWatchlist returns the current list of watchlist symbols.
func (h *StocksHandler) GetWatchlist(w http.ResponseWriter, r *http.Request) {
	syms := h.service.GetSymbols()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"symbols": syms})
}

// AddSymbol adds a symbol to the watchlist.
// Body: {"symbol":"TSLA"}
// Returns 201 with updated list, or 409 if already present.
func (h *StocksHandler) AddSymbol(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Symbol string `json:"symbol"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.AddSymbol(body.Symbol); err != nil {
		if errors.Is(err, service.ErrSymbolExists) {
			http.Error(w, "symbol already in watchlist", http.StatusConflict)
		} else {
			http.Error(w, "invalid symbol", http.StatusBadRequest)
		}
		return
	}

	syms := h.service.GetSymbols()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"symbols": syms})
}

// RemoveSymbol removes a symbol from the watchlist.
// Returns 200 with updated list, or 404 if not found.
func (h *StocksHandler) RemoveSymbol(w http.ResponseWriter, r *http.Request) {
	sym := chi.URLParam(r, "symbol")
	if err := h.service.RemoveSymbol(sym); err != nil {
		if errors.Is(err, service.ErrSymbolNotFound) {
			http.Error(w, "symbol not found", http.StatusNotFound)
		} else {
			log.Printf("stocks remove error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	syms := h.service.GetSymbols()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"symbols": syms})
}

// SearchSymbols proxies Finnhub symbol search to keep the API key server-side.
// Query param: q (required)
func (h *StocksHandler) SearchSymbols(w http.ResponseWriter, r *http.Request) {
	if !searchLimiter.Allow() {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "missing query parameter 'q'", http.StatusBadRequest)
		return
	}
	if len(q) > 50 {
		http.Error(w, "query too long", http.StatusBadRequest)
		return
	}

	results, err := h.service.SearchSymbols(r.Context(), q)
	if err != nil {
		log.Printf("stocks search error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"results": results})
}
