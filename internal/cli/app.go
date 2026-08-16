// Package cli wires Cobra commands to service layer operations.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"craftops/internal/config"
	"craftops/internal/service"
	"craftops/internal/ui"
)

type app struct {
	Config       *config.Config
	Logger       *zap.Logger
	Terminal     *ui.Terminal
	Server       *service.Server
	Mods         *service.Mods
	Backup       *service.Backup
	Notification *service.Notification
}

func newLogger(cfg *config.Config) *zap.Logger {
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if cfg.Logging.Level == "DEBUG" {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	if cfg.Logging.Format == "text" {
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var cores []zapcore.Core

	if cfg.Logging.FileEnabled && cfg.Paths.Logs != "" {
		if err := os.MkdirAll(cfg.Paths.Logs, 0o750); err != nil {
			loggerWarnf("failed to create log directory %s: %v", cfg.Paths.Logs, err)
		} else {
			logPath := filepath.Join(cfg.Paths.Logs, "craftops.log")
			f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
			if err != nil {
				loggerWarnf("failed to open log file %s: %v", logPath, err)
			} else {
				var enc zapcore.Encoder
				if cfg.Logging.Format == "text" {
					enc = zapcore.NewConsoleEncoder(encoderCfg)
				} else {
					enc = zapcore.NewJSONEncoder(encoderCfg)
				}
				cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(f), level))
			}
		}
	}

	if cfg.Logging.ConsoleEnabled {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.AddSync(os.Stderr),
			level,
		))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.AddSync(os.Stderr),
			level,
		))
	}
	return zap.New(zapcore.NewTee(cores...))
}

func loggerWarnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "craftops: warning: "+format+"\n", args...)
}

func newApp(cfg *config.Config) *app {
	logger := newLogger(cfg)
	return &app{
		Config:       cfg,
		Logger:       logger,
		Terminal:     ui.NewTerminal(),
		Server:       service.NewServer(cfg, logger),
		Mods:         service.NewMods(cfg, logger),
		Backup:       service.NewBackup(cfg, logger),
		Notification: service.NewNotification(cfg, logger),
	}
}

func (a *app) Close() {
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
}
