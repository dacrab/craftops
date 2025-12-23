package domain

import (
	"errors"
	"testing"
	"time"
)

func TestHealthStatus(t *testing.T) {
	tests := []HealthStatus{StatusOK, StatusWarn, StatusError}
	for _, s := range tests {
		if s == "" {
			t.Error("HealthStatus should not be empty")
		}
	}
}

func TestHealthCheck(t *testing.T) {
	check := HealthCheck{Name: "Test", Status: StatusOK, Message: "OK"}
	if check.Name == "" {
		t.Error("Name should not be empty")
	}
}

func TestServerStatus(t *testing.T) {
	status := ServerStatus{IsRunning: true, SessionName: "minecraft", CheckedAt: time.Now()}
	if !status.IsRunning {
		t.Error("IsRunning should be true")
	}
}

func TestModUpdateResult(t *testing.T) {
	result := ModUpdateResult{
		UpdatedMods: []string{"mod1"},
		FailedMods:  map[string]string{"mod2": "error"},
		SkippedMods: []string{"mod3"},
	}
	if len(result.UpdatedMods) != 1 {
		t.Error("UpdatedMods should have 1 item")
	}
}

func TestBackupInfoSizeFormatted(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{1024 * 1024, "1.0 MB"},
		{10 * 1024 * 1024, "10.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tt := range tests {
		b := BackupInfo{Size: tt.size}
		if got := b.SizeFormatted(); got != tt.expected {
			t.Errorf("SizeFormatted(%d) = %q, want %q", tt.size, got, tt.expected)
		}
	}
}

func TestSentinelErrors(t *testing.T) {
	errs := []error{
		ErrServerJarNotFound,
		ErrBackupsDisabled,
	}
	for _, err := range errs {
		if err == nil || err.Error() == "" {
			t.Error("sentinel error should have message")
		}
	}
}

func TestServiceError(t *testing.T) {
	inner := errors.New("inner error")
	err := &ServiceError{Service: "test", Op: "do", Err: inner}
	if err.Error() == "" || errors.Unwrap(err) != inner {
		t.Error("ServiceError mismatch")
	}

	errNoOp := &ServiceError{Service: "test", Err: inner}
	if errNoOp.Error() == "" {
		t.Error("ServiceError without op should have message")
	}
}

func TestNewServiceError(t *testing.T) {
	err := NewServiceError("svc", "op", errors.New("test"))
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Error("should be ServiceError")
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{URL: "http://test.com", StatusCode: 500, Message: "error"}
	if err.Error() == "" || !err.IsRetryable() {
		t.Error("APIError mismatch")
	}

	tests := []struct {
		code      int
		retryable bool
	}{
		{200, false}, {400, false}, {404, false},
		{429, true}, {500, true}, {502, true}, {503, true},
	}
	for _, tt := range tests {
		e := &APIError{StatusCode: tt.code}
		if e.IsRetryable() != tt.retryable {
			t.Errorf("IsRetryable(%d) = %v, want %v", tt.code, e.IsRetryable(), tt.retryable)
		}
	}

	errNoCode := &APIError{URL: "http://test.com", Message: "connection failed"}
	if errNoCode.Error() == "" {
		t.Error("APIError without status should have message")
	}
}
