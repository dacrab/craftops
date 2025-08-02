package services_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/services"
)

func TestNewModService(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	service := services.NewModService(cfg, logger)
	if service == nil {
		t.Fatal("NewModService returned nil")
	}

	// Test that service was created successfully (can't access private fields)
	// Verify by calling a public method
	mods, err := service.ListInstalledMods()
	if err != nil {
		t.Errorf("Service not properly initialized: %v", err)
	}
	if mods == nil {
		t.Error("Service should return empty slice, not nil")
	}
}

func TestModServiceHealthCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	service := services.NewModService(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	checks := service.HealthCheck(ctx)
	if len(checks) == 0 {
		t.Error("HealthCheck should return at least one check")
	}

	// Verify check structure
	for _, check := range checks {
		if check.Name == "" {
			t.Error("Health check name should not be empty")
		}
		if check.Status == "" {
			t.Error("Health check status should not be empty")
		}
		if check.Message == "" {
			t.Error("Health check message should not be empty")
		}
	}
}

func TestModServicePublicMethods(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := zap.NewNop()
	service := services.NewModService(cfg, logger)

	// Test ListInstalledMods (public method)
	mods, err := service.ListInstalledMods()
	if err != nil {
		t.Errorf("ListInstalledMods should not error: %v", err)
	}
	if mods == nil {
		t.Error("ListInstalledMods should return empty slice, not nil")
	}

	// Test UpdateAllMods with dry run (won't actually do anything)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will fail gracefully if no mods are configured, which is expected
	_, _ = service.UpdateAllMods(ctx, false)
	// We don't check for error here as it's expected to fail without proper config
}
