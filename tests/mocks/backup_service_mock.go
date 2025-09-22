package mocks

import (
	"context"

	"craftops/internal/services"
)

// MockBackupService is a mock implementation of the BackupServiceInterface.
type MockBackupService struct {
	CreateBackupFunc func(ctx context.Context) (string, error)
	ListBackupsFunc  func() ([]services.BackupInfo, error)
    RestoreBackupFunc func(ctx context.Context, path string, force bool) error
	HealthCheckFunc  func(ctx context.Context) []services.HealthCheck
}

// CreateBackup mocks the CreateBackup method.
func (m *MockBackupService) CreateBackup(ctx context.Context) (string, error) {
	if m.CreateBackupFunc != nil {
		return m.CreateBackupFunc(ctx)
	}
	return "mock_backup.tar.gz", nil
}

// ListBackups mocks the ListBackups method.
func (m *MockBackupService) ListBackups() ([]services.BackupInfo, error) {
	if m.ListBackupsFunc != nil {
		return m.ListBackupsFunc()
	}
	return []services.BackupInfo{
		{Name: "mock_backup.tar.gz", CreatedAt: "2023-01-01 12:00:00", Size: "10 MB"},
	}, nil
}

// RestoreBackup mocks the RestoreBackup method.
func (m *MockBackupService) RestoreBackup(ctx context.Context, path string, force bool) error {
    if m.RestoreBackupFunc != nil {
        return m.RestoreBackupFunc(ctx, path, force)
    }
    return nil
}

// HealthCheck mocks the HealthCheck method.
func (m *MockBackupService) HealthCheck(ctx context.Context) []services.HealthCheck {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return []services.HealthCheck{
		{Name: "Mock Backup Check", Status: "✅", Message: "OK"},
	}
}
