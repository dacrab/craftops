//nolint:revive // util is a common package name for shared utilities
package util

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"craftops/internal/domain"
)

// RetryConfig holds configuration for retry operations
type RetryConfig struct {
	MaxRetries int
	RetryDelay float64
}

// WithRetry wraps an operation with a backoff-based retry loop using github.com/cenkalti/backoff/v4
func WithRetry(ctx context.Context, cfg RetryConfig, op func() error) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = time.Duration(cfg.RetryDelay) * time.Second
	b.MaxInterval = time.Duration(cfg.RetryDelay*10) * time.Second
	b.Reset()

	// Use a counter to respect MaxRetries while using exponential backoff for delays
	attempt := 0
	return backoff.Retry(func() error {
		err := op()
		if err == nil {
			return nil
		}

		// Don't retry if error is not retryable
		if apiErr, ok := err.(*domain.APIError); ok && !apiErr.IsRetryable() {
			return backoff.Permanent(err)
		}

		attempt++
		if attempt > cfg.MaxRetries {
			return backoff.Permanent(err)
		}

		return err
	}, backoff.WithContext(b, ctx))
}

