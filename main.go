package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
)

type apiConfig struct {
	db					*database.Queries //
	sessionStore		*sessions.CookieStore
	oauthConfig			*oauth2.Config
	auth 				*auth.Service
	redirectTokenURI	string
	user42InfoURL		string
	jwtSecret			string
}

func main() {
	const port = "8080"


	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system env vars")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET must be set")
	}

    clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		log.Fatal("CLIENT_ID must be set")
	}

	clientSecret := os.Getenv("SECRET")
	if clientSecret == "" {
		log.Fatal("SECRET must be set")
	}

	redirectURI := os.Getenv("REDIRECT_URI")
	if redirectURI == "" {
		log.Fatal("REDIRECT_URI must be set")
	}
	
	redirectTokenURI := os.Getenv("REDIRECT_TOKEN_URI") 
	if redirectTokenURI == "" {
		log.Fatal("REDIRECT_TOKEN_URI must be set")
	}
	
	user42InfoURL := os.Getenv("USER_INFO_URL")
	if user42InfoURL == "" {
		log.Fatal("USER_INFO_URL must be set")
	}
	
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	// open a connection to the database
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", dbURL)
	}

	//Using SQLC generated database package to create a new *database.Queries,
	// and storing in apiConfig struct so that handlers can access it:
	dbQueries := database.New(dbConn)

	// Configure OAuth2 for 42 Intra
	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"public"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  os.Getenv("OAUTH_AUTH_URI"),
			TokenURL: os.Getenv("OAUTH_TOKEN_URI"),
		},
	}

	apiCfg := &apiConfig{
		db: dbQueries,
		sessionStore: sessions.NewCookieStore([]byte(sessionSecret)),
        oauthConfig: oauthCfg,
		auth: auth.NewService(jwtSecret),
		redirectTokenURI: redirectTokenURI,
		user42InfoURL: user42InfoURL,
		jwtSecret: jwtSecret,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", healthHandler)

	mux.HandleFunc("GET /api/oauth/login", apiCfg.loginHandler)
	mux.HandleFunc("GET /oauth/callback", apiCfg.handlerCallback)

	mux.Handle(
		"POST /reservation",
		apiCfg.auth.Authenticate(
			auth.RequireAuth(
				http.HandlerFunc(apiCfg.handlerCreateReservation))))
	mux.Handle(
		"GET /reservation",
		apiCfg.auth.Authenticate(
			auth.RequireAuth(
				http.HandlerFunc(apiCfg.handlerFetchReservations))))
	// mux.HandleFunc("DELETE /reservation/{id}", apiCfg.handlerCallback) // to change

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
