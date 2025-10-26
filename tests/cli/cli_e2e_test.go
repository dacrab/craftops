package cli_e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"craftops/internal/cli"
)

// writeTempConfig creates a minimal config pointing to temp dirs.
func writeTempConfig(t *testing.T, base string) string {
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
debug = false
dry_run = false

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
timeout = 2
modrinth_sources = ["https://modrinth.com/mod/example"]

[backup]
enabled = true
max_backups = 2
compression_level = 1
include_logs = false
exclude_patterns = ["*.log"]

[notifications]
discord_webhook = ""
warning_intervals = [15,10,5,1]
warning_message = "Server will restart in {minutes} minute(s)"
success_notifications = false
error_notifications = false

[logging]
level = "INFO"
format = "json"
file_enabled = false
console_enabled = false
`

	cfgPath := filepath.Join(base, "config.toml")
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return cfgPath
}

func runCLI(t *testing.T, args ...string) error {
	t.Helper()
	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = append([]string{"craftops"}, args...)
	return cli.Execute()
}

func TestCLI_InitConfig_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "out-config.toml")
	if err := runCLI(t, "init-config", "-o", out, "--force"); err != nil {
		t.Fatalf("init-config failed: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("config not created: %v", err)
	}
}

func TestCLI_BackupCreate_DryRun(t *testing.T) {
	tmp := t.TempDir()
	cfg := writeTempConfig(t, tmp)
	if err := runCLI(t, "--dry-run", "--config", cfg, "backup", "create"); err != nil {
		t.Fatalf("backup create (dry-run) failed: %v", err)
	}
}

func TestCLI_ServerLifecycle_DryRun(t *testing.T) {
	tmp := t.TempDir()
	cfg := writeTempConfig(t, tmp)

	if err := runCLI(t, "--dry-run", "--config", cfg, "server", "start"); err != nil {
		t.Fatalf("server start (dry-run) failed: %v", err)
	}
	if err := runCLI(t, "--dry-run", "--config", cfg, "server", "status"); err != nil {
		t.Fatalf("server status failed: %v", err)
	}
	if err := runCLI(t, "--dry-run", "--config", cfg, "server", "stop"); err != nil {
		t.Fatalf("server stop (dry-run) failed: %v", err)
	}
}

func TestCLI_UpdateMods_DryRun_NoBackup(t *testing.T) {
	tmp := t.TempDir()
	cfg := writeTempConfig(t, tmp)
	if err := runCLI(t, "--dry-run", "--config", cfg, "update-mods", "--no-backup"); err != nil {
		t.Fatalf("update-mods (dry-run) failed: %v", err)
	}
}
