package app

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"

	"craftops/internal/config"
	"craftops/internal/service"
	"craftops/internal/ui"
)

// App is the application container holding all dependencies
type App struct {
	Config       *config.Config
	Logger       *zap.Logger
	Terminal     *ui.Terminal
	Server       service.ServerManager
	Mods         service.ModManager
	Backup       service.BackupManager
	Notification service.Notifier
}

// New creates a new application instance with all dependencies
func New(cfg *config.Config) *App {
	logger := initLogger(cfg)
	terminal := ui.NewTerminal()

	return &App{
		Config:       cfg,
		Logger:       logger,
		Terminal:     terminal,
		Server:       service.NewServer(cfg, logger),
		Mods:         service.NewMods(cfg, logger),
		Backup:       service.NewBackup(cfg, logger),
		Notification: service.NewNotification(cfg, logger),
	}
}

// Close cleans up resources
func (a *App) Close() {
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
}

func initLogger(cfg *config.Config) *zap.Logger {
	level := parseLogLevel(cfg.Logging.Level)

	var encoderConfig zapcore.EncoderConfig
	isTTY := term.IsTerminal(int(os.Stderr.Fd()))

	if cfg.Logging.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		if isTTY {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		} else {
			encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		}
	}

	var cores []zapcore.Core

	if cfg.Logging.ConsoleEnabled {
		var encoder zapcore.Encoder
		if cfg.Logging.Format == "json" {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), level))
	}

	if cfg.Logging.FileEnabled {
		_ = os.MkdirAll(cfg.Paths.Logs, 0o750)
		logFile := filepath.Join(cfg.Paths.Logs, "craftops.log")
		if file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
			var encoder zapcore.Encoder
			if cfg.Logging.Format == "json" {
				encoder = zapcore.NewJSONEncoder(encoderConfig)
			} else {
				encoder = zapcore.NewConsoleEncoder(encoderConfig)
			}
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(file), level))
		}
	}

	if len(cores) == 0 {
		return zap.NewNop()
	}

	return zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func parseLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARNING":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "CRITICAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
