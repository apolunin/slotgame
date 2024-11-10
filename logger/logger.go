package logger

import (
	"fmt"
	"github.com/apolunin/slotgame/config"
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
	"time"
)

const (
	FieldError  = "error"
	FieldUser   = "user"
	FieldLimit  = "limit"
	FieldOffset = "offset"
	FieldHash   = "hash"
)

func NewLogger(cfg config.LogConfig) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = slog.LevelInfo
	}

	switch cfg.Format {
	case config.TextFormat:
		return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:  cfg.AddSource,
			Level:      level,
			TimeFormat: time.DateTime,
		})), nil
	case config.JSONFormat:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: cfg.AddSource,
			Level:     level,
		})), nil
	default:
		return nil, fmt.Errorf("invalid log format: %s", cfg.Format)
	}
}
