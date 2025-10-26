package services_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "time"

    "go.uber.org/zap"

    "craftops/internal/config"
    "craftops/internal/services"
)

func TestNewServerService(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	service := services.NewServerService(cfg, logger)
	if service == nil {
		t.Fatal("NewServerService returned nil")
	}

	// Test that service was created successfully by calling public methods
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// GetStatus should work even if server is not running
	status, err := service.GetStatus(ctx)
	if err != nil {
		t.Errorf("GetStatus should not error: %v", err)
	}
	if status == nil {
		t.Error("GetStatus should return status object")
	}
}

func TestServerServiceHealthCheck(t *testing.T) {
    cfg := config.DefaultConfig()
    // Improve coverage: create a temp server dir and an empty server.jar to exercise OK branches
    tmp := t.TempDir()
    cfg.Paths.Server = tmp
    if err := os.WriteFile(filepath.Join(tmp, "server.jar"), []byte("jar"), 0o644); err != nil { t.Fatal(err) }
    logger := zap.NewNop()
    service := services.NewServerService(cfg, logger)

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

func TestServerStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	service := services.NewServerService(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test getting status (should not error even if server not running)
	status, err := service.GetStatus(ctx)
	if err != nil {
		t.Errorf("GetStatus should not error: %v", err)
	}

	if status == nil {
		t.Error("GetStatus should return a status object")
		return
	}

	// Status should have IsRunning field
	_ = status.IsRunning // This will compile-fail if field doesn't exist
}

func TestServerServiceDryRunStartStop(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DryRun = true
	logger := zap.NewNop()
	service := services.NewServerService(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start in dry-run should succeed: %v", err)
	}
	if err := service.Stop(ctx); err != nil {
		t.Fatalf("Stop in dry-run should succeed: %v", err)
	}
	if err := service.Restart(ctx); err != nil {
		t.Fatalf("Restart in dry-run should succeed: %v", err)
	}
}
