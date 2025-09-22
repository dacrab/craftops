package cli_test

import (
	"testing"

	"go.uber.org/zap"

	"craftops/internal/cli"
	"craftops/internal/config"
)

func TestServiceFactory(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	factory := cli.NewServiceFactory(cfg, logger)

	if factory == nil {
		t.Fatal("NewServiceFactory should not return nil")
	}

	if factory.GetBackupService() == nil {
		t.Error("GetBackupService should not return nil")
	}

	if factory.GetModService() == nil {
		t.Error("GetModService should not return nil")
	}

	if factory.GetNotificationService() == nil {
		t.Error("GetNotificationService should not return nil")
	}

	if factory.GetServerService() == nil {
		t.Error("GetServerService should not return nil")
	}
}
