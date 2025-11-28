package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	if cfg.Server.JarName == "" {
		t.Error("default jar_name should not be empty")
	}
	if cfg.Paths.Server == "" {
		t.Error("default server path should not be empty")
	}
	if cfg.Mods.ConcurrentDownloads <= 0 {
		t.Error("default concurrent_downloads should be positive")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
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

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Server.JarName != "test-server.jar" {
		t.Errorf("jar_name = %q, want %q", cfg.Server.JarName, "test-server.jar")
	}
	if cfg.Mods.ConcurrentDownloads != 5 {
		t.Errorf("concurrent_downloads = %d, want 5", cfg.Mods.ConcurrentDownloads)
	}
	if cfg.Paths.Server != "/test/server" {
		t.Errorf("server path = %q, want /test/server", cfg.Paths.Server)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	if _, err := LoadConfig("/nonexistent/config.toml"); err == nil {
		t.Error("expected error for non-existent config file")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"valid defaults", func(c *Config) {}, false},
		{"invalid modloader", func(c *Config) { c.Minecraft.Modloader = "invalid" }, true},
		{"invalid log format", func(c *Config) { c.Logging.Format = "xml" }, true},
		{"invalid log level", func(c *Config) { c.Logging.Level = "INVALID" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.mutate(cfg)
			err := cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved_config.toml")

	cfg := DefaultConfig()
	cfg.Server.JarName = "custom-server.jar"

	if err := cfg.SaveConfig(configPath); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.Server.JarName != "custom-server.jar" {
		t.Errorf("jar_name = %q, want custom-server.jar", loaded.Server.JarName)
	}
}
