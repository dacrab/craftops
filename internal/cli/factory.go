package cli

import (
	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/services"
)

// serviceFactory handles the creation of services and other dependencies.
type serviceFactory struct {
	config *config.Config
	logger *zap.Logger
}

// NewServiceFactory creates a new service factory.
func NewServiceFactory(config *config.Config, logger *zap.Logger) ServiceFactory {
	return &serviceFactory{
		config: config,
		logger: logger,
	}
}

// GetBackupService creates a new backup service.
func (f *serviceFactory) GetBackupService() services.BackupServiceInterface {
	return services.NewBackupService(f.config, f.logger)
}

// GetModService creates a new mod service.
func (f *serviceFactory) GetModService() services.ModServiceInterface {
	return services.NewModService(f.config, f.logger)
}

// GetNotificationService creates a new notification service.
func (f *serviceFactory) GetNotificationService() services.NotificationServiceInterface {
	return services.NewNotificationService(f.config, f.logger)
}

// GetServerService creates a new server service.
func (f *serviceFactory) GetServerService() services.ServerServiceInterface {
	return services.NewServerService(f.config, f.logger)
}

// GetConfig returns the configuration.
func (f *serviceFactory) GetConfig() *config.Config {
	return f.config
}
