package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/domain"
	"craftops/internal/service"
)

// setup creates a fully isolated test environment with temp directories.
func setup(t *testing.T) (*config.Config, *zap.Logger, context.Context) {
	t.Helper()
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Server = filepath.Join(tmp, "server")
	cfg.Paths.Mods = filepath.Join(tmp, "mods")
	cfg.Paths.Backups = filepath.Join(tmp, "backups")
	cfg.Paths.Logs = filepath.Join(tmp, "logs")

	for _, p := range []string{cfg.Paths.Server, cfg.Paths.Mods, cfg.Paths.Backups, cfg.Paths.Logs} {
		if err := os.MkdirAll(p, 0o750); err != nil {
			t.Fatalf("setup: MkdirAll(%s): %v", p, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	return cfg, zap.NewNop(), ctx
}

// ── Backup ────────────────────────────────────────────────────────────────────

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
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	cfg.Backup.MaxBackups = 10
	svc := service.NewBackup(cfg, logger)

	// Create two backups with a small sleep to ensure distinct timestamps
	_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "a.txt"), []byte("a"), 0o600)
	path1, _ := svc.Create(ctx)
	time.Sleep(1100 * time.Millisecond) // ensure different mod times
	_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "b.txt"), []byte("b"), 0o600)
	path2, _ := svc.Create(ctx)

	backups, err := svc.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(backups) < 2 {
		t.Fatalf("expected at least 2 backups, got %d", len(backups))
	}
	// Newest first — path2 should appear before path1
	if backups[0].Path != path2 {
		t.Errorf("expected newest backup first: got %s, want %s", backups[0].Path, path2)
	}
	_ = path1
}

func TestBackup_Retention(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Backup.Enabled = true
	cfg.Backup.MaxBackups = 2
	svc := service.NewBackup(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Server, "x.txt"), []byte("x"), 0o600)
	for range 3 {
		time.Sleep(1100 * time.Millisecond)
		if _, err := svc.Create(ctx); err != nil {
			t.Fatalf("Create failed: %v", err)
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
}

// ── Mods ─────────────────────────────────────────────────────────────────────

func TestMods_ListInstalled_Empty(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewMods(cfg, logger)

	mods, err := svc.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled error: %v", err)
	}
	if len(mods) != 0 {
		t.Errorf("expected 0 mods, got %d", len(mods))
	}
}

func TestMods_ListInstalled(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewMods(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "fabric-api.jar"), []byte("jar"), 0o600)
	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "sodium.jar"), []byte("jar"), 0o600)
	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "readme.txt"), []byte("text"), 0o600) // should be ignored

	mods, err := svc.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled error: %v", err)
	}
	if len(mods) != 2 {
		t.Errorf("expected 2 mods, got %d", len(mods))
	}
	for _, m := range mods {
		if !strings.HasSuffix(m.Filename, ".jar") {
			t.Errorf("non-jar file returned: %s", m.Filename)
		}
		if m.Name == "" {
			t.Error("mod name should not be empty")
		}
	}
}

