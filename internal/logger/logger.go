package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/akhilr007/tasks/internal/config"
)

func New(cfg config.LoggerConfig) *slog.Logger {
	level := parseLevel(cfg.Level)

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
