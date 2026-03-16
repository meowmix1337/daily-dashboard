package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"github.com/daily-dashboard/backend/internal/service"
)

type StocksHandler struct {
	service *service.StocksService
}

func NewStocksHandler(svc *service.StocksService) *StocksHandler {
	return &StocksHandler{service: svc}
}

func (h *StocksHandler) AddRoutes(r chi.Router) {
	r.Get("/api/stocks", h.Get)
	r.Get("/api/stocks/watchlist", h.GetWatchlist)
	r.With(httprate.LimitByIP(10, time.Second)).Post("/api/stocks/watchlist", h.AddSymbol)
	r.With(httprate.LimitByIP(10, time.Second)).Delete("/api/stocks/watchlist/{symbol}", h.RemoveSymbol)
	r.With(httprate.LimitByIP(2, time.Second)).Get("/api/stocks/search", h.SearchSymbols)
}

// Get returns quotes for the current watchlist.
func (h *StocksHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}
	data, err := h.service.Fetch(r.Context(), userID)
	if err != nil {
		slog.Error("stocks fetch error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetWatchlist returns the current list of watchlist symbols.
func (h *StocksHandler) GetWatchlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}
	syms, err := h.service.GetSymbols(r.Context(), userID)
	if err != nil {
		slog.Error("stocks get watchlist error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WatchlistResponse{Symbols: syms})
}

// AddSymbol adds a symbol to the watchlist.
// Body: {"symbol":"TSLA"}
// Returns 201 with updated list. Idempotent — re-adding an existing symbol succeeds silently.
func (h *StocksHandler) AddSymbol(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096) // 4 KB is generous for these small JSON bodies
	var body AddSymbolRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.AddSymbol(r.Context(), userID, body.Symbol); err != nil {
		slog.Error("stocks add error", "error", err)
		http.Error(w, "invalid symbol", http.StatusBadRequest)
		return
	}

	syms, err := h.service.GetSymbols(r.Context(), userID)
	if err != nil {
		slog.Error("stocks get watchlist error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(WatchlistResponse{Symbols: syms})
}

// RemoveSymbol removes a symbol from the watchlist.
// Returns 200 with updated list, or 404 if not found.
func (h *StocksHandler) RemoveSymbol(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}
	sym := chi.URLParam(r, "symbol")
	if err := h.service.RemoveSymbol(r.Context(), userID, sym); err != nil {
		if errors.Is(err, service.ErrSymbolNotFound) {
			http.Error(w, "symbol not found", http.StatusNotFound)
		} else {
			slog.Error("stocks remove error", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	syms, err := h.service.GetSymbols(r.Context(), userID)
	if err != nil {
		slog.Error("stocks get watchlist error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WatchlistResponse{Symbols: syms})
}

// SearchSymbols proxies Finnhub symbol search to keep the API key server-side.
// Query param: q (required)
func (h *StocksHandler) SearchSymbols(w http.ResponseWriter, r *http.Request) {
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
		slog.Error("stocks search error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	dtos := make([]SymbolSearchResultDTO, len(results))
	for i, r := range results {
		dtos[i] = SymbolSearchResultDTO{
			Symbol:      r.Symbol,
			Description: r.Description,
			Type:        r.Type,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SearchResponse{Results: dtos})
}

