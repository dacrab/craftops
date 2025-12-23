package cli

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"

	"craftops/internal/config"
	"craftops/internal/service"
	"craftops/internal/ui"
)

// AppContainer is the central dependency injection container for the application
type AppContainer struct {
	Config       *config.Config
	Logger       *zap.Logger
	Terminal     *ui.Terminal
	Server       service.ServerManager
	Mods         service.ModManager
	Backup       service.BackupManager
	Notification service.Notifier
}

// NewApp wires up all services and dependencies based on the provided config
func NewApp(cfg *config.Config) *AppContainer {
	logger := initLogger(cfg)
	terminal := ui.NewTerminal()

	return &AppContainer{
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
func (a *AppContainer) Close() {
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
}

// initLogger configures zap logging based on debug mode and config settings
func initLogger(cfg *config.Config) *zap.Logger {
	var config zap.Config
	if cfg.Debug || cfg.Logging.Level == "DEBUG" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(parseLogLevel(cfg.Logging.Level))

	if cfg.Logging.Format == "text" {
		config.Encoding = "console"
		if term.IsTerminal(int(os.Stderr.Fd())) {
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	} else {
		config.Encoding = "json"
	}

	if !cfg.Logging.ConsoleEnabled {
		config.OutputPaths = []string{}
	}
	if cfg.Logging.FileEnabled {
		_ = os.MkdirAll(cfg.Paths.Logs, 0o750)
		config.OutputPaths = append(config.OutputPaths, filepath.Join(cfg.Paths.Logs, "craftops.log"))
	}

	logger, err := config.Build()
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
