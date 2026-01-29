// internal/application/interfaces/service.go
package interfaces

import (
	"context"
	"myapp/internal/domain"
)

// External services interfaces
type NotificationService interface {
	SendStockAlert(ctx context.Context, event domain.StockLimitAlertEvent) error
	SendLowStockAlert(ctx context.Context, product *domain.Product, threshold int) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}

// Validator interface
type Validator interface {
	Validate(ctx context.Context, data interface{}) error
}