func TestMods_ParseProjectID(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewMods(cfg, logger)

	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"fabric-api", "fabric-api", false},
		{"https://modrinth.com/mod/fabric-api", "fabric-api", false},
		{"https://modrinth.com/mod/sodium/versions", "sodium", false},
		{"https://invalid.com/notamod", "", true},
	}
	for _, tt := range tests {
		got, err := svc.ParseProjectID(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseProjectID(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseProjectID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMods_UpdateAll_NoSources(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Mods.ModrinthSources = []string{}
	svc := service.NewMods(cfg, logger)

	result, err := svc.UpdateAll(ctx, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.UpdatedMods) != 0 || len(result.FailedMods) != 0 || len(result.SkippedMods) != 0 {
		t.Error("expected empty results with no sources")
	}
}

func TestMods_HealthCheck(t *testing.T) {
	cfg, logger, ctx := setup(t)
	svc := service.NewMods(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if len(checks) < 2 {
		t.Fatalf("expected at least 2 health checks, got %d", len(checks))
	}
	// Mods dir check
	found := false
	for _, c := range checks {
		if strings.Contains(c.Name, "Mods") || strings.Contains(c.Name, "directory") {
			found = true
		}
	}
	if !found {
		t.Error("expected a mods directory health check")
	}
}

// ── Notification ──────────────────────────────────────────────────────────────

func TestNotification_HealthCheck_NoWebhook(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.DiscordWebhook = ""
	svc := service.NewNotification(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if len(checks) < 1 {
		t.Fatal("expected at least one health check")
	}
	webhookCheck := checks[0]
	if webhookCheck.Status != domain.StatusWarn {
		t.Errorf("expected WARN for missing webhook, got %s", webhookCheck.Status)
	}
}

func TestNotification_HealthCheck_InvalidWebhook(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.DiscordWebhook = "https://invalid.example.com/webhook"
	svc := service.NewNotification(cfg, logger)

	checks := svc.HealthCheck(ctx)
	webhookCheck := checks[0]
	if webhookCheck.Status != domain.StatusError {
		t.Errorf("expected ERROR for invalid webhook URL, got %s", webhookCheck.Status)
	}
}

func TestNotification_HealthCheck_ValidWebhook(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.DiscordWebhook = "https://discord.com/api/webhooks/123/abc"
	svc := service.NewNotification(cfg, logger)

	checks := svc.HealthCheck(ctx)
	webhookCheck := checks[0]
	if webhookCheck.Status != domain.StatusOK {
		t.Errorf("expected OK for valid webhook URL, got %s: %s", webhookCheck.Status, webhookCheck.Message)
	}
}

func TestNotification_HealthCheck_AllDisabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.SuccessNotifications = false
	cfg.Notifications.ErrorNotifications = false
	svc := service.NewNotification(cfg, logger)

	checks := svc.HealthCheck(ctx)
	if len(checks) < 2 {
		t.Fatalf("expected 2 checks, got %d", len(checks))
	}
	settingsCheck := checks[1]
	if settingsCheck.Status != domain.StatusWarn {
		t.Errorf("expected WARN when all notifications disabled, got %s", settingsCheck.Status)
	}
}

func TestNotification_SendSuccess_Disabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.SuccessNotifications = false
	svc := service.NewNotification(cfg, logger)

	// Should be a no-op, no error
	if err := svc.SendSuccess(ctx, "hello"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestNotification_SendError_Disabled(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.ErrorNotifications = false
	svc := service.NewNotification(cfg, logger)

	if err := svc.SendError(ctx, "boom"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestNotification_DryRun(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.DryRun = true
	cfg.Notifications.DiscordWebhook = "https://discord.com/api/webhooks/123/abc"
	cfg.Notifications.SuccessNotifications = true
	svc := service.NewNotification(cfg, logger)

	// Should not make any real HTTP calls in dry-run
	if err := svc.SendSuccess(ctx, "test"); err != nil {
		t.Errorf("dry-run SendSuccess failed: %v", err)
	}
}

func TestNotification_SendRestartWarnings_Empty(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Notifications.WarningIntervals = []int{}
	svc := service.NewNotification(cfg, logger)

	if err := svc.SendRestartWarnings(ctx); err != nil {
		t.Errorf("expected nil with empty intervals, got %v", err)
	}
}

func TestNotification_SendRestartWarnings_NoWebhook(t *testing.T) {
	cfg, logger, ctx := setup(t)
	// With no webhook configured, each sendDiscord call is a no-op.
	// Use a single interval so no inter-warning sleep occurs.
	cfg.Notifications.DiscordWebhook = ""
	cfg.Notifications.WarningIntervals = []int{5}
	svc := service.NewNotification(cfg, logger)

	if err := svc.SendRestartWarnings(ctx); err != nil {
		t.Errorf("expected nil when no webhook configured, got %v", err)
	}
}

// ── Server ────────────────────────────────────────────────────────────────────

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

func TestServer_Status_DryRun(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.DryRun = true
	svc := service.NewServer(cfg, logger)

	// Status should always work (reads screen output)
	status, err := svc.Status(ctx)
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if status == nil {
		t.Fatal("Status() returned nil")
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
