package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/daily-dashboard/backend/internal/service"
)

type CalendarHandler struct {
	service *service.CalendarService
}

func NewCalendarHandler(svc *service.CalendarService) *CalendarHandler {
	return &CalendarHandler{service: svc}
}

func (h *CalendarHandler) AddRoutes(r chi.Router) {
	r.Get("/api/calendar", h.Get)
}

func (h *CalendarHandler) Get(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.Fetch(r.Context())
	if err != nil {
		slog.Error("calendar fetch error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
