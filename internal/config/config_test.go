package config

import (
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		cfg := DefaultConfig()
		if cfg.Server.JarName == "" || cfg.Paths.Server == "" {
			t.Error("DefaultConfig missing critical values")
		}
	})

	t.Run("LoadAndSave", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "config.toml")

		cfg := DefaultConfig()
		cfg.Server.JarName = "test.jar"
		if err := cfg.SaveConfig(path); err != nil {
			t.Fatalf("SaveConfig failed: %v", err)
		}

		loaded, err := LoadConfig(path)
		if err != nil || loaded.Server.JarName != "test.jar" {
			t.Fatalf("LoadConfig failed or data mismatch: %v", err)
		}
	})

	t.Run("Validation", func(t *testing.T) {
		tests := []struct {
			name    string
			mutate  func(*Config)
			wantErr bool
		}{
			{"valid", func(_ *Config) {}, false},
			{"invalid-modloader", func(c *Config) { c.Minecraft.Modloader = "invalid" }, true},
			{"invalid-level", func(c *Config) { c.Logging.Level = "INVALID" }, true},
			{"invalid-format", func(c *Config) { c.Logging.Format = "xml" }, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cfg := DefaultConfig()
				tt.mutate(cfg)
				if err := cfg.Validate(); (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}
