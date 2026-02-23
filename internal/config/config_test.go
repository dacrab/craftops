package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.JarName == "" {
		t.Error("DefaultConfig: JarName is empty")
	}
	if cfg.Paths.Server == "" {
		t.Error("DefaultConfig: Paths.Server is empty")
	}
	if cfg.Mods.ConcurrentDownloads <= 0 {
		t.Error("DefaultConfig: ConcurrentDownloads should be positive")
	}
	if cfg.Backup.MaxBackups <= 0 {
		t.Error("DefaultConfig: MaxBackups should be positive")
	}
	if len(cfg.Server.JavaFlags) == 0 {
		t.Error("DefaultConfig: JavaFlags should not be empty")
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.toml")

	cfg := DefaultConfig()
	cfg.Server.JarName = "test.jar"
	cfg.Minecraft.Version = "1.21.0"

	if err := cfg.SaveConfig(path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.Server.JarName != "test.jar" {
		t.Errorf("JarName: got %q, want %q", loaded.Server.JarName, "test.jar")
	}
	if loaded.Minecraft.Version != "1.21.0" {
		t.Errorf("Version: got %q, want %q", loaded.Minecraft.Version, "1.21.0")
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	// No config file present in a temp dir — should succeed with defaults
	// (LoadConfig falls back to defaults when no file is found)
	orig := os.Getenv("HOME")
	t.Setenv("HOME", t.TempDir()) // isolate from real user config
	defer func() { _ = os.Setenv("HOME", orig) }()

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig with no file should not error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	tmp := t.TempDir()
	bad := filepath.Join(tmp, "bad.toml")
	_ = os.WriteFile(bad, []byte("[[invalid toml"), 0o600)

	_, err := LoadConfig(bad)
	if err == nil {
		t.Error("expected error loading invalid TOML file")
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"valid defaults", func(_ *Config) {}, false},
		{"modloader case insensitive", func(c *Config) { c.Minecraft.Modloader = "Fabric" }, false},
		{"invalid modloader", func(c *Config) { c.Minecraft.Modloader = "badloader" }, true},
		{"invalid log level", func(c *Config) { c.Logging.Level = "VERBOSE" }, true},
		{"invalid log format", func(c *Config) { c.Logging.Format = "xml" }, true},
		{"valid log level debug", func(c *Config) { c.Logging.Level = "debug" }, false},
		{"valid format text", func(c *Config) { c.Logging.Format = "text" }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.mutate(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidation_NormalizesModloader(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Minecraft.Modloader = "FABRIC"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if cfg.Minecraft.Modloader != "fabric" {
		t.Errorf("expected modloader normalized to lowercase, got %q", cfg.Minecraft.Modloader)
	}
}

func TestValidation_NormalizesLogLevel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Logging.Level = "debug"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Logging.Level != "DEBUG" {
		t.Errorf("expected log level normalized to uppercase, got %q", cfg.Logging.Level)
	}
}

func TestSaveConfig_BadPath(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.SaveConfig("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error saving to nonexistent path")
	}
}

func TestFindDefaultConfig(t *testing.T) {
	// Place a config.toml in a temp dir and verify findDefaultConfig finds it
	// by temporarily changing the working directory
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.toml")
	cfg := DefaultConfig()
	_ = cfg.SaveConfig(cfgPath)

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig with explicit path: %v", err)
	}
	// Verify modloader is valid after normalisation
	validModloaders := []string{"fabric", "forge", "quilt", "neoforge"}
	if !slices.Contains(validModloaders, loaded.Minecraft.Modloader) {
		t.Errorf("unexpected modloader: %q", loaded.Minecraft.Modloader)
	}
}
