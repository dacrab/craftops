package services

import "context"

// BackupServiceInterface defines the interface for backup operations.
type BackupServiceInterface interface {
	CreateBackup(ctx context.Context) (string, error)
	ListBackups() ([]BackupInfo, error)
	RestoreBackup(ctx context.Context, backupPath string, force bool) error
	HealthCheck(ctx context.Context) []HealthCheck
}

// ModServiceInterface defines the interface for mod management operations.
type ModServiceInterface interface {
	UpdateAllMods(ctx context.Context, force bool) (*ModUpdateResult, error)
	ListInstalledMods() ([]map[string]interface{}, error)
	HealthCheck(ctx context.Context) []HealthCheck
}

// NotificationServiceInterface defines the interface for notification operations.
type NotificationServiceInterface interface {
	SendSuccessNotification(ctx context.Context, message string) error
	SendErrorNotification(ctx context.Context, message string) error
	SendRestartWarnings(ctx context.Context) error
	HealthCheck(ctx context.Context) []HealthCheck
}

// ServerServiceInterface defines the interface for server management operations.
type ServerServiceInterface interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Restart(ctx context.Context) error
	GetStatus(ctx context.Context) (*ServerStatus, error)
	HealthCheck(ctx context.Context) []HealthCheck
}
