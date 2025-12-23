// Package util provides shared utilities for the craftops application.
//nolint:revive // util is a common package name for shared utilities
package util

import (
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPClient wraps http.Client with common configuration and utilities
type HTTPClient struct {
	*http.Client
	logger *zap.Logger
}

// NewHTTPClient creates a new HTTP client with the specified timeout
func NewHTTPClient(timeout time.Duration, logger *zap.Logger) *HTTPClient {
	return &HTTPClient{
		Client: &http.Client{Timeout: timeout},
		logger: logger,
	}
}

// Do performs an HTTP request. Callers must close the response body.
// Use CloseResponseBody for safe cleanup.
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

// CloseResponseBody safely closes a response body, logging any errors
func (c *HTTPClient) CloseResponseBody(body io.Closer) {
	if body == nil {
		return
	}
	if err := body.Close(); err != nil {
		c.logger.Warn("Failed to close response body", zap.Error(err))
	}
}

// CloseResponseBodySilent closes a response body without logging (for health checks)
func CloseResponseBodySilent(body io.Closer) {
	if body != nil {
		_ = body.Close()
	}
}

