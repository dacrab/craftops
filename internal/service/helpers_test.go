package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

const testDiscordWebhook = "https://discord.com/api/webhooks/123/abc"

func setup(t *testing.T) (*config.Config, *zap.Logger, context.Context) {
	t.Helper()
	cfg := config.DefaultConfig()
	tmp := t.TempDir()
	cfg.Paths.Server = filepath.Join(tmp, "server")
	cfg.Paths.Mods = filepath.Join(tmp, "mods")
	cfg.Paths.Backups = filepath.Join(tmp, "backups")
	cfg.Paths.Logs = filepath.Join(tmp, "logs")

	for _, p := range []string{cfg.Paths.Server, cfg.Paths.Mods, cfg.Paths.Backups, cfg.Paths.Logs} {
		if err := os.MkdirAll(p, 0o750); err != nil {
			t.Fatalf("setup: MkdirAll(%s): %v", p, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	return cfg, zap.NewNop(), ctx
}
