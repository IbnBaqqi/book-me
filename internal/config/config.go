package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration needed to run the API
type Config struct {
	App    AppConfig
	Google GoogleConfig
	Email  EmailConfig
}

type AppConfig struct {
	Port             string
	DBURL            string
	SessionSecret    string
	ClientID         string
	ClientSecret     string
	RedirectURI      string
	RedirectTokenURI string
	User42InfoURL    string
	JWTSecret        string
	OAuthAuthURI     string
	OAuthTokenURI    string
}

// Google Calendar config
type GoogleConfig struct {
	CredentialsFile string
	CalendarScope   string
	CalendarID      string
}

// Email config
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	UseTLS       bool
}

// LoadConfig loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	cfg := &Config{
		App: AppConfig{
			Port:             getEnv("PORT", "8080"),
			DBURL:            mustGetEnv("DB_URL"),
			SessionSecret:    mustGetEnv("SESSION_SECRET"),
			ClientID:         mustGetEnv("CLIENT_ID"),
			ClientSecret:     mustGetEnv("SECRET"),
			RedirectURI:      mustGetEnv("REDIRECT_URI"),
			RedirectTokenURI: mustGetEnv("REDIRECT_TOKEN_URI"),
			User42InfoURL:    mustGetEnv("USER_INFO_URL"),
			JWTSecret:        mustGetEnv("JWT_SECRET"),
			OAuthAuthURI:     mustGetEnv("OAUTH_AUTH_URI"),
			OAuthTokenURI:    mustGetEnv("OAUTH_TOKEN_URI"),
		},
		Google: GoogleConfig{
			CredentialsFile: mustGetEnv("GOOGLE_CREDENTIALS_FILE"),
			CalendarScope:   mustGetEnv("GOOGLE_CALENDAR_SCOPE"),
			CalendarID:      mustGetEnv("GOOGLE_CALENDAR_ID"),
		},
		Email: EmailConfig{
			SMTPHost:     mustGetEnv("SMTP_HOST"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUsername: mustGetEnv("SMTP_USERNAME"),
			SMTPPassword: mustGetEnv("SMTP_PASSWORD"),
			FromEmail:    mustGetEnv("FROM_EMAIL"),
			FromName:     getEnv("FROM_NAME", "BookMe"),
			UseTLS:       getEnv("SMTP_USE_TLS", "true") == "true",
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable must be set", key)
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		log.Printf("Invalid value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return value
}
