package services_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/services"
)

func TestNewNotificationService(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	service := services.NewNotificationService(cfg, logger)
	if service == nil {
		t.Fatal("NewNotificationService returned nil")
	}

	// Test that service was created successfully by calling public methods
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// These should not error even without webhook configured
	err := service.SendSuccessNotification(ctx, "test")
	if err != nil {
		t.Errorf("SendSuccessNotification should not error without webhook: %v", err)
	}
}

func TestNotificationServiceHealthCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	service := services.NewNotificationService(cfg, logger)

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

func TestNotificationDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DryRun = true // Enable dry run mode
	logger := zap.NewNop()
	service := services.NewNotificationService(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// These should not error in dry run mode
	err := service.SendSuccessNotification(ctx, "test message")
	if err != nil {
		t.Errorf("SendSuccessNotification should not error in dry run: %v", err)
	}

	err = service.SendErrorNotification(ctx, "test error")
	if err != nil {
		t.Errorf("SendErrorNotification should not error in dry run: %v", err)
	}
}
