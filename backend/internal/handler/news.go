package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/daily-dashboard/backend/internal/service"
)

type NewsHandler struct {
	service *service.NewsService
}

func NewNewsHandler(svc *service.NewsService) *NewsHandler {
	return &NewsHandler{service: svc}
}

func (h *NewsHandler) AddRoutes(r chi.Router) {
	r.Get("/api/news", h.Get)
}

func (h *NewsHandler) Get(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.Fetch(r.Context())
	if err != nil {
		slog.Error("news fetch error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
