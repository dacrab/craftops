package mocks

import (
	"context"

	"craftops/internal/services"
)

// MockModService is a mock implementation of the ModServiceInterface.
type MockModService struct {
	UpdateAllModsFunc     func(ctx context.Context, force bool) (*services.ModUpdateResult, error)
	ListInstalledModsFunc func() ([]map[string]interface{}, error)
	HealthCheckFunc       func(ctx context.Context) []services.HealthCheck
}

// UpdateAllMods mocks the UpdateAllMods method.
func (m *MockModService) UpdateAllMods(ctx context.Context, force bool) (*services.ModUpdateResult, error) {
	if m.UpdateAllModsFunc != nil {
		return m.UpdateAllModsFunc(ctx, force)
	}
	return &services.ModUpdateResult{
		UpdatedMods: []string{"mock-mod"},
	}, nil
}

// ListInstalledMods mocks the ListInstalledMods method.
func (m *MockModService) ListInstalledMods() ([]map[string]interface{}, error) {
	if m.ListInstalledModsFunc != nil {
		return m.ListInstalledModsFunc()
	}
	return []map[string]interface{}{
		{"name": "mock-mod"},
	}, nil
}

// HealthCheck mocks the HealthCheck method.
func (m *MockModService) HealthCheck(ctx context.Context) []services.HealthCheck {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return []services.HealthCheck{
		{Name: "Mock Mod Check", Status: "✅", Message: "OK"},
	}
}
