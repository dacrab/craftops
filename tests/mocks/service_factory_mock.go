package mocks

import (
	"craftops/internal/config"
	"craftops/internal/services"
)

// MockServiceFactory is a mock implementation of the ServiceFactory interface.
type MockServiceFactory struct {
	BackupService       services.BackupServiceInterface
	ModService          services.ModServiceInterface
	NotificationService services.NotificationServiceInterface
	ServerService       services.ServerServiceInterface
	Config              *config.Config
}

// GetBackupService returns the mock backup service.
func (m *MockServiceFactory) GetBackupService() services.BackupServiceInterface {
	return m.BackupService
}

// GetModService returns the mock mod service.
func (m *MockServiceFactory) GetModService() services.ModServiceInterface {
	return m.ModService
}

// GetNotificationService returns the mock notification service.
func (m *MockServiceFactory) GetNotificationService() services.NotificationServiceInterface {
	return m.NotificationService
}

// GetServerService returns the mock server service.
func (m *MockServiceFactory) GetServerService() services.ServerServiceInterface {
	return m.ServerService
}

// (metrics removed)

// GetConfig returns the mock configuration.
func (m *MockServiceFactory) GetConfig() *config.Config {
	if m.Config == nil {
		m.Config = &config.Config{
			Backup: config.BackupConfig{Enabled: false},
			Notifications: config.NotificationConfig{
				WarningIntervals: []int{},
			},
		}
	}
	return m.Config
}
