package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/repository"
	"github.com/daily-dashboard/backend/internal/service"
)

// UserSettingsHandler handles reading and updating user-scoped settings.
type UserSettingsHandler struct {
	service  *service.UserSettingsService
	validate *validator.Validate
}

// NewUserSettingsHandler creates a new UserSettingsHandler.
func NewUserSettingsHandler(svc *service.UserSettingsService, v *validator.Validate) *UserSettingsHandler {
	return &UserSettingsHandler{service: svc, validate: v}
}

// AddRoutes registers user settings routes on the given router.
func (h *UserSettingsHandler) AddRoutes(r chi.Router) {
	r.Get("/api/settings", h.Get)
	r.With(httprate.LimitByIP(10, time.Second)).Put("/api/settings", h.Upsert)
	r.Get("/api/settings/news-categories", h.GetNewsCategories)
	r.With(httprate.LimitByIP(10, time.Second)).Put("/api/settings/news-categories", h.SetNewsCategories)
}

func (h *UserSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	settings, err := h.service.Get(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get user settings", "error", err, "user_id", userID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settingsToResponse(settings))
}

func (h *UserSettingsHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req UpsertSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	upsert := repository.UserSettingsUpsert{
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		CalendarICSURL: req.CalendarICSURL,
		Timezone:       req.Timezone,
	}

	settings, err := h.service.Upsert(r.Context(), userID, upsert)
	if err != nil {
		if errors.Is(err, service.ErrSettingsValidation) {
			http.Error(w, "invalid request body", http.StatusBadRequest)
		} else {
			slog.Error("failed to upsert user settings", "error", err, "user_id", userID)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settingsModelToResponse(settings))
}

func (h *UserSettingsHandler) GetNewsCategories(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	available, err := h.service.ListAllCategories(r.Context())
	if err != nil {
		slog.Error("failed to list all news categories", "error", err, "user_id", userID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	selected, err := h.service.ListSelectedCategories(r.Context(), userID)
	if err != nil {
		slog.Error("failed to list selected news categories", "error", err, "user_id", userID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := NewsCategoriesResponse{
		Available: categoriesToResponse(available),
		Selected:  categoriesToResponse(selected),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserSettingsHandler) SetNewsCategories(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req SetNewsCategoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.SetSelectedCategories(r.Context(), userID, req.CategoryIDs); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			http.Error(w, "invalid category id", http.StatusBadRequest)
		} else {
			slog.Error("failed to set news categories", "error", err, "user_id", userID)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func settingsToResponse(s *model.UserSettings) UserSettingsResponse {
	if s == nil {
		return UserSettingsResponse{}
	}
	return settingsModelToResponse(*s)
}

func settingsModelToResponse(s model.UserSettings) UserSettingsResponse {
	return UserSettingsResponse{
		Latitude:       s.Latitude,
		Longitude:      s.Longitude,
		CalendarICSURL: s.CalendarICSURL,
		Timezone:       s.Timezone,
	}
}

func categoriesToResponse(cats []model.NewsCategoryType) []NewsCategoryTypeResponse {
	resp := make([]NewsCategoryTypeResponse, 0, len(cats))
	for _, c := range cats {
		resp = append(resp, NewsCategoryTypeResponse{
			ID:        c.ID,
			Label:     c.Label,
			SortOrder: c.SortOrder,
		})
	}
	return resp
}
