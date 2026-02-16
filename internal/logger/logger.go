// Package logger provides structured logging utilities.
package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/IbnBaqqi/book-me/internal/config"
)

// Log is the global structured logger instance.
var Log *slog.Logger

// RetryLogger adapts the retryablehttp.LeveledLogger interface to slog logger
type RetryLogger struct{}

// Error logs a error-level message.
func (l *RetryLogger) Error(msg string, keysAndValues ...interface{}) {
	Log.Error(msg, keysAndValues...)
}

// Info logs an info-level message.
func (l *RetryLogger) Info(msg string, keysAndValues ...interface{}) {
	Log.Info(msg, keysAndValues...)
}

// Debug logs a debug-level message.
func (l *RetryLogger) Debug(msg string, keysAndValues ...interface{}) {
	Log.Debug(msg, keysAndValues...)
}

// Warn logs a warning-level message.
func (l *RetryLogger) Warn(msg string, keysAndValues ...interface{}) {
	Log.Warn(msg, keysAndValues...)
}

// Init creates the global structured logger based on configuration
func Init(c *config.LoggerConfig, env string) {
	var handler slog.Handler

	level := parseLogLevel(c.Level)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug || level == slog.LevelError,
	}

	// Use text handler in dev, JSON in prod
	if env == "dev" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
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
