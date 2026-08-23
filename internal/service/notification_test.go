package service

import (
	"context"
	"testing"
	"time"

	"craftops/internal/domain"
)

func TestNotification_HealthCheck_NoWebhook(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.DiscordWebhook = ""
	svc := NewNotification(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if len(checks) < 1 {
		t.Fatal("expected at least one health check")
	}
	if checks[0].Status != domain.StatusWarn {
		t.Errorf("expected WARN for missing webhook, got %s", checks[0].Status)
	}
}

func TestNotification_HealthCheck_InvalidWebhook(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.DiscordWebhook = "https://invalid.example.com/webhook"
	svc := NewNotification(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if checks[0].Status != domain.StatusError {
		t.Errorf("expected ERROR for invalid webhook URL, got %s", checks[0].Status)
	}
}

func TestNotification_HealthCheck_ValidWebhook(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.DiscordWebhook = testDiscordWebhook
	svc := NewNotification(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if checks[0].Status != domain.StatusOK {
		t.Errorf("expected OK for valid webhook URL, got %s: %s", checks[0].Status, checks[0].Message)
	}
}

func TestNotification_HealthCheck_AllDisabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.SuccessNotifications = false
	cfg.Notifications.ErrorNotifications = false
	svc := NewNotification(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if len(checks) < 2 {
		t.Fatalf("expected 2 checks, got %d", len(checks))
	}
	if checks[1].Status != domain.StatusWarn {
		t.Errorf("expected WARN when all notifications disabled, got %s", checks[1].Status)
	}
}

func TestNotification_SendSuccess_Disabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.SuccessNotifications = false
	svc := NewNotification(cfg, logger)

	if err := svc.SendSuccess(ctx, "hello"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestNotification_SendError_Disabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.ErrorNotifications = false
	svc := NewNotification(cfg, logger)

	if err := svc.SendError(ctx, "boom"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestNotification_DryRun(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.DryRun = true
	cfg.Notifications.DiscordWebhook = testDiscordWebhook
	cfg.Notifications.SuccessNotifications = true
	svc := NewNotification(cfg, logger)

	if err := svc.SendSuccess(ctx, "test"); err != nil {
		t.Errorf("dry-run SendSuccess failed: %v", err)
	}
}

func TestNotification_SendRestartWarnings_Empty(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.WarningIntervals = []int{}
	svc := NewNotification(cfg, logger)

	if err := svc.SendRestartWarnings(ctx); err != nil {
		t.Errorf("expected nil with empty intervals, got %v", err)
	}
}

func TestNotification_SendRestartWarnings_NoWebhook(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.DiscordWebhook = ""
	cfg.Notifications.WarningIntervals = []int{0}
	svc := NewNotification(cfg, logger)

	if err := svc.SendRestartWarnings(ctx); err != nil {
		t.Errorf("expected nil when no webhook configured, got %v", err)
	}
}

func TestNotification_SendRestartWarnings_SortedLongestFirst(t *testing.T) {
	cfg, logger, _ := setup(t)
	cfg.Notifications.DiscordWebhook = ""
	cfg.Notifications.WarningIntervals = []int{0}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	svc := NewNotification(cfg, logger)
	if err := svc.SendRestartWarnings(ctx); err != nil {
		t.Errorf("SendRestartWarnings: %v", err)
	}
}

func TestNotification_SendSuccess_WithWebhook_DryRun(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.DryRun = true
	cfg.Notifications.DiscordWebhook = testDiscordWebhook
	cfg.Notifications.SuccessNotifications = true
	cfg.Notifications.ErrorNotifications = true
	svc := NewNotification(cfg, logger)

	if err := svc.SendSuccess(ctx, "server started"); err != nil {
		t.Errorf("SendSuccess dry-run: %v", err)
	}
	if err := svc.SendError(ctx, "something broke"); err != nil {
		t.Errorf("SendError dry-run: %v", err)
	}
}
