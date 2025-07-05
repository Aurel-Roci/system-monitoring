package models

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

func SetupLogger(config *Config) error {
	// Determine log level
	var level slog.Level
	switch strings.ToLower(config.Logging.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Set up output destination
	var writer io.Writer
	if config.Logging.File == "" || config.Logging.File == "stdout" {
		writer = os.Stdout
	} else {
		file, err := os.OpenFile(config.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		writer = file
	}

	// Create structured logger
	opts := &slog.HandlerOptions{
		Level: level,
	}
	var handler slog.Handler
	if strings.ToLower(config.Environment) == "local" || strings.ToLower(config.Environment) == "development" {
		// Pretty printed text for local development
		handler = slog.NewTextHandler(writer, opts)
	} else {
		// JSON for production
		handler = slog.NewJSONHandler(writer, opts)
	}

	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	return nil
}
