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
		_ = os.MkdirAll(p, 0o750) //nolint:gosec // test directory permissions
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	return cfg, zap.NewNop(), ctx
}

func TestBackup_Create(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	svc := service.NewBackup(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "data.txt"), []byte("data"), 0o600) //nolint:gosec // test file permissions
	path, err := svc.Create(ctx)
	if err != nil || path == "" {
		t.Fatalf("Create failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Backup file missing: %v", err)
	}
}

func TestMods_ListInstalled(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewMods(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "test.jar"), []byte("jar"), 0o600) //nolint:gosec // test file permissions
	mods, err := svc.ListInstalled()
	if err != nil || len(mods) != 1 {
		t.Errorf("ListInstalled failed: %v", err)
	}
}
