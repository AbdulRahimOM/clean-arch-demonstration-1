// internal/application/interfaces/repository.go
package interfaces

import (
	"context"
	"myapp/internal/domain"
)

// Repository interfaces defined by application layer
type ProductRepository interface {
	FindByID(ctx context.Context, productID string) (*domain.Product, error)
	Save(ctx context.Context, product *domain.Product) error
	UpdateStock(ctx context.Context, productID string, newStock domain.StockQuantity) error
}

type TenantRepository interface {
	FindByID(ctx context.Context, tenantID string) (*domain.Tenant, error)
}

type StockHistoryRepository interface {
	Create(ctx context.Context, event domain.StockAddedEvent) error
}

// Unit of Work pattern for transaction
type UnitOfWork interface {
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Products() ProductRepository
	Tenants() TenantRepository
	StockHistory() StockHistoryRepository
}