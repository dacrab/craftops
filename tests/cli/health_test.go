package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"craftops/internal/cli"
	"craftops/tests/mocks"
)

func TestHealthCheckCommand(t *testing.T) {
	factory := &mocks.MockServiceFactory{
		BackupService:       &mocks.MockBackupService{},
		ModService:          &mocks.MockModService{},
		NotificationService: &mocks.MockNotificationService{},
		ServerService:       &mocks.MockServerService{},
	}

	cmd := cli.NewRootCmd(factory)
	cmd.SetArgs([]string{"health-check"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute() failed with %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "System Health Check") {
		t.Errorf("Expected output to contain 'System Health Check', got %q", output)
	}
}
