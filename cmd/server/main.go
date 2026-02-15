package main

import (
	"context"

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

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
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
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	// Ensure database connection close on exit
	defer func() {
		if err = db.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	// Initialize & setup rservices
	apiCfg, err := api.New(cfg, db)
	if err != nil {
		slog.Error("Failed to initialize api services", "error", err)
		os.Exit(1)
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

	// Start server in a goroutine
	go func() {
		slog.Info("Server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Setup signal catching
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Server is shutting down...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited gracefully")

}
