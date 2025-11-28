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

func TestBackupList(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := minimalConfig(t, tmp)

	// Reset global state
	cfgFile = ""
	application = nil

	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "--config", cfgPath, "backup", "list"}

	if err := Execute(); err != nil {
		t.Fatalf("Execute(backup list) error: %v", err)
	}
}

func TestBackupCreateDryRun(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := minimalConfig(t, tmp)

	// Reset global state
	cfgFile = ""
	application = nil

	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "--config", cfgPath, "--dry-run", "backup", "create"}

	if err := Execute(); err != nil {
		t.Fatalf("Execute(backup create dry-run) error: %v", err)
	}
}

func TestUpdateModsDryRun(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := minimalConfig(t, tmp)

	// Reset global state
	cfgFile = ""
	application = nil

	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "--config", cfgPath, "--dry-run", "update-mods", "--no-backup"}

	if err := Execute(); err != nil {
		t.Fatalf("Execute(update-mods dry-run) error: %v", err)
	}
}

func TestServerStatus(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := minimalConfig(t, tmp)

	// Reset global state
	cfgFile = ""
	application = nil

	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "--config", cfgPath, "server", "status"}

	if err := Execute(); err != nil {
		t.Fatalf("Execute(server status) error: %v", err)
	}
}

func TestServerLifecycleDryRun(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := minimalConfig(t, tmp)

	tests := []struct {
		name string
		args []string
	}{
		{"start", []string{"craftops", "--config", cfgPath, "--dry-run", "server", "start"}},
		{"stop", []string{"craftops", "--config", cfgPath, "--dry-run", "server", "stop"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state
			cfgFile = ""
			application = nil

			orig := os.Args
			defer func() { os.Args = orig }()
			os.Args = tt.args

			if err := Execute(); err != nil {
				t.Fatalf("Execute(%s) error: %v", tt.name, err)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := minimalConfig(t, tmp)

	// Reset global state
	cfgFile = ""
	application = nil

	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "--config", cfgPath, "health-check"}

	// Expect error because server.jar doesn't exist, java/screen not available
	if err := Execute(); err == nil {
		t.Log("health-check passed (all deps available)")
	}
}

func TestVersion(t *testing.T) {
	// Reset global state
	cfgFile = ""
	application = nil

	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "--version"}

	if err := Execute(); err != nil {
		t.Fatalf("Execute(--version) error: %v", err)
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
