package main

import (
	"log"
	"net/http"

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

	// Setup routes
	mux := api.SetupRoutes(apiCfg)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
