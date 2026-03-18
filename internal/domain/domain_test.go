package domain

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAPIError_IsRetryable(t *testing.T) {
	tests := []struct {
		code      int
		retryable bool
	}{
		{200, false},
		{400, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
	}
	for _, tt := range tests {
		e := &APIError{StatusCode: tt.code}
		if got := e.IsRetryable(); got != tt.retryable {
			t.Errorf("IsRetryable(%d) = %v, want %v", tt.code, got, tt.retryable)
		}
	}
}

func TestAPIError_Error(t *testing.T) {
	e := &APIError{URL: "http://example.com", StatusCode: 404, Message: "not found"}
	got := e.Error()
	if !strings.Contains(got, "404") {
		t.Errorf("APIError.Error() missing status code: %q", got)
	}
	if !strings.Contains(got, "not found") {
		t.Errorf("APIError.Error() missing message: %q", got)
	}
	if !strings.Contains(got, "http://example.com") {
		t.Errorf("APIError.Error() missing URL: %q", got)
	}

	// Without status code — different format branch
	e2 := &APIError{URL: "http://example.com", Message: "bad gateway"}
	got2 := e2.Error()
	if !strings.Contains(got2, "bad gateway") {
		t.Errorf("APIError.Error() (no status) missing message: %q", got2)
	}
	if strings.Contains(got2, "[0]") {
		t.Errorf("APIError.Error() (no status) should not print zero status code: %q", got2)
	}
}

func TestServiceError_WithOp(t *testing.T) {
	inner := errors.New("inner error")
	se := NewServiceError("server", "start", inner)

	// errors.As unwrap
	var target *ServiceError
	if !errors.As(se, &target) {
		t.Fatal("expected *ServiceError via errors.As")
	}
	if target.Service != "server" {
		t.Errorf("Service = %q, want %q", target.Service, "server")
	}
	if target.Op != "start" {
		t.Errorf("Op = %q, want %q", target.Op, "start")
	}

	// Unwrap exposes inner
	if !errors.Is(se, inner) {
		t.Error("errors.Is should find inner error via Unwrap")
	}

	// Error string format: "service.op: inner"
	want := "server.start: inner error"
	if got := se.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestServiceError_WithoutOp(t *testing.T) {
	inner := errors.New("disk full")
	se := &ServiceError{Service: "backup", Err: inner}

	// Error string format: "service: inner" (no Op component)
	want := "backup: disk full"
	if got := se.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestCheckPath(t *testing.T) {
	tmp := t.TempDir()

	t.Run("exists and is dir", func(t *testing.T) {
		c := CheckPath("test", tmp)
		if c.Status != StatusOK {
			t.Errorf("expected OK, got %s: %s", c.Status, c.Message)
		}
		if c.Name != "test" {
			t.Errorf("Name = %q, want %q", c.Name, "test")
		}
	})

	t.Run("does not exist", func(t *testing.T) {
		c := CheckPath("test", filepath.Join(tmp, "nonexistent"))
		if c.Status != StatusWarn {
			t.Errorf("expected WARN, got %s: %s", c.Status, c.Message)
		}
	})

	t.Run("exists but is a file", func(t *testing.T) {
		f := filepath.Join(tmp, "file.txt")
		_ = os.WriteFile(f, []byte("x"), 0o600)
		c := CheckPath("test", f)
		if c.Status != StatusError {
			t.Errorf("expected ERROR, got %s: %s", c.Status, c.Message)
		}
	})
}

func TestBackupInfo_SizeFormatted(t *testing.T) {
	tests := []struct {
		size int64
		want string // substring expected in output
	}{
		{0, "B"},
		{1000, "kB"},       // go-humanize uses SI (decimal) units
		{1000 * 1000, "MB"}, // 1 MB in SI
		{-1, "0 B"},         // negative clamped
	}
	for _, tt := range tests {
		b := BackupInfo{Size: tt.size}
		got := b.SizeFormatted()
		if !strings.Contains(got, tt.want) {
			t.Errorf("SizeFormatted(%d) = %q, want it to contain %q", tt.size, got, tt.want)
		}
	}
}

func TestServerStatus_Fields(t *testing.T) {
	before := time.Now()
	s := &ServerStatus{
		IsRunning:   true,
		SessionName: "minecraft",
		CheckedAt:   time.Now(),
	}
	if !s.IsRunning {
		t.Error("IsRunning should be true")
	}
	if s.SessionName != "minecraft" {
		t.Errorf("SessionName = %q, want %q", s.SessionName, "minecraft")
	}
	if s.CheckedAt.Before(before) {
		t.Error("CheckedAt should not be before test start")
	}
}

func TestModUpdateResult_ZeroValues(t *testing.T) {
	r := &ModUpdateResult{
		UpdatedMods: []string{},
		FailedMods:  map[string]string{},
		SkippedMods: []string{},
	}
	if len(r.UpdatedMods) != 0 {
		t.Error("UpdatedMods should be empty")
	}
	if len(r.FailedMods) != 0 {
		t.Error("FailedMods should be empty")
	}
	if len(r.SkippedMods) != 0 {
		t.Error("SkippedMods should be empty")
	}
}

func TestHealthCheck_StatusConstants(t *testing.T) {
	if StatusOK == StatusWarn || StatusOK == StatusError || StatusWarn == StatusError {
		t.Error("health status constants must all be distinct")
	}
}
