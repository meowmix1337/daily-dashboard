package handler

import (
	"encoding/json"
	"net/http"

	"github.com/daily-dashboard/backend/internal/middleware"
)

// MeHandler serves GET /api/auth/me — returns the current user from the session cookie.
type MeHandler struct{}

func NewMeHandler() *MeHandler { return &MeHandler{} }

func (h *MeHandler) Get(w http.ResponseWriter, r *http.Request) {
	sess, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": sess.UserID,
		"email":   sess.Email,
		"name":    sess.Name,
	})
}
