// Package service defines the business logic and external integrations for CraftOps.
package service

import (
	"context"

	"craftops/internal/domain"
)

// ServerManager defines operations for managing the Minecraft server process.
type ServerManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Restart(ctx context.Context) error
	Status(ctx context.Context) (*domain.ServerStatus, error)
	HealthCheck(ctx context.Context) []domain.HealthCheck
}

// ModManager defines operations for automated mod management.
type ModManager interface {
	UpdateAll(ctx context.Context, force bool) (*domain.ModUpdateResult, error)
	ListInstalled() ([]domain.InstalledMod, error)
	HealthCheck(ctx context.Context) []domain.HealthCheck
}

// BackupManager defines operations for server backups and retention.
type BackupManager interface {
	Create(ctx context.Context) (string, error)
	List() ([]domain.BackupInfo, error)
	HealthCheck(ctx context.Context) []domain.HealthCheck
}

// Notifier defines operations for sending alerts and status updates.
type Notifier interface {
	SendSuccess(ctx context.Context, message string) error
	SendError(ctx context.Context, message string) error
	SendRestartWarnings(ctx context.Context) error
	HealthCheck(ctx context.Context) []domain.HealthCheck
}

// Ensure compile-time interface satisfaction.
var (
	_ ServerManager = (*Server)(nil)
	_ ModManager    = (*Mods)(nil)
	_ BackupManager = (*Backup)(nil)
	_ Notifier      = (*Notification)(nil)
)
