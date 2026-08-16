package ui

import (
	"bytes"
	"strings"
	"testing"

	"craftops/internal/domain"
)

func newTestTerminal() (*Terminal, *bytes.Buffer) {
	out := &bytes.Buffer{}
	return NewTerminalWithWriter(out, false), out
}

func TestTerminal_Banner(t *testing.T) {
	term, out := newTestTerminal()
	term.Banner("Hello World")
	if !strings.Contains(out.String(), "Hello World") {
		t.Errorf("Banner output missing title: %q", out.String())
	}
}

func TestTerminal_Section(t *testing.T) {
	term, out := newTestTerminal()
	term.Section("My Section")
	if !strings.Contains(out.String(), "My Section") {
		t.Errorf("Section output missing title: %q", out.String())
	}
}

func TestTerminal_Messages(t *testing.T) {
	tests := []struct {
		name  string
		call  func(*Terminal, string)
		label string
	}{
		{"Success", (*Terminal).Success, "SUCCESS"},
		{"Error", (*Terminal).Error, "ERROR"},
		{"Warning", (*Terminal).Warning, "WARNING"},
		{"Info", (*Terminal).Info, "INFO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term, out := newTestTerminal()
			tt.call(term, "test msg")
			got := out.String()
			if !strings.Contains(got, "test msg") || !strings.Contains(got, tt.label) {
				t.Errorf("got %q, want %q and %q", got, "test msg", tt.label)
			}
		})
	}
}

func TestTerminal_Step(t *testing.T) {
	term, out := newTestTerminal()
	term.Step(2, 5, "doing something")
	got := out.String()
	if !strings.Contains(got, "[2/5]") || !strings.Contains(got, "doing something") {
		t.Errorf("Step output wrong: %q", got)
	}
}

func TestTerminal_Printf(t *testing.T) {
	term, out := newTestTerminal()
	term.Printf("value=%d", 42)
	if !strings.Contains(out.String(), "value=42") {
		t.Errorf("Printf output wrong: %q", out.String())
	}
}

func TestTerminal_SprintColors_NoTTY(t *testing.T) {
	term, _ := newTestTerminal()
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
	term, out := newTestTerminal()
	term.Table([]string{"Name", "Value"}, [][]string{{"foo", "bar"}, {"baz", "qux"}})
	got := out.String()
	if !strings.Contains(got, "foo") || !strings.Contains(got, "bar") {
		t.Errorf("Table output missing data: %q", got)
	}
}

func TestTerminal_HealthCheckTable(t *testing.T) {
	term, out := newTestTerminal()
	checks := []domain.HealthCheck{
		{Name: "Database", Status: domain.StatusOK, Message: "connected"},
		{Name: "Cache", Status: domain.StatusWarn, Message: "slow"},
		{Name: "Queue", Status: domain.StatusError, Message: "down"},
	}
	term.HealthCheckTable(checks)
	got := out.String()
	for _, want := range []string{"Database", "Cache", "Queue", "connected", "slow", "down"} {
		if !strings.Contains(got, want) {
			t.Errorf("HealthCheckTable output missing %q", want)
		}
	}
}
