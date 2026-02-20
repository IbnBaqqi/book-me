// Package config provides application configuration loading.
package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration needed to run the API
type Config struct {
	Server ServerConfig
	Logger LoggerConfig
	App    AppConfig
	Google GoogleConfig
	Email  EmailConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level string // debug, info, warn, error
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Env              string
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

// GoogleConfig holds Google Calendar configuration.
type GoogleConfig struct {
	CredentialsFile string
	CalendarScope   string
	CalendarID      string
}

// EmailConfig holds email service configuration.
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	UseTLS       bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, relying on system environment variables",
			"error", err,
		)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", "15s"),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", "60s"),
		},
		App: AppConfig{
			Env:              getEnv("ENV", "dev"),
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
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
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
		slog.Error("required environment variable not set",
			"key", key,
		)
		os.Exit(1)
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
		slog.Warn("invalid int environment variable, using default",
			"key", key,
			"value", valueStr,
			"default", defaultValue,
			"error", err,
		)
		return defaultValue
	}
	return value
}

//nolint:unused // kept for future use
func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		slog.Warn("invalid int environment variable, using default",
			"key", key,
			"value", valueStr,
			"default", defaultValue,
			"error", err,
		)
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key, defaultValue string) time.Duration {
	valueStr := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		// Fallback to parsing the default if provided value is invalid
		duration, err = time.ParseDuration(defaultValue)
		if err != nil {
			return 0
		}
	}
	return duration
}
