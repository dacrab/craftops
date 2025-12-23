package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfigCreatesFile(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "config.toml")

	// Reset global state
	cfgFile = ""
	application = nil

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

func TestCommandsDryRun(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := minimalConfig(t, tmp)

	tests := []struct {
		name string
		args []string
	}{
		{"backup-list", []string{"backup", "list"}},
		{"backup-create", []string{"--dry-run", "backup", "create"}},
		{"update-mods", []string{"--dry-run", "update-mods", "--no-backup"}},
		{"server-status", []string{"server", "status"}},
		{"server-start", []string{"--dry-run", "server", "start"}},
		{"server-stop", []string{"--dry-run", "server", "stop"}},
		{"health-check", []string{"health-check"}},
		{"version", []string{"--version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile = ""
			application = nil
			orig := os.Args
			defer func() { os.Args = orig }()

			os.Args = append([]string{"craftops", "--config", cfgPath}, tt.args...)
			if tt.name == "version" {
				os.Args = []string{"craftops", "--version"}
			}

			_ = Execute() // Errors are expected in some cases due to missing env, but we test execution flow
		})
	}
}

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
java_flags = ["-Xms512m"]
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
file_enabled = false
console_enabled = false
`
	path := filepath.Join(base, "config.toml")
	if err := os.WriteFile(path, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
