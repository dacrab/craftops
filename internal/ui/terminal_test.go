package ui

import (
	"bytes"
	"testing"

	"craftops/internal/domain"
)

func TestTerminal(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminalWithWriter(&out, &out, false)

	t.Run("Messages", func(t *testing.T) {
		tests := []struct {
			name     string
			call     func()
			contains string
		}{
			{"Success", func() { term.Success("ok") }, "SUCCESS"},
			{"Error", func() { term.Error("err") }, "ERROR"},
			{"Warning", func() { term.Warning("warn") }, "WARNING"},
			{"Info", func() { term.Info("info") }, "INFO"},
			{"Banner", func() { term.Banner("title") }, "title"},
			{"Section", func() { term.Section("sec") }, "sec"},
			{"Step", func() { term.Step(1, 1, "msg") }, "[1/1]"},
		}

		for _, tt := range tests {
			out.Reset()
			tt.call()
			if !bytes.Contains(out.Bytes(), []byte(tt.contains)) {
				t.Errorf("%s: output missing %q", tt.name, tt.contains)
			}
		}
	})

	t.Run("Tables", func(t *testing.T) {
		out.Reset()
		term.Table([]string{"H1"}, [][]string{{"V1"}})
		if !bytes.Contains(out.Bytes(), []byte("H1")) || !bytes.Contains(out.Bytes(), []byte("V1")) {
			t.Error("Table output missing data")
		}

		out.Reset()
		term.HealthCheckTable([]domain.HealthCheck{{Name: "C1", Status: domain.StatusOK, Message: "M1"}})
		if !bytes.Contains(out.Bytes(), []byte("C1")) || !bytes.Contains(out.Bytes(), []byte("M1")) {
			t.Error("HealthCheckTable output missing data")
		}
	})

	t.Run("SprintMethods", func(t *testing.T) {
		if term.SuccessSprint("t") != "t" || term.ErrorSprint("t") != "t" {
			t.Error("Sprint methods should return plain text in non-TTY")
		}
	})
}