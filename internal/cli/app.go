package cli

import (
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
// cfg.Logging.Level is normalised to uppercase by config.Validate().
func newLogger(cfg *config.Config) *zap.Logger {
	level := zap.InfoLevel
	if cfg.Logging.Level == "DEBUG" {
		level = zap.DebugLevel
	}
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	if cfg.Logging.Format == "text" {
		zapCfg.Encoding = "console"
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	logger, err := zapCfg.Build()
	if err != nil {
		// Fall back to a no-op logger rather than panicking on misconfiguration.
		return zap.NewNop()
	}
	return logger
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
