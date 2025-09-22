package cli

import (
	"craftops/internal/config"
	"craftops/internal/services"
)

// ServiceFactory defines the interface for creating services.
type ServiceFactory interface {
	GetBackupService() services.BackupServiceInterface
	GetModService() services.ModServiceInterface
	GetNotificationService() services.NotificationServiceInterface
	GetServerService() services.ServerServiceInterface
	GetConfig() *config.Config
}
