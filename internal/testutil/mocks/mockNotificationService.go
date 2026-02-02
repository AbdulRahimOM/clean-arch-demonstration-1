package mocks

import (
	"context"
	"myapp/internal/domain"
)

// MockNotificationService implements interfaces.NotificationService for tests.
// StockAlerts and LowStockCalls record invocations for assertions.
type MockNotificationService struct {
	SendStockAlertErr    error
	SendLowStockAlertErr error
	StockAlerts          []domain.StockLimitAlertEvent
	LowStockCalls        int
}

func (m *MockNotificationService) SendStockAlert(ctx context.Context, event domain.StockLimitAlertEvent) error {
	m.StockAlerts = append(m.StockAlerts, event)
	return m.SendStockAlertErr
}

func (m *MockNotificationService) SendLowStockAlert(ctx context.Context, product *domain.Product, threshold int) error {
	m.LowStockCalls++
	return m.SendLowStockAlertErr
}
