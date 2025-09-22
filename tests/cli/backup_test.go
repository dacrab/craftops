package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"craftops/internal/cli"
	"craftops/tests/mocks"
)

func TestBackupCreateCommand(t *testing.T) {
	factory := &mocks.MockServiceFactory{
		BackupService: &mocks.MockBackupService{},
	}

	cmd := cli.NewRootCmd(factory)
	cmd.SetArgs([]string{"backup", "create"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute() failed with %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Backup created") {
		t.Errorf("Expected output to contain 'Backup created', got %q", output)
	}
}

func TestBackupListCommand(t *testing.T) {
	factory := &mocks.MockServiceFactory{
		BackupService: &mocks.MockBackupService{},
	}

	cmd := cli.NewRootCmd(factory)
	cmd.SetArgs([]string{"backup", "list"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute() failed with %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "mock_backup.tar.gz") {
		t.Errorf("Expected output to contain 'mock_backup.tar.gz', got %q", output)
	}
}
