package mocks

import (
	"context"

	"craftops/internal/services"
)

// MockNotificationService is a mock implementation of the NotificationServiceInterface.
type MockNotificationService struct {
	SendSuccessNotificationFunc func(ctx context.Context, message string) error
	SendErrorNotificationFunc   func(ctx context.Context, message string) error
	SendRestartWarningsFunc     func(ctx context.Context) error
	HealthCheckFunc             func(ctx context.Context) []services.HealthCheck
}

// SendSuccessNotification mocks the SendSuccessNotification method.
func (m *MockNotificationService) SendSuccessNotification(ctx context.Context, message string) error {
	if m.SendSuccessNotificationFunc != nil {
		return m.SendSuccessNotificationFunc(ctx, message)
	}
	return nil
}

// SendErrorNotification mocks the SendErrorNotification method.
func (m *MockNotificationService) SendErrorNotification(ctx context.Context, message string) error {
	if m.SendErrorNotificationFunc != nil {
		return m.SendErrorNotificationFunc(ctx, message)
	}
	return nil
}

// SendRestartWarnings mocks the SendRestartWarnings method.
func (m *MockNotificationService) SendRestartWarnings(ctx context.Context) error {
	if m.SendRestartWarningsFunc != nil {
		return m.SendRestartWarningsFunc(ctx)
	}
	return nil
}

// HealthCheck mocks the HealthCheck method.
func (m *MockNotificationService) HealthCheck(ctx context.Context) []services.HealthCheck {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return []services.HealthCheck{
		{Name: "Mock Notification Check", Status: "✅", Message: "OK"},
	}
}
