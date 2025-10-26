package cli

import (
    "os"
    "path/filepath"
    "testing"

    "craftops/internal/config"
)

func TestInitConfigCreatesFile(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "config.toml")
    // Ensure global CLI state doesn't carry over between tests
    cfgFile = ""
	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "init-config", "-o", out, "--force"}
	if err := Execute(); err != nil {
		t.Fatalf("Execute(init-config) error: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("config not created: %v", err)
	}
}

func TestBackupListWithTempConfig(t *testing.T) {
	tmp := t.TempDir()
	cfg := minimalConfig(t, tmp)
	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "--config", cfg, "backup", "list"}
	if err := Execute(); err != nil {
		t.Fatalf("Execute(backup list) error: %v", err)
	}

    // Also exercise backup create (dry-run) and update-mods (dry-run, no-backup)
    os.Args = []string{"craftops", "--config", cfg, "--dry-run", "backup", "create"}
    if err := Execute(); err != nil {
        t.Fatalf("Execute(backup create dry-run) error: %v", err)
    }
    os.Args = []string{"craftops", "--config", cfg, "--dry-run", "update-mods", "--no-backup"}
    if err := Execute(); err != nil {
        t.Fatalf("Execute(update-mods dry-run) error: %v", err)
    }

    // Exercise health-check to cover checkPaths/displayHealthResults/printTable
    os.Args = []string{"craftops", "--config", cfg, "health-check"}
    if err := Execute(); err == nil {
        // Expect error due to missing external deps (java/screen/server.jar)
        t.Fatalf("expected health-check to return error when environment is missing deps")
    }

    // Also hit update-mods without --no-backup to exercise createBackupIfNeeded path
    os.Args = []string{"craftops", "--config", cfg, "--dry-run", "update-mods"}
    if err := Execute(); err != nil {
        t.Fatalf("Execute(update-mods with backup, dry-run) error: %v", err)
    }
}

func TestCheckPathsMixedStates(t *testing.T) {
    // Directly exercise checkPaths on a crafted cfg with a file instead of dir
    tmp := t.TempDir()
    filePath := filepath.Join(tmp, "notadir")
    if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil { t.Fatal(err) }
    cfg = config.DefaultConfig()
    cfg.Paths.Server = filePath
    cfg.Paths.Mods = filepath.Join(tmp, "mods")
    cfg.Paths.Backups = filepath.Join(tmp, "backups")
    cfg.Paths.Logs = filepath.Join(tmp, "logs")
    checks := checkPaths()
    if len(checks) != 4 { t.Fatalf("want 4 checks, got %d", len(checks)) }
}

func TestServerStatusNoDeps(t *testing.T) {
	tmp := t.TempDir()
	cfg := minimalConfig(t, tmp)
	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "--config", cfg, "server", "status"}
	if err := Execute(); err != nil {
		t.Fatalf("Execute(server status) error: %v", err)
	}
}

// minimalConfig writes a minimal TOML config to base and returns its path.
func minimalConfig(t *testing.T, base string) string {
	t.Helper()
	server := filepath.Join(base, "server")
	mods := filepath.Join(base, "mods")
	backups := filepath.Join(base, "backups")
	logs := filepath.Join(base, "logs")
	for _, p := range []string{server, mods, backups, logs} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}
	cfg := `
[minecraft]
version = "1.20.1"
modloader = "fabric"

[paths]
server = "` + server + `"
mods = "` + mods + `"
backups = "` + backups + `"
logs = "` + logs + `"

[server]
jar_name = "server.jar"
java_flags = ["-Xms512m", "-Xmx512m"]
stop_command = "stop"
max_stop_wait = 5
startup_timeout = 2
session_name = "minecraft"

[mods]
concurrent_downloads = 2
max_retries = 0
retry_delay = 0.1
timeout = 1
modrinth_sources = []

[backup]
enabled = true
max_backups = 1
compression_level = 1
include_logs = false
exclude_patterns = ["*.log"]

[notifications]
discord_webhook = ""
warning_intervals = [1]
warning_message = "restart in {minutes}"
success_notifications = false
error_notifications = false

[logging]
level = "INFO"
format = "json"
file_enabled = true
console_enabled = false
`
	path := filepath.Join(base, "config.toml")
	if err := os.WriteFile(path, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
