package config

import (
	"log/slog"
	"os"
)

func NewLogger(cfg *Config) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, nil)

	return slog.New(handler).With("service", cfg.App.Name)
}
