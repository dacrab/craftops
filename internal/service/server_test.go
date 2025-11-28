package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/domain"
	"craftops/internal/service"
)

func TestNewServer(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	srv := service.NewServer(cfg, logger)
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServerStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	srv := service.NewServer(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := srv.Status(ctx)
	if err != nil {
		t.Errorf("Status should not error: %v", err)
	}
	if status == nil {
		t.Fatal("Status should return a status object")
	}
	_ = status.IsRunning
}

func TestServerHealthCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Server = tmp
	if err := os.WriteFile(filepath.Join(tmp, "server.jar"), []byte("jar"), 0o644); err != nil {
		t.Fatal(err)
	}

	logger := zap.NewNop()
	srv := service.NewServer(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	checks := srv.HealthCheck(ctx)
	if len(checks) == 0 {
		t.Error("HealthCheck should return at least one check")
	}

	for _, check := range checks {
		if check.Name == "" {
			t.Error("Health check name should not be empty")
		}
		if check.Status == "" {
			t.Error("Health check status should not be empty")
		}
	}
}

func TestServerDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DryRun = true
	logger := zap.NewNop()
	srv := service.NewServer(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		t.Fatalf("Start dry-run should succeed: %v", err)
	}
	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("Stop dry-run should succeed: %v", err)
	}
	if err := srv.Restart(ctx); err != nil {
		t.Fatalf("Restart dry-run should succeed: %v", err)
	}
}

func TestServerInterface(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	var _ service.ServerManager = service.NewServer(cfg, logger)
}

func TestHealthCheckStatus(t *testing.T) {
	tests := []domain.HealthStatus{
		domain.StatusOK,
		domain.StatusWarn,
		domain.StatusError,
	}

	for _, status := range tests {
		if status == "" {
			t.Errorf("status should not be empty")
		}
	}
}
