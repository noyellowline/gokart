package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Setup configures the internal sidecar logger.
// This is ONLY for sidecar internal logging (bootstrap, errors, debug).
// HTTP request logging is handled by the http_logging feature.
func Setup() {
	level := getLogLevel(os.Getenv("LOG_LEVEL"))

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Debug("sidecar logger initialized", "level", level.String())
}

// getLogLevel converts string log level to slog.Level
// Defaults to Info if invalid or empty
func getLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
