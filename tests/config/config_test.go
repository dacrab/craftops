package config_test

import (
	"craftops/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading default config
	cfg := config.DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Verify essential defaults
	if cfg.Server.JarName == "" {
		t.Error("Default server jar_name should not be empty")
	}

	if cfg.Paths.Server == "" {
		t.Error("Default server path should not be empty")
	}

	if cfg.Mods.ConcurrentDownloads <= 0 {
		t.Error("Default concurrent_downloads should be positive")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
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

	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config from file
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded values
	if cfg.Server.JarName != "test-server.jar" {
		t.Errorf("Expected jar_name 'test-server.jar', got '%s'", cfg.Server.JarName)
	}

	if cfg.Mods.ConcurrentDownloads != 5 {
		t.Errorf("Expected concurrent_downloads 5, got %d", cfg.Mods.ConcurrentDownloads)
	}

	if cfg.Paths.Server != "/test/server" {
		t.Errorf("Expected server path '/test/server', got '%s'", cfg.Paths.Server)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	_, err := config.LoadConfig("/nonexistent/config.toml")
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
}
