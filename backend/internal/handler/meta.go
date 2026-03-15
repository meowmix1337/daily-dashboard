package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/service"
)

type MetaHandler struct {
	sunrise *service.SunriseService
	quotes  *service.QuotesService
}

func NewMetaHandler(sunrise *service.SunriseService, quotes *service.QuotesService) *MetaHandler {
	return &MetaHandler{sunrise: sunrise, quotes: quotes}
}

func (h *MetaHandler) AddRoutes(r chi.Router) {
	r.Get("/api/meta", h.Get)
}

func (h *MetaHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sunriseTime, sunsetTime, daylight, sunriseErr := h.sunrise.Fetch(ctx)
	quote, quoteErr := h.quotes.Fetch(ctx)

	if sunriseErr != nil && quoteErr != nil {
		http.Error(w, "meta services unavailable", http.StatusServiceUnavailable)
		return
	}

	meta := model.MetaData{
		Sunrise:  sunriseTime,
		Sunset:   sunsetTime,
		Daylight: daylight,
		Quote:    quote,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}
