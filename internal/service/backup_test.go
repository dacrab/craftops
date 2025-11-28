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

func TestNewBackup(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	svc := service.NewBackup(cfg, logger)
	if svc == nil {
		t.Fatal("NewBackup returned nil")
	}
}

func TestBackupCreateDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DryRun = true
	cfg.Backup.Enabled = true
	logger := zap.NewNop()
	svc := service.NewBackup(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	path, err := svc.Create(ctx)
	if err != nil {
		t.Fatalf("Create dry-run failed: %v", err)
	}
	if path == "" {
		t.Error("Create should return path in dry-run")
	}
}

func TestBackupCreateDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Backup.Enabled = false
	logger := zap.NewNop()
	svc := service.NewBackup(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := svc.Create(ctx)
	if err != domain.ErrBackupsDisabled {
		t.Errorf("expected ErrBackupsDisabled, got: %v", err)
	}
}

func TestBackupCreate(t *testing.T) {
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Server = filepath.Join(tmp, "server")
	cfg.Paths.Backups = filepath.Join(tmp, "backups")
	cfg.Backup.Enabled = true
	cfg.Backup.CompressionLevel = 1

	// Create server dir with some content
	if err := os.MkdirAll(cfg.Paths.Server, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.Paths.Server, "test.txt"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	logger := zap.NewNop()
	svc := service.NewBackup(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	path, err := svc.Create(ctx)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if path == "" {
		t.Error("Create should return backup path")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Backup file should exist: %v", err)
	}
}

func TestBackupList(t *testing.T) {
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Backups = tmp

	// Create a test backup file
	testBackup := filepath.Join(tmp, "test_backup.tar.gz")
	if err := os.WriteFile(testBackup, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	logger := zap.NewNop()
	svc := service.NewBackup(cfg, logger)

	backups, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("expected 1 backup, got %d", len(backups))
	}
}

func TestBackupHealthCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Backups = tmp
	cfg.Backup.Enabled = true

	logger := zap.NewNop()
	svc := service.NewBackup(cfg, logger)

	ctx := context.Background()
	checks := svc.HealthCheck(ctx)
	if len(checks) == 0 {
		t.Error("HealthCheck should return checks")
	}
}

func TestBackupInterface(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	var _ service.BackupManager = service.NewBackup(cfg, logger)
}
