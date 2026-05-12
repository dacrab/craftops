package service_test

import (
	"testing"

	"craftops/internal/service"
)

func TestServer_HealthCheck(t *testing.T) {
	cfg, logger, ctx := setup(t)
	svc := service.NewServer(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if len(checks) < 2 {
		t.Fatalf("expected at least 2 health checks, got %d", len(checks))
	}
	names := make(map[string]bool)
	for _, c := range checks {
		names[c.Name] = true
	}
	if !names["Server directory"] {
		t.Error("expected 'Server directory' check")
	}
	if !names["Server JAR"] {
		t.Error("expected 'Server JAR' check")
	}
}

func TestServer_Status_ReturnsResult(t *testing.T) {
	cfg, logger, ctx := setup(t)
	svc := service.NewServer(cfg, logger)

	status, err := svc.Status(ctx)
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if status == nil {
		t.Fatal("Status() returned nil")
	}
	if status.CheckedAt.IsZero() {
		t.Error("Status.CheckedAt should not be zero")
	}
	if status.SessionName == "" {
		t.Error("Status.SessionName should not be empty")
	}
}

func TestServer_Start_DryRun(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.DryRun = true
	svc := service.NewServer(cfg, logger)

	if err := svc.Start(ctx); err != nil {
		t.Errorf("Start() dry-run error: %v", err)
	}
}

func TestServer_Stop_DryRun(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.DryRun = true
	svc := service.NewServer(cfg, logger)

	if err := svc.Stop(ctx); err != nil {
		t.Errorf("Stop() dry-run error: %v", err)
	}
}
