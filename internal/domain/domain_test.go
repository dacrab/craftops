package domain

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAPIError_IsRetryable(t *testing.T) {
	tests := []struct {
		code      int
		retryable bool
	}{
		{200, false}, {400, false}, {404, false},
		{429, true}, {500, true}, {502, true}, {503, true},
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
	if got := e.Error(); got == "" {
		t.Error("APIError.Error() returned empty string")
	}
	// Without status code
	e2 := &APIError{URL: "http://example.com", Message: "bad"}
	if got := e2.Error(); got == "" {
		t.Error("APIError.Error() returned empty string")
	}
}

func TestServiceError(t *testing.T) {
	inner := errors.New("inner error")
	se := NewServiceError("server", "start", inner)

	var target *ServiceError
	if !errors.As(se, &target) {
		t.Fatal("expected *ServiceError via errors.As")
	}
	if target.Service != "server" || target.Op != "start" {
		t.Errorf("unexpected fields: %+v", target)
	}
	if !errors.Is(se, inner) {
		t.Error("Unwrap should expose inner error")
	}
	// Without Op
	se2 := &ServiceError{Service: "backup", Err: inner}
	if se2.Error() == "" {
		t.Error("ServiceError.Error() returned empty string")
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
			t.Errorf("expected WARN, got %s", c.Status)
		}
	})

	t.Run("exists but is a file", func(t *testing.T) {
		f := filepath.Join(tmp, "file.txt")
		_ = os.WriteFile(f, []byte("x"), 0o600)
		c := CheckPath("test", f)
		if c.Status != StatusError {
			t.Errorf("expected ERROR, got %s", c.Status)
		}
	})
}

func TestBackupInfo_SizeFormatted(t *testing.T) {
	tests := []struct {
		size    int64
		wantErr bool // just verify no empty string
	}{
		{0, false},
		{1024, false},
		{1024 * 1024, false},
		{-1, false}, // negative clamped to 0 B
	}
	for _, tt := range tests {
		b := BackupInfo{Size: tt.size}
		if got := b.SizeFormatted(); got == "" {
			t.Errorf("SizeFormatted(%d) returned empty string", tt.size)
		}
	}
}
