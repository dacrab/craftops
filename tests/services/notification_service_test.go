package services_test

import (
    "context"
    "net/http"
    "net/http/httptest"
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
    cfg.Notifications.WarningIntervals = []int{1}
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

    // Also exercise restart warnings path with a single short interval
    if err := service.SendRestartWarnings(ctx); err != nil {
        t.Errorf("SendRestartWarnings should not error in dry run: %v", err)
    }

    // Increase coverage: exercise non-dry-run webhook path with a local server
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(204)
    }))
    defer srv.Close()

    cfg2 := config.DefaultConfig()
    cfg2.DryRun = false
    cfg2.Notifications.DiscordWebhook = srv.URL
    service2 := services.NewNotificationService(cfg2, logger)
    if err := service2.SendSuccessNotification(ctx, "ok"); err != nil {
        t.Errorf("SendSuccessNotification (webhook) error: %v", err)
    }
}

func TestNotificationWebhookCheckVariants(t *testing.T) {
    cfg := config.DefaultConfig()
    logger := zap.NewNop()
    svc := services.NewNotificationService(cfg, logger)
    // Empty -> WARN exists already in other test via HealthCheck, add ERROR and OK
    cfg.Notifications.DiscordWebhook = "http://foo"
    svc = services.NewNotificationService(cfg, logger)
    found := false
    for _, c := range svc.HealthCheck(context.Background()) {
        if c.Name == "Discord webhook" {
            found = true
            if c.Status != "ERROR" {
                t.Fatalf("expected ERROR, got %s", c.Status)
            }
        }
    }
    if !found { t.Fatalf("missing webhook check") }

    cfg.Notifications.DiscordWebhook = "https://discord.com/api/webhooks/abc"
    svc = services.NewNotificationService(cfg, logger)
    ok := false
    for _, c := range svc.HealthCheck(context.Background()) {
        if c.Name == "Discord webhook" && c.Status == "OK" { ok = true }
    }
    if !ok { t.Fatalf("expected OK webhook status") }
}
