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
	t.Setenv("HOME", t.TempDir())
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
		mutate  func(*Config)
		name    string
		wantErr bool
	}{
		{name: "valid defaults", mutate: func(_ *Config) {}, wantErr: false},
		{name: "modloader case insensitive", mutate: func(c *Config) { c.Minecraft.Modloader = "Fabric" }, wantErr: false},
		{name: "invalid modloader", mutate: func(c *Config) { c.Minecraft.Modloader = "badloader" }, wantErr: true},
		{name: "invalid log level", mutate: func(c *Config) { c.Logging.Level = "VERBOSE" }, wantErr: true},
		{name: "invalid log format", mutate: func(c *Config) { c.Logging.Format = "xml" }, wantErr: true},
		{name: "valid log level debug", mutate: func(c *Config) { c.Logging.Level = "debug" }, wantErr: false},
		{name: "valid format text", mutate: func(c *Config) { c.Logging.Format = "text" }, wantErr: false},
		{name: "negative concurrent downloads", mutate: func(c *Config) { c.Mods.ConcurrentDownloads = -1 }, wantErr: true},
		{name: "zero concurrent downloads", mutate: func(c *Config) { c.Mods.ConcurrentDownloads = 0 }, wantErr: true},
		{name: "zero mods timeout", mutate: func(c *Config) { c.Mods.Timeout = 0 }, wantErr: true},
		{name: "negative mods timeout", mutate: func(c *Config) { c.Mods.Timeout = -5 }, wantErr: true},
		{name: "zero max retries", mutate: func(c *Config) { c.Mods.MaxRetries = 0 }, wantErr: false},
		{name: "negative max retries", mutate: func(c *Config) { c.Mods.MaxRetries = -1 }, wantErr: true},
		{name: "negative retry delay", mutate: func(c *Config) { c.Mods.RetryDelay = -1 }, wantErr: true},
		{name: "zero retry delay", mutate: func(c *Config) { c.Mods.RetryDelay = 0 }, wantErr: false},
		{name: "negative max stop wait", mutate: func(c *Config) { c.Server.MaxStopWait = -1 }, wantErr: true},
		{name: "zero startup timeout", mutate: func(c *Config) { c.Server.StartupTimeout = 0 }, wantErr: false},
		{name: "negative startup timeout", mutate: func(c *Config) { c.Server.StartupTimeout = -1 }, wantErr: true},
		{name: "zero max backups unlimited", mutate: func(c *Config) { c.Backup.MaxBackups = 0 }, wantErr: false},
		{name: "negative max backups", mutate: func(c *Config) { c.Backup.MaxBackups = -1 }, wantErr: true},
		{name: "compression level below range", mutate: func(c *Config) { c.Backup.CompressionLevel = -1 }, wantErr: true},
		{name: "compression level above range", mutate: func(c *Config) { c.Backup.CompressionLevel = 10 }, wantErr: true},
		{name: "valid compression level", mutate: func(c *Config) { c.Backup.CompressionLevel = 9 }, wantErr: false},
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

func TestValidation_Normalizes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Minecraft.Modloader = "FABRIC"
	cfg.Logging.Level = "debug"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if cfg.Minecraft.Modloader != "fabric" {
		t.Errorf("modloader not normalized: got %q", cfg.Minecraft.Modloader)
	}
	if cfg.Logging.Level != "DEBUG" {
		t.Errorf("log level not normalized: got %q", cfg.Logging.Level)
	}
}

func TestSaveConfig_BadPath(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.SaveConfig("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error saving to nonexistent path")
	}
}

func TestLoadConfig_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.toml")
	cfg := DefaultConfig()
	_ = cfg.SaveConfig(cfgPath)

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig with explicit path: %v", err)
	}
	validModloaders := []string{"fabric", "forge", "quilt", "neoforge"}
	if !slices.Contains(validModloaders, loaded.Minecraft.Modloader) {
		t.Errorf("unexpected modloader after round-trip: %q", loaded.Minecraft.Modloader)
	}
	if loaded.Logging.Level != "INFO" {
		t.Errorf("expected log level INFO after round-trip, got %q", loaded.Logging.Level)
	}
}
