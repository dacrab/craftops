package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"craftops/internal/config"
)

func TestLoadConfigVariants(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cfg := config.DefaultConfig()
		if cfg == nil {
			t.Fatal("DefaultConfig() returned nil")
		}
		if cfg.Server.JarName == "" {
			t.Error("default server jar_name should not be empty")
		}
		if cfg.Paths.Server == "" {
			t.Error("default server path should not be empty")
		}
		if cfg.Mods.ConcurrentDownloads <= 0 {
			t.Error("default concurrent_downloads should be positive")
		}
	})

	t.Run("from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test_config.toml")

		testConfig := `
[server]
jar_name = "test-server.jar"
java_flags = ["-Xmx4G", "-Xms2G"]

[paths]
server = "/test/server"
mods = "/test/mods"

[mods]
concurrent_downloads = 5
`
		if err := os.WriteFile(configPath, []byte(testConfig), 0o644); err != nil {
			t.Fatalf("write test config: %v", err)
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if got, want := cfg.Server.JarName, "test-server.jar"; got != want {
			t.Errorf("jar_name = %q, want %q", got, want)
		}
		if got, want := cfg.Mods.ConcurrentDownloads, 5; got != want {
			t.Errorf("concurrent_downloads = %d, want %d", got, want)
		}
		if got, want := cfg.Paths.Server, "/test/server"; got != want {
			t.Errorf("server path = %q, want %q", got, want)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		if _, err := config.LoadConfig("/nonexistent/config.toml"); err == nil {
			t.Error("expected error for non-existent config file")
		}
	})
}
