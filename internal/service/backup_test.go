package service_test

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"craftops/internal/domain"
	"craftops/internal/service"
)

func TestBackup_Create(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	svc := service.NewBackup(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "data.txt"), []byte("data"), 0o600)
	path, err := svc.Create(ctx)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if path == "" {
		t.Fatal("Create returned empty path")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Backup file missing: %v", err)
	}
}

func TestBackup_Create_Disabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = false
	svc := service.NewBackup(cfg, logger)

	_, err := svc.Create(ctx)
	if !errors.Is(err, domain.ErrBackupsDisabled) {
		t.Errorf("expected ErrBackupsDisabled, got %v", err)
	}
}

func TestBackup_Create_DryRun(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	cfg.DryRun = true
	svc := service.NewBackup(cfg, logger)

	path, err := svc.Create(ctx)
	if err != nil {
		t.Fatalf("unexpected error in dry-run: %v", err)
	}
	if path == "" {
		t.Error("dry-run should return a non-empty path placeholder")
	}
}

func TestBackup_List_Empty(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewBackup(cfg, logger)

	backups, err := svc.List()
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}
}

func TestBackup_List_SortedNewestFirst(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewBackup(cfg, logger)

	now := time.Now()
	older := filepath.Join(cfg.Paths.Backups, "minecraft_backup_20000101_000001.tar.gz")
	newer := filepath.Join(cfg.Paths.Backups, "minecraft_backup_20000101_000002.tar.gz")
	_ = os.WriteFile(older, []byte("old"), 0o600)
	_ = os.WriteFile(newer, []byte("new"), 0o600)
	_ = os.Chtimes(older, now.Add(-2*time.Second), now.Add(-2*time.Second))
	_ = os.Chtimes(newer, now, now)

	backups, err := svc.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(backups) < 2 {
		t.Fatalf("expected at least 2 backups, got %d", len(backups))
	}
	if backups[0].Path != newer {
		t.Errorf("expected newest backup first: got %s, want %s", backups[0].Path, newer)
	}
}

func TestBackup_Retention(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	cfg.Backup.MaxBackups = 2
	svc := service.NewBackup(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "x.txt"), []byte("x"), 0o600)
	for i := range 3 {
		p, err := svc.Create(ctx)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if p != "" {
			ts := time.Now().Add(time.Duration(i-3) * time.Second)
			_ = os.Chtimes(p, ts, ts)
		}
	}

	backups, err := svc.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(backups) > cfg.Backup.MaxBackups {
		t.Errorf("retention: expected max %d backups, got %d", cfg.Backup.MaxBackups, len(backups))
	}
}

func TestBackup_HealthCheck_Disabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = false
	svc := service.NewBackup(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if len(checks) == 0 {
		t.Fatal("expected at least one health check")
	}
	if checks[0].Status != domain.StatusWarn {
		t.Errorf("expected WARN when disabled, got %s", checks[0].Status)
	}
}

func TestBackup_HealthCheck_Enabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	svc := service.NewBackup(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if len(checks) < 2 {
		t.Fatalf("expected at least 2 checks, got %d", len(checks))
	}
	names := make(map[string]domain.HealthStatus)
	for _, c := range checks {
		names[c.Name] = c.Status
	}
	if _, ok := names["Backup directory"]; !ok {
		t.Error("expected 'Backup directory' health check")
	}
	if _, ok := names["Backup retention"]; !ok {
		t.Error("expected 'Backup retention' health check")
	}
}

func TestBackup_Create_InvalidServerDir(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	cfg.Paths.Server = filepath.Join(t.TempDir(), "nonexistent")
	svc := service.NewBackup(cfg, logger)

	_, err := svc.Create(ctx)
	if err == nil {
		t.Error("expected error when server directory does not exist")
	}
}

func TestBackup_List_IgnoresNonTarGz(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewBackup(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Backups, "minecraft_backup_20000101_000001.tar.gz"), []byte("real"), 0o600)
	_ = os.WriteFile(filepath.Join(cfg.Paths.Backups, "readme.txt"), []byte("ignore"), 0o600)
	_ = os.WriteFile(filepath.Join(cfg.Paths.Backups, "backup.zip"), []byte("ignore"), 0o600)
	_ = os.Mkdir(filepath.Join(cfg.Paths.Backups, "subdir"), 0o750)

	backups, err := svc.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("expected exactly 1 backup (only .tar.gz), got %d", len(backups))
	}
}

func TestBackup_ExcludePatterns(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	cfg.Backup.ExcludePatterns = []string{"*.log"}
	cfg.Backup.IncludeLogs = false
	svc := service.NewBackup(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "server.log"), []byte("log data"), 0o600)
	_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "data.txt"), []byte("data"), 0o600)

	path, err := svc.Create(ctx)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer f.Close() //nolint:errcheck

	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	tr := tar.NewReader(gz)

	var found []string
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		found = append(found, hdr.Name)
	}

	for _, name := range found {
		if strings.HasSuffix(name, ".log") {
			t.Errorf("excluded .log file found in archive: %s", name)
		}
	}
	hasData := false
	for _, name := range found {
		if strings.Contains(name, "data.txt") {
			hasData = true
		}
	}
	if !hasData {
		t.Error("data.txt should be present in archive")
	}
}
