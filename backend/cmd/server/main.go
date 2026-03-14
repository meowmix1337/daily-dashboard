package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daily-dashboard/backend/internal/config"
	"github.com/daily-dashboard/backend/internal/database"
	"github.com/daily-dashboard/backend/internal/server"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Fail fast on missing or weak required configuration.
	// All three must be present together — a partial OAuth config is always wrong.
	missing := false
	if cfg.GoogleClientID == "" {
		slog.Error("missing required env var", "var", "GOOGLE_CLIENT_ID")
		missing = true
	}
	if cfg.GoogleClientSecret == "" {
		slog.Error("missing required env var", "var", "GOOGLE_CLIENT_SECRET")
		missing = true
	}
	if len(cfg.SessionSecret) < 32 {
		slog.Error("SESSION_SECRET must be at least 32 bytes (generate with: openssl rand -hex 32)")
		missing = true
	}
	if missing {
		os.Exit(1)
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

	srv := server.New(cfg, db)

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
