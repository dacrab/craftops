package ui

import (
	"bytes"
	"testing"

	"craftops/internal/domain"
)

func TestNewTerminal(t *testing.T) {
	term := NewTerminal()
	if term == nil {
		t.Fatal("NewTerminal returned nil")
	}
}

func TestTerminalWithWriter(t *testing.T) {
	var out, errOut bytes.Buffer
	term := NewTerminalWithWriter(&out, &errOut, false)

	term.Success("test success")
	if !bytes.Contains(out.Bytes(), []byte("SUCCESS")) {
		t.Error("Success should output SUCCESS prefix in non-TTY mode")
	}
}

func TestTerminalOutput(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	term.Error("error message")
	if !bytes.Contains(out.Bytes(), []byte("ERROR")) {
		t.Error("Error should output ERROR prefix")
	}

	out.Reset()
	term.Warning("warning message")
	if !bytes.Contains(out.Bytes(), []byte("WARNING")) {
		t.Error("Warning should output WARNING prefix")
	}

	out.Reset()
	term.Info("info message")
	if !bytes.Contains(out.Bytes(), []byte("INFO")) {
		t.Error("Info should output INFO prefix")
	}
}

func TestTerminalBanner(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	term.Banner("Test Banner")
	if !bytes.Contains(out.Bytes(), []byte("Test Banner")) {
		t.Error("Banner should contain title")
	}
}

func TestTerminalSection(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	term.Section("Test Section")
	if !bytes.Contains(out.Bytes(), []byte("Test Section")) {
		t.Error("Section should contain title")
	}
}

func TestTerminalStep(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	term.Step(1, 3, "step message")
	if !bytes.Contains(out.Bytes(), []byte("[1/3]")) {
		t.Error("Step should contain step numbers")
	}
}

func TestTerminalTable(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"key1", "val1"},
		{"key2", "val2"},
	}

	term.Table(headers, rows)
	output := out.String()

	if !bytes.Contains([]byte(output), []byte("Name")) {
		t.Error("Table should contain header")
	}
	if !bytes.Contains([]byte(output), []byte("key1")) {
		t.Error("Table should contain row data")
	}
}

func TestTerminalHealthCheckTable(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	checks := []domain.HealthCheck{
		{Name: "Test1", Status: domain.StatusOK, Message: "OK"},
		{Name: "Test2", Status: domain.StatusWarn, Message: "Warn"},
		{Name: "Test3", Status: domain.StatusError, Message: "Error"},
	}

	term.HealthCheckTable(checks)
	output := out.String()

	if !bytes.Contains([]byte(output), []byte("Test1")) {
		t.Error("HealthCheckTable should contain check names")
	}
}

func TestTerminalColorMethods(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	// In non-TTY mode, these should just return the plain text
	if term.SuccessSprint("test") != "test" {
		t.Error("SuccessSprint should return plain text in non-TTY")
	}
	if term.ErrorSprint("test") != "test" {
		t.Error("ErrorSprint should return plain text in non-TTY")
	}
	if term.WarningSprint("test") != "test" {
		t.Error("WarningSprint should return plain text in non-TTY")
	}
	if term.DimSprint("test") != "test" {
		t.Error("DimSprint should return plain text in non-TTY")
	}
	if term.AccentSprintf("test %s", "arg") != "test arg" {
		t.Error("AccentSprintf should format text")
	}
}

func TestTerminalIsTTY(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	if term.IsTTY() {
		t.Error("Terminal with isTTY=false should return false")
	}
}

func TestTerminalPrintf(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	term.Printf("hello %s", "world")
	if !bytes.Contains(out.Bytes(), []byte("hello world")) {
		t.Error("Printf should format output")
	}
}

func TestTerminalPrintln(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	term.Println("test line")
	if !bytes.Contains(out.Bytes(), []byte("test line")) {
		t.Error("Println should output text")
	}
}
