package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IbnBaqqi/book-me/internal/api"
	"github.com/IbnBaqqi/book-me/internal/config"
)


func main() {
	const port = "8080"

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize services
	apiCfg, err := api.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize API config: %v", err)
	}

	// defer func() {
	// 	if err = apiCfg.DB.Close(); err != nil {
	// 		logger.Error("failed to close database connection", "error", err)
	// 	}
	// }()

	// Setup routes
	mux := api.SetupRoutes(apiCfg)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port: %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup signal catching
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	if err := apiCfg.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server exited gracefully")

}
