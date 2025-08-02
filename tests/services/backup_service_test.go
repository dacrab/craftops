package services_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/services"
)

func TestNewBackupService(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	service := services.NewBackupService(cfg, logger)
	if service == nil {
		t.Fatal("NewBackupService returned nil")
	}

	// Test that service was created successfully (can't access private fields)
	// Verify by calling a public method
	backups, err := service.ListBackups()
	if err != nil {
		t.Errorf("Service not properly initialized: %v", err)
	}
	if backups == nil {
		t.Error("Service should return empty slice, not nil")
	}
}

func TestBackupServiceHealthCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	service := services.NewBackupService(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	checks := service.HealthCheck(ctx)
	if len(checks) == 0 {
		t.Error("HealthCheck should return at least one check")
	}

	// Verify check structure
	for _, check := range checks {
		if check.Name == "" {
			t.Error("Health check name should not be empty")
		}
		if check.Status == "" {
			t.Error("Health check status should not be empty")
		}
		if check.Message == "" {
			t.Error("Health check message should not be empty")
		}
	}
}

func TestListBackups(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	service := services.NewBackupService(cfg, logger)

	// Should not error even if backup directory doesn't exist
	backups, err := service.ListBackups()
	if err != nil {
		t.Errorf("ListBackups should not error when directory doesn't exist: %v", err)
	}

	if backups == nil {
		t.Error("ListBackups should return empty slice, not nil")
	}
}
