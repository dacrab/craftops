package cli

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"craftops/internal/config"
	"craftops/internal/service"
	"craftops/internal/ui"
)

// appContainer is the central dependency injection container for the application
type appContainer struct {
	Config       *config.Config
	Logger       *zap.Logger
	Terminal     *ui.Terminal
	Server       service.ServerManager
	Mods         service.ModManager
	Backup       service.BackupManager
	Notification service.Notifier
}

// newLogger builds a zap logger configured from the app config.
// Respects FileEnabled (writes to Paths.Logs/craftops.log) and ConsoleEnabled.
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

	encoding := "json"
	if cfg.Logging.Format == "text" {
		encoding = "console"
	}

	var cores []zapcore.Core

	if cfg.Logging.ConsoleEnabled {
		consoleEnc := zapcore.NewConsoleEncoder(encoderCfg)
		cores = append(cores, zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stderr), level))
	}

	if cfg.Logging.FileEnabled && cfg.Paths.Logs != "" {
		if err := os.MkdirAll(cfg.Paths.Logs, 0o750); err == nil {
			logPath := filepath.Join(cfg.Paths.Logs, "craftops.log")
			if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil { //nolint:gosec
				var enc zapcore.Encoder
				if encoding == "console" {
					enc = zapcore.NewConsoleEncoder(encoderCfg)
				} else {
					enc = zapcore.NewJSONEncoder(encoderCfg)
				}
				cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(f), level))
			}
		}
	}

	if len(cores) == 0 {
		return zap.NewNop()
	}
	return zap.New(zapcore.NewTee(cores...))
}

// newApp wires up all services and dependencies based on the provided config
func newApp(cfg *config.Config) *appContainer {
	logger := newLogger(cfg)
	terminal := ui.NewTerminal()

	return &appContainer{
		Config:       cfg,
		Logger:       logger,
		Terminal:     terminal,
		Server:       service.NewServer(cfg, logger),
		Mods:         service.NewMods(cfg, logger),
		Backup:       service.NewBackup(cfg, logger),
		Notification: service.NewNotification(cfg, logger),
	}
}

// Close ensures all resources (like log buffers) are properly flushed
func (a *appContainer) Close() {
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
}
