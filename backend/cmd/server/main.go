package main

import (
	"context"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daily-dashboard/backend/internal/config"
	"github.com/daily-dashboard/backend/internal/database"
	"github.com/daily-dashboard/backend/internal/server"
	"github.com/daily-dashboard/backend/internal/service"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Fail fast on missing or weak required configuration.
	missing := false
	for _, check := range []struct {
		val  string
		name string
	}{
		{cfg.GoogleClientID, "GOOGLE_CLIENT_ID"},
		{cfg.GoogleClientSecret, "GOOGLE_CLIENT_SECRET"},
		{cfg.GoogleCallbackURL, "GOOGLE_CALLBACK_URL"},
		{cfg.FrontendURL, "FRONTEND_URL"},
	} {
		if check.val == "" {
			slog.Error("missing required env var", "var", check.name)
			missing = true
		}
	}

	// Decode session secret from hex; must be at least 32 bytes (64 hex chars).
	sessionKey, err := hex.DecodeString(cfg.SessionSecret)
	if err != nil || len(sessionKey) < 32 {
		slog.Error("SESSION_SECRET must be a valid hex string of at least 64 chars (openssl rand -hex 32)")
		missing = true
	} else {
		cfg.SessionKey = sessionKey
	}

	// Validate FRONTEND_URL is an absolute http(s) URL to prevent open redirects.
	if cfg.FrontendURL != "" {
		if u, parseErr := url.Parse(cfg.FrontendURL); parseErr != nil ||
			(u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			slog.Error("FRONTEND_URL must be an absolute http(s) URL", "value", cfg.FrontendURL)
			missing = true
		}
	}

	// Decode optional encryption key from hex; must be exactly 32 bytes (64 hex chars) for AES-256.
	if cfg.EncryptionKey != "" {
		encKey, encErr := hex.DecodeString(cfg.EncryptionKey)
		if encErr != nil || len(encKey) != 32 {
			slog.Error("ENCRYPTION_KEY must be a valid hex string of exactly 64 chars (openssl rand -hex 32)")
			missing = true
		} else {
			cfg.EncryptionKeyBytes = encKey
			cfg.EncryptionKey = "" // clear hex string from memory
		}
	} else {
		slog.Warn("ENCRYPTION_KEY not set — calendar ICS URLs will be stored unencrypted")
	}

	if missing {
		os.Exit(1)
	}

	// Build EncryptionService from validated key bytes (before server wiring).
	var encSvc *service.EncryptionService
	if len(cfg.EncryptionKeyBytes) > 0 {
		var encErr error
		encSvc, encErr = service.NewEncryptionService(cfg.EncryptionKeyBytes)
		if encErr != nil {
			slog.Error("failed to create encryption service", "error", encErr)
			os.Exit(1)
		}
	}

	db, err := database.Open(cfg.SQLitePath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrateCancel()
	if err := database.Migrate(migrateCtx, db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg, db, encSvc)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exited")
}
