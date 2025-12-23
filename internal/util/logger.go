package util

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"

	"craftops/internal/config"
)

// NewLogger configures zap logging based on debug mode and config settings
func NewLogger(cfg *config.Config) *zap.Logger {
	var zapConfig zap.Config
	if cfg.Debug || cfg.Logging.Level == "DEBUG" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(parseLogLevel(cfg.Logging.Level))

	if cfg.Logging.Format == "text" {
		zapConfig.Encoding = "console"
		if term.IsTerminal(int(os.Stderr.Fd())) {
			zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	} else {
		zapConfig.Encoding = "json"
	}

	if !cfg.Logging.ConsoleEnabled {
		zapConfig.OutputPaths = []string{}
	}
	if cfg.Logging.FileEnabled {
		_ = os.MkdirAll(cfg.Paths.Logs, 0o750)
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, filepath.Join(cfg.Paths.Logs, "craftops.log"))
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return zap.NewNop()
	}
	return logger
}

// parseLogLevel safely converts a string to a zap log level
func parseLogLevel(levelStr string) zapcore.Level {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(levelStr))); err != nil {
		return zapcore.InfoLevel
	}
	return level
}

