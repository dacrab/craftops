package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// BaseService provides common functionality for all services
type BaseService struct {
	config *config.Config
	logger *zap.Logger
}

// NewBaseService creates a new base service
func NewBaseService(cfg *config.Config, logger *zap.Logger) *BaseService {
	return &BaseService{
		config: cfg,
		logger: logger,
	}
}

// GetConfig returns the service configuration
func (bs *BaseService) GetConfig() *config.Config {
	return bs.config
}

// GetLogger returns the service logger
func (bs *BaseService) GetLogger() *zap.Logger {
	return bs.logger
}

// LogOperation logs the start and completion of an operation
func (bs *BaseService) LogOperation(operation string, fn func() error) error {
	bs.logger.Info("Starting operation", zap.String("operation", operation))
	start := time.Now()

	err := fn()
	duration := time.Since(start)

	if err != nil {
		bs.logger.Error("Operation failed",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
			zap.Error(err))
		return err
	}

	bs.logger.Info("Operation completed successfully",
		zap.String("operation", operation),
		zap.Duration("duration", duration))
	return nil
}

// ExecuteWithTimeout executes a function with a timeout
func (bs *BaseService) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v: %w", timeout, ctx.Err())
	}
}

// HandleDryRun logs dry run operations and returns early
func (bs *BaseService) HandleDryRun(operation string, details ...string) bool {
	if !bs.config.DryRun {
		return false
	}

	message := fmt.Sprintf("Dry run: Would %s", operation)
	if len(details) > 0 {
		message += fmt.Sprintf(" (%s)", details[0])
	}

	bs.logger.Info(message)
	return true
}

// CreateHealthCheck creates a standardized health check result
func (bs *BaseService) CreateHealthCheck(name, status, message string) HealthCheck {
	return HealthCheck{
		Name:    name,
		Status:  status,
		Message: message,
	}
}

// ValidateDirectory checks if a directory exists and is accessible
func (bs *BaseService) ValidateDirectory(path, name string) HealthCheck {
	if path == "" {
		return bs.CreateHealthCheck(name, "❌", "Path not configured")
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return bs.CreateHealthCheck(name, "✅", "OK")
	}

	return bs.CreateHealthCheck(name, "❌", "Directory not found or not accessible")
}
