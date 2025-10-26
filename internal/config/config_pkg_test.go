package config

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid defaults",
			mutate:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "invalid modloader",
			mutate: func(c *Config) {
				c.Minecraft.Modloader = "invalid-loader"
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			mutate: func(c *Config) {
				c.Logging.Format = "xml"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.mutate(cfg)
			err := cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
