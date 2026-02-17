// Package main is the entry point for the book-me server.
package main

import (
	"context"
	"fmt"

	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IbnBaqqi/book-me/internal/api"
	"github.com/IbnBaqqi/book-me/internal/config"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize & Setup global structured logging
	logger.Init(&cfg.Logger, cfg.App.Env)

	slog.Info("starting book-me server",
		"port", cfg.Server.Port,
		"log_level", cfg.Logger.Level,
	)

	// Initialize database
	ctx := context.Background()
	db, err := database.Connect(ctx, &cfg.App)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	// Initialize & setup rservices
	apiCfg, err := api.New(cfg, db)
	if err != nil {
		return fmt.Errorf("failed to initialize api services: %w", err)
	}

	// Setup routes
	mux := api.SetupRoutes(apiCfg)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Channel to catch server errors
	errCh := make(chan error, 1)
	go func() {
		slog.Info("Server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Setup signal catching
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Wait for either a signal or a server error
	select {
	case <-quit:
		slog.Info("Server is shutting down...")
	case err := <-errCh:
		return fmt.Errorf("server failed to start: %w", err)
	}

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("Server exited gracefully")
	return nil
}
