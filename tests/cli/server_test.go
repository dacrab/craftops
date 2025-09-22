package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"craftops/internal/cli"
	"craftops/tests/mocks"
)

func TestServerStartCommand(t *testing.T) {
	factory := &mocks.MockServiceFactory{
		ServerService: &mocks.MockServerService{},
	}

	cmd := cli.NewRootCmd(factory)
	cmd.SetArgs([]string{"server", "start"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute() failed with %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Server is now running") {
		t.Errorf("Expected output to contain 'Server is now running', got %q", output)
	}
}

func TestServerStopCommand(t *testing.T) {
	factory := &mocks.MockServiceFactory{
		ServerService: &mocks.MockServerService{},
	}

	cmd := cli.NewRootCmd(factory)
	cmd.SetArgs([]string{"server", "stop"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute() failed with %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Server has been stopped") {
		t.Errorf("Expected output to contain 'Server has been stopped', got %q", output)
	}
}
