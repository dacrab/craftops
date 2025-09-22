package mocks

import (
	"context"

	"craftops/internal/services"
)

// MockServerService is a mock implementation of the ServerServiceInterface.
type MockServerService struct {
	StartFunc     func(ctx context.Context) error
	StopFunc      func(ctx context.Context) error
	RestartFunc   func(ctx context.Context) error
	GetStatusFunc func(ctx context.Context) (*services.ServerStatus, error)
	HealthCheckFunc func(ctx context.Context) []services.HealthCheck
}

// Start mocks the Start method.
func (m *MockServerService) Start(ctx context.Context) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx)
	}
	return nil
}

// Stop mocks the Stop method.
func (m *MockServerService) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

// Restart mocks the Restart method.
func (m *MockServerService) Restart(ctx context.Context) error {
	if m.RestartFunc != nil {
		return m.RestartFunc(ctx)
	}
	return nil
}

// GetStatus mocks the GetStatus method.
func (m *MockServerService) GetStatus(ctx context.Context) (*services.ServerStatus, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx)
	}
	return &services.ServerStatus{IsRunning: true}, nil
}

// HealthCheck mocks the HealthCheck method.
func (m *MockServerService) HealthCheck(ctx context.Context) []services.HealthCheck {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return []services.HealthCheck{
		{Name: "Mock Server Check", Status: "✅", Message: "OK"},
	}
}
