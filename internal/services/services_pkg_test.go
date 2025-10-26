package services

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

func TestServerService_StatusDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DryRun = true
	s := NewServerService(cfg, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("start (dry run) error: %v", err)
	}
	if err := s.Stop(ctx); err != nil {
		t.Fatalf("stop (dry run) error: %v", err)
	}
	if _, err := s.GetStatus(ctx); err != nil {
		t.Fatalf("get status error: %v", err)
	}
}

func TestBackupService_ListBackupsEmpty(t *testing.T) {
	cfg := config.DefaultConfig()
	b := NewBackupService(cfg, zap.NewNop())
	if list, err := b.ListBackups(); err != nil || list == nil {
		t.Fatalf("list backups err=%v nil=%v", err, list == nil)
	}
}

func TestNotificationService_NoWebhook(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Notifications.DiscordWebhook = ""
	n := NewNotificationService(cfg, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := n.SendSuccessNotification(ctx, "test"); err != nil {
		t.Fatalf("unexpected error without webhook: %v", err)
	}
}
