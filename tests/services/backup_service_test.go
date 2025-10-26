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

func TestBackupServiceVariants(t *testing.T) {
	t.Run("new service + list", func(t *testing.T) {
		cfg := config.DefaultConfig()
		logger := zap.NewNop()

		svc := services.NewBackupService(cfg, logger)
		if svc == nil {
			t.Fatal("NewBackupService returned nil")
		}
		backups, err := svc.ListBackups()
		if err != nil {
			t.Errorf("ListBackups error: %v", err)
		}
		if backups == nil {
			t.Error("ListBackups should return empty slice, not nil")
		}
	})

	t.Run("health check", func(t *testing.T) {
		cfg := config.DefaultConfig()
		logger := zap.NewNop()
		svc := services.NewBackupService(cfg, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checks := svc.HealthCheck(ctx)
		if len(checks) == 0 {
			t.Error("HealthCheck should return at least one check")
		}
		for _, check := range checks {
			if check.Name == "" {
				t.Error("health check name should not be empty")
			}
			if check.Status == "" {
				t.Error("health check status should not be empty")
			}
			if check.Message == "" {
				t.Error("health check message should not be empty")
			}
		}
	})

	t.Run("create backup (dry-run)", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.DryRun = true
		logger := zap.NewNop()
		svc := services.NewBackupService(cfg, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		path, err := svc.CreateBackup(ctx)
		if err != nil {
			t.Fatalf("CreateBackup(dry-run) error: %v", err)
		}
		if path != "dry-run-backup.tar.gz" {
			t.Errorf("unexpected dry-run path = %q", path)
		}
	})

	// Exercise real backup creation path to improve coverage
	t.Run("create backup (real)", func(t *testing.T) {
		tmp := t.TempDir()
		server := tmp + "/server"
		backups := tmp + "/backups"
		if err := os.MkdirAll(server+"/world", 0o755); err != nil { t.Fatal(err) }
		if err := os.MkdirAll(server+"/logs", 0o755); err != nil { t.Fatal(err) }
		if err := os.WriteFile(server+"/world/level.dat", []byte("x"), 0o644); err != nil { t.Fatal(err) }
		if err := os.WriteFile(server+"/logs/server.log", []byte("log"), 0o644); err != nil { t.Fatal(err) }

		cfg := config.DefaultConfig()
		cfg.DryRun = false
		cfg.Paths.Server = server
		cfg.Paths.Backups = backups
		cfg.Backup.Enabled = true
		cfg.Backup.IncludeLogs = false
		cfg.Backup.MaxBackups = 2

		logger := zap.NewNop()
		svc := services.NewBackupService(cfg, logger)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		path, err := svc.CreateBackup(ctx)
		if err != nil { t.Fatalf("CreateBackup error: %v", err) }
		if path == "" { t.Fatalf("CreateBackup returned empty path") }
		if info, err := os.Stat(path); err != nil || info.Size() == 0 {
			t.Fatalf("backup not created properly: err=%v size_ok=%v", err, err==nil)
		}

        // Create two more backups to trigger retention cleanup (MaxBackups=2)
        time.Sleep(1100 * time.Millisecond)
        if _, err := svc.CreateBackup(ctx); err != nil { t.Fatalf("second backup: %v", err) }
        time.Sleep(1100 * time.Millisecond)
        if _, err := svc.CreateBackup(ctx); err != nil { t.Fatalf("third backup: %v", err) }
		files, err := filepath.Glob(filepath.Join(backups, "*.tar.gz"))
		if err != nil { t.Fatalf("glob backups: %v", err) }
		if len(files) != 2 {
			t.Fatalf("retention not enforced: got %d backups, want 2", len(files))
		}

		// ListBackups should return newest first
		list, err := svc.ListBackups()
		if err != nil { t.Fatalf("ListBackups: %v", err) }
		if len(list) == 0 { t.Fatalf("ListBackups empty") }
		// the last created file should be first in the list
		newest := list[0].Path
		if _, err := os.Stat(newest); err != nil { t.Fatalf("newest path invalid: %v", err) }
	})
}
