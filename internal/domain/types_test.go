package domain

import (
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
