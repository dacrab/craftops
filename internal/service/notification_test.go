package service_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/service"
)

func TestNewNotification(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	svc := service.NewNotification(cfg, logger)
	if svc == nil {
		t.Fatal("NewNotification returned nil")
	}
}

func TestNotificationSendDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DryRun = true
	cfg.Notifications.DiscordWebhook = "https://discord.com/api/webhooks/test"
	cfg.Notifications.SuccessNotifications = true
	cfg.Notifications.ErrorNotifications = true
	logger := zap.NewNop()
	svc := service.NewNotification(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := svc.SendSuccess(ctx, "test success"); err != nil {
		t.Errorf("SendSuccess dry-run failed: %v", err)
	}
	if err := svc.SendError(ctx, "test error"); err != nil {
		t.Errorf("SendError dry-run failed: %v", err)
	}
}

func TestNotificationNoWebhook(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Notifications.DiscordWebhook = ""
	cfg.Notifications.SuccessNotifications = true
	cfg.Notifications.ErrorNotifications = true
	logger := zap.NewNop()
	svc := service.NewNotification(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should succeed even without webhook (just skip)
	if err := svc.SendSuccess(ctx, "test"); err != nil {
		t.Errorf("SendSuccess should not fail without webhook: %v", err)
	}
}

func TestNotificationDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Notifications.DiscordWebhook = "https://discord.com/api/webhooks/test"
	cfg.Notifications.SuccessNotifications = false
	cfg.Notifications.ErrorNotifications = false
	logger := zap.NewNop()
	svc := service.NewNotification(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should succeed because notifications are disabled
	if err := svc.SendSuccess(ctx, "test"); err != nil {
		t.Errorf("SendSuccess should succeed when disabled: %v", err)
	}
	if err := svc.SendError(ctx, "test"); err != nil {
		t.Errorf("SendError should succeed when disabled: %v", err)
	}
}

func TestNotificationHealthCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	svc := service.NewNotification(cfg, logger)

	ctx := context.Background()
	checks := svc.HealthCheck(ctx)
	if len(checks) == 0 {
		t.Error("HealthCheck should return checks")
	}
}

func TestNotificationInterface(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	var _ service.Notifier = service.NewNotification(cfg, logger)
}
