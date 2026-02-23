package ui

import (
	"bytes"
	"strings"
	"testing"

	"craftops/internal/domain"
)

// newTestTerminal returns a non-TTY terminal writing to controlled buffers.
func newTestTerminal() (*Terminal, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	t := NewTerminalWithWriter(out, errOut, false)
	return t, out, errOut
}

func TestTerminal_IsTTY(t *testing.T) {
	term, _, _ := newTestTerminal()
	if term.IsTTY() {
		t.Error("expected IsTTY=false for test terminal")
	}
}

func TestTerminal_Banner(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Banner("Hello World")
	if !strings.Contains(out.String(), "Hello World") {
		t.Errorf("Banner output missing title: %q", out.String())
	}
}

func TestTerminal_Section(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Section("My Section")
	if !strings.Contains(out.String(), "My Section") {
		t.Errorf("Section output missing title: %q", out.String())
	}
}

func TestTerminal_Success(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Success("it worked")
	got := out.String()
	if !strings.Contains(got, "it worked") {
		t.Errorf("Success output missing message: %q", got)
	}
	if !strings.Contains(got, "SUCCESS") {
		t.Errorf("Success output missing label: %q", got)
	}
}

func TestTerminal_Error(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Error("it broke")
	got := out.String()
	if !strings.Contains(got, "it broke") {
		t.Errorf("Error output missing message: %q", got)
	}
	if !strings.Contains(got, "ERROR") {
		t.Errorf("Error output missing label: %q", got)
	}
}

func TestTerminal_Warning(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Warning("be careful")
	got := out.String()
	if !strings.Contains(got, "be careful") {
		t.Errorf("Warning output missing message: %q", got)
	}
	if !strings.Contains(got, "WARNING") {
		t.Errorf("Warning output missing label: %q", got)
	}
}

func TestTerminal_Info(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Info("some info")
	got := out.String()
	if !strings.Contains(got, "some info") {
		t.Errorf("Info output missing message: %q", got)
	}
	if !strings.Contains(got, "INFO") {
		t.Errorf("Info output missing label: %q", got)
	}
}

func TestTerminal_Step(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Step(2, 5, "doing something")
	got := out.String()
	if !strings.Contains(got, "[2/5]") {
		t.Errorf("Step output missing progress indicator: %q", got)
	}
	if !strings.Contains(got, "doing something") {
		t.Errorf("Step output missing message: %q", got)
	}
}

func TestTerminal_Printf(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Printf("value=%d", 42)
	if !strings.Contains(out.String(), "value=42") {
		t.Errorf("Printf output wrong: %q", out.String())
	}
}

func TestTerminal_Println(t *testing.T) {
	term, out, _ := newTestTerminal()
	term.Println("hello", "world")
	if !strings.Contains(out.String(), "hello") {
		t.Errorf("Println output wrong: %q", out.String())
	}
}

func TestTerminal_SprintColors_NoTTY(t *testing.T) {
	term, _, _ := newTestTerminal()
	// Without TTY, sprint methods should return the raw text unchanged
	if got := term.SuccessSprint("ok"); got != "ok" {
		t.Errorf("SuccessSprint non-TTY: got %q, want %q", got, "ok")
	}
	if got := term.ErrorSprint("fail"); got != "fail" {
		t.Errorf("ErrorSprint non-TTY: got %q, want %q", got, "fail")
	}
	if got := term.WarningSprint("warn"); got != "warn" {
		t.Errorf("WarningSprint non-TTY: got %q, want %q", got, "warn")
	}
	if got := term.DimSprint("dim"); got != "dim" {
		t.Errorf("DimSprint non-TTY: got %q, want %q", got, "dim")
	}
}

func TestTerminal_Table(t *testing.T) {
	term, out, _ := newTestTerminal()
	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"foo", "bar"},
		{"baz", "qux"},
	}
	term.Table(headers, rows)
	got := out.String()
	if !strings.Contains(got, "foo") || !strings.Contains(got, "bar") {
		t.Errorf("Table output missing data: %q", got)
	}
}

func TestTerminal_HealthCheckTable(t *testing.T) {
	term, out, _ := newTestTerminal()
	checks := []domain.HealthCheck{
		{Name: "Database", Status: domain.StatusOK, Message: "connected"},
		{Name: "Cache", Status: domain.StatusWarn, Message: "slow"},
		{Name: "Queue", Status: domain.StatusError, Message: "down"},
	}
	term.HealthCheckTable(checks)
	got := out.String()
	for _, want := range []string{"Database", "Cache", "Queue", "connected", "slow", "down"} {
		if !strings.Contains(got, want) {
			t.Errorf("HealthCheckTable output missing %q: %q", want, got)
		}
	}
}

func TestTerminal_AccentSprintf_NoTTY(t *testing.T) {
	term, _, _ := newTestTerminal()
	got := term.AccentSprintf("count=%d", 7)
	if got != "count=7" {
		t.Errorf("AccentSprintf non-TTY: got %q, want %q", got, "count=7")
	}
}
