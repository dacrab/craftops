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

func setup(t *testing.T) (*config.Config, *zap.Logger, context.Context) {
	t.Helper()
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Server = filepath.Join(tmp, "server")
	cfg.Paths.Mods = filepath.Join(tmp, "mods")
	cfg.Paths.Backups = filepath.Join(tmp, "backups")
	cfg.Paths.Logs = filepath.Join(tmp, "logs")

	for _, p := range []string{cfg.Paths.Server, cfg.Paths.Mods, cfg.Paths.Backups, cfg.Paths.Logs} {
		_ = os.MkdirAll(p, 0o755)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	return cfg, zap.NewNop(), ctx
}

// ============================================================================
// SERVER TESTS
// ============================================================================

func TestServer(t *testing.T) {
	cfg, logger, ctx := setup(t)
	srv := service.NewServer(cfg, logger)

	t.Run("Status", func(t *testing.T) {
		status, err := srv.Status(ctx)
		if err != nil || status == nil {
			t.Errorf("Status failed: %v", err)
		}
	})

	t.Run("DryRun", func(t *testing.T) {
		cfg.DryRun = true
		if err := srv.Start(ctx); err != nil {
			t.Errorf("Start dry-run failed: %v", err)
		}
		if err := srv.Stop(ctx); err != nil {
			t.Errorf("Stop dry-run failed: %v", err)
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		if checks := srv.HealthCheck(ctx); len(checks) == 0 {
			t.Error("HealthCheck returned no checks")
		}
	})
}

// ============================================================================
// MOD TESTS
// ============================================================================

func TestMods(t *testing.T) {
	cfg, logger, ctx := setup(t)
	svc := service.NewMods(cfg, logger)

	t.Run("ListInstalled", func(t *testing.T) {
		_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "test.jar"), []byte("jar"), 0o644)
		mods, err := svc.ListInstalled()
		if err != nil || len(mods) != 1 {
			t.Errorf("ListInstalled failed: %v", err)
		}
	})

	t.Run("UpdateDryRun", func(t *testing.T) {
		cfg.DryRun = true
		result, err := svc.UpdateAll(ctx, false)
		if err != nil || result == nil {
			t.Errorf("UpdateAll dry-run failed: %v", err)
		}
	})
}

// ============================================================================
// BACKUP TESTS
// ============================================================================

func TestBackup(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	svc := service.NewBackup(cfg, logger)

	t.Run("Create", func(t *testing.T) {
		_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "data.txt"), []byte("data"), 0o644)
		path, err := svc.Create(ctx)
		if err != nil || path == "" {
			t.Fatalf("Create failed: %v", err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Backup file missing: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		backups, err := svc.List()
		if err != nil {
			t.Errorf("List failed: %v", err)
		}
		if len(backups) == 0 {
			t.Error("Expected at least one backup")
		}
	})
}

// ============================================================================
// NOTIFICATION TESTS
// ============================================================================

func TestNotification(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.DryRun = true
	cfg.Notifications.DiscordWebhook = "http://localhost"
	svc := service.NewNotification(cfg, logger)

	t.Run("Send", func(t *testing.T) {
		if err := svc.SendSuccess(ctx, "test"); err != nil {
			t.Errorf("SendSuccess failed: %v", err)
		}
		if err := svc.SendError(ctx, "test"); err != nil {
			t.Errorf("SendError failed: %v", err)
		}
	})

	t.Run("NoWebhook", func(t *testing.T) {
		cfg.Notifications.DiscordWebhook = ""
		if err := svc.SendSuccess(ctx, "test"); err != nil {
			t.Errorf("SendSuccess failed without webhook: %v", err)
		}
	})
}
