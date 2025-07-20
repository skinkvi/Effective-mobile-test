package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/skinkvi/effective_mobile/internal/config"
)

const (
	defaultLoggerLevel = zapcore.DebugLevel
)

var (
	once   sync.Once
	logger *zap.Logger
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	var logLevel zapcore.Level

	once.Do(func() {
		switch cfg.Env {
		case "local", "dev":
			logLevel = zapcore.DebugLevel
		case "prod":
			logLevel = zapcore.InfoLevel
		default:
			logLevel = defaultLoggerLevel
		}

		zapConfig := zap.Config{
			Level:       zap.NewAtomicLevelAt(logLevel),
			Development: cfg.Env != "prod",
			Encoding: func() string {
				if cfg.Env == "local" {
					return "console"
				}
				return "json"
			}(),
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "time",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		logger, _ = zapConfig.Build()
	})

	return logger, nil
}
