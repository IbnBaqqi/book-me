package config

import (
	"log/slog"
	"os"
	"strings"
)

// New creates a new structured logger based on configuration
func (c *LoggerConfig) New() *slog.Logger {
	var handler slog.Handler

	level := parseLogLevel(c.Level)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug || level == slog.LevelError,
	}

	// Use text handler in dev, JSON in prod
	if os.Getenv("ENVIRONMENT") == "dev" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
