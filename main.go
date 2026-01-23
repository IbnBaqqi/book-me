package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" //we had to import drivers but no actual direct use in code, Go :)
)

type apiConfig struct {
	db		*database.Queries
}


func main() {
	const port = "8080"

	// load .env file into environment variables
	// get DB_URL from the environment variable
	// open a connection to the database

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", dbURL)
	}

	//Using SQLC generated database package to create a new *database.Queries,
	// and storing in apiConfig struct so that handlers can access it:
	dbQueries := database.New(dbConn)

	apiCfg := &apiConfig{
		db: dbQueries,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", healthHandler)

	server := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}