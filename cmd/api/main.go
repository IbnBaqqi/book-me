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
)

func main() {

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup structured logging
	logger := cfg.Logger.New()
	slog.SetDefault(logger)

	logger.Info("starting book-me server",
		"port", cfg.Server.Port,
		"log_level", cfg.Logger.Level,
	)

	// Initialize database & services
	ctx := context.Background()
	apiCfg, err := api.New(ctx, cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize api services:", "error", err)
	}

	// Ensure database connection close on exit
	defer func() {
		if err = apiCfg.Close(); err != nil {
			logger.Error("failed to close database connection", "error", err)
		}
	}()

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
		logger.Info("Server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start:", "error", err)
			os.Exit(1)
		}
	}()

	// Setup signal catching
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Server is shutting down...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", "error", err)
	}

	logger.Info("Server exited gracefully")

}
