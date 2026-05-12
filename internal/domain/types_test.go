package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if !strings.Contains(got, "404") || !strings.Contains(got, "not found") || !strings.Contains(got, "http://example.com") {
		t.Errorf("APIError.Error() = %q, missing expected content", got)
	}

	e2 := &APIError{URL: "http://example.com", Message: "bad gateway"}
	got2 := e2.Error()
	if !strings.Contains(got2, "bad gateway") || strings.Contains(got2, "[0]") {
		t.Errorf("APIError.Error() (no status) = %q", got2)
	}
}

func TestCheckPath(t *testing.T) {
	tmp := t.TempDir()

	t.Run("exists and is dir", func(t *testing.T) {
		c := CheckPath("test", tmp)
		if c.Status != StatusOK {
			t.Errorf("expected OK, got %s: %s", c.Status, c.Message)
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

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{0, "0 B"},
		{999, "999 B"},
		{1000, "kB"},
		{1000 * 1000, "MB"},
		{-1, "0 B"},
	}
	for _, tt := range tests {
		got := FormatSize(tt.size)
		if !strings.Contains(got, tt.want) {
			t.Errorf("FormatSize(%d) = %q, want it to contain %q", tt.size, got, tt.want)
		}
	}
}
