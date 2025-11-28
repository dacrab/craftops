package service

import (
	"context"

	"craftops/internal/domain"
)

// ServerManager handles server lifecycle operations
type ServerManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Restart(ctx context.Context) error
	Status(ctx context.Context) (*domain.ServerStatus, error)
	HealthCheck(ctx context.Context) []domain.HealthCheck
}

// ModManager handles mod management operations
type ModManager interface {
	UpdateAll(ctx context.Context, force bool) (*domain.ModUpdateResult, error)
	ListInstalled() ([]domain.InstalledMod, error)
	HealthCheck(ctx context.Context) []domain.HealthCheck
}

// BackupManager handles backup operations
type BackupManager interface {
	Create(ctx context.Context) (string, error)
	List() ([]domain.BackupInfo, error)
	HealthCheck(ctx context.Context) []domain.HealthCheck
}

// Notifier handles notification operations
type Notifier interface {
	SendSuccess(ctx context.Context, message string) error
	SendError(ctx context.Context, message string) error
	SendRestartWarnings(ctx context.Context) error
	HealthCheck(ctx context.Context) []domain.HealthCheck
}
