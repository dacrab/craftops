package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/service"
)

func TestNewMods(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	svc := service.NewMods(cfg, logger)
	if svc == nil {
		t.Fatal("NewMods returned nil")
	}
}

func TestModsListInstalled(t *testing.T) {
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Mods = tmp

	// Create some test jar files
	for _, name := range []string{"mod1.jar", "mod2.jar"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("jar"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	logger := zap.NewNop()
	svc := service.NewMods(cfg, logger)

	mods, err := svc.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}
	if len(mods) != 2 {
		t.Errorf("expected 2 mods, got %d", len(mods))
	}
}

func TestModsHealthCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Mods = tmp

	logger := zap.NewNop()
	svc := service.NewMods(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	checks := svc.HealthCheck(ctx)
	if len(checks) == 0 {
		t.Error("HealthCheck should return at least one check")
	}
}

func TestModsUpdateDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DryRun = true
	logger := zap.NewNop()
	svc := service.NewMods(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := svc.UpdateAll(ctx, false)
	if err != nil {
		t.Fatalf("UpdateAll dry-run failed: %v", err)
	}
	if result == nil {
		t.Fatal("UpdateAll should return result")
	}
}

func TestModsUpdateNoSources(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Mods.ModrinthSources = []string{}
	logger := zap.NewNop()
	svc := service.NewMods(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := svc.UpdateAll(ctx, false)
	if err != nil {
		t.Fatalf("UpdateAll should not fail with no sources: %v", err)
	}
	if len(result.UpdatedMods) != 0 {
		t.Error("No sources should mean no updates")
	}
}

func TestModsInterface(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	var _ service.ModManager = service.NewMods(cfg, logger)
}
