package domain

import (
	"testing"
)

// TestAPIError_IsRetryable tests the retry logic for HTTP status codes
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
		if e.IsRetryable() != tt.retryable {
			t.Errorf("IsRetryable(%d) = %v, want %v", tt.code, e.IsRetryable(), tt.retryable)
		}
	}
}
