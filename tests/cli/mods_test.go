package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"craftops/internal/cli"
	"craftops/tests/mocks"
)

func TestUpdateModsCommand(t *testing.T) {
	factory := &mocks.MockServiceFactory{
		BackupService: &mocks.MockBackupService{},
		ModService:    &mocks.MockModService{},
	}

	cmd := cli.NewRootCmd(factory)
	cmd.SetArgs([]string{"update-mods"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute() failed with %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Successfully updated") {
		t.Errorf("Expected output to contain 'Successfully updated', got %q", output)
	}
}
