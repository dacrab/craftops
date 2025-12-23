package cli

import (
	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/service"
	"craftops/internal/ui"
	"craftops/internal/util"
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
	logger := util.NewLogger(cfg)
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
