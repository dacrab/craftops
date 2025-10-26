package services_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/services"
)

func TestBackupServiceVariants(t *testing.T) {
	t.Run("new service + list", func(t *testing.T) {
		cfg := config.DefaultConfig()
		logger := zap.NewNop()

		svc := services.NewBackupService(cfg, logger)
		if svc == nil {
			t.Fatal("NewBackupService returned nil")
		}
		backups, err := svc.ListBackups()
		if err != nil {
			t.Errorf("ListBackups error: %v", err)
		}
		if backups == nil {
			t.Error("ListBackups should return empty slice, not nil")
		}
	})

	t.Run("health check", func(t *testing.T) {
		cfg := config.DefaultConfig()
		logger := zap.NewNop()
		svc := services.NewBackupService(cfg, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checks := svc.HealthCheck(ctx)
		if len(checks) == 0 {
			t.Error("HealthCheck should return at least one check")
		}
		for _, check := range checks {
			if check.Name == "" {
				t.Error("health check name should not be empty")
			}
			if check.Status == "" {
				t.Error("health check status should not be empty")
			}
			if check.Message == "" {
				t.Error("health check message should not be empty")
			}
		}
	})

	t.Run("create backup (dry-run)", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.DryRun = true
		logger := zap.NewNop()
		svc := services.NewBackupService(cfg, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		path, err := svc.CreateBackup(ctx)
		if err != nil {
			t.Fatalf("CreateBackup(dry-run) error: %v", err)
		}
		if path != "dry-run-backup.tar.gz" {
			t.Errorf("unexpected dry-run path = %q", path)
		}
	})
}
