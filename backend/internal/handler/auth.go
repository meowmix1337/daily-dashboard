package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/daily-dashboard/backend/internal/service"
	"github.com/daily-dashboard/backend/internal/session"
)

const sessionMaxAge = int(7 * 24 * time.Hour / time.Second) // 7 days in seconds

// AuthHandler serves the Google OAuth login / callback / logout endpoints.
type AuthHandler struct {
	authSvc    *service.AuthService
	sessionKey []byte
	successURL string // where to redirect after a successful login
}

func NewAuthHandler(authSvc *service.AuthService, sessionKey []byte, successURL string) *AuthHandler {
	return &AuthHandler{
		authSvc:    authSvc,
		sessionKey: sessionKey,
		successURL: successURL,
	}
}

// Login redirects the browser to Google's consent screen.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := randomHex(16)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes — just long enough to complete the flow
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.authSvc.AuthCodeURL(state), http.StatusFound)
}

// Callback handles the redirect from Google after the user consents.
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	// --- CSRF state check ---
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}
	// Immediately expire the state cookie.
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: "", Path: "/", MaxAge: -1})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	googleUser, err := h.authSvc.ExchangeAndGetUser(r.Context(), code)
	if err != nil {
		slog.Error("oauth exchange failed", "error", err)
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	userID, err := h.authSvc.UpsertUser(r.Context(), googleUser)
	if err != nil {
		slog.Error("user upsert failed", "error", err)
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	encoded, err := session.Encode(h.sessionKey, session.Data{
		UserID:    userID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	if err != nil {
		slog.Error("session encode failed", "error", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Value:    encoded,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure is intentionally omitted here; set it at the reverse-proxy
		// layer (nginx) or add a config flag when TLS is enabled.
	})

	http.Redirect(w, r, h.successURL, http.StatusFound)
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   session.CookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
