package config

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger(cfg *Config) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(parseLogLevel(cfg.Log.Level))

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler).With("service", cfg.App.Name)
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
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
