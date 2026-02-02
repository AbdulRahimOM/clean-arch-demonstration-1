// Package mocks provides test doubles for application-layer interfaces.
// Shared by usecase tests, handler tests, and other tests that need to stub
// UnitOfWork, repositories, and external services.
package mocks

import (
	"context"
	"myapp/internal/domain"
)

// MockProductRepo implements interfaces.ProductRepository for tests.
type MockProductRepo struct {
	Product *domain.Product
	FindErr error
	SaveErr error
}

func (m *MockProductRepo) FindByID(ctx context.Context, productID string) (*domain.Product, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	return m.Product, nil
}

func (m *MockProductRepo) Save(ctx context.Context, product *domain.Product) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	if m.Product != nil {
		*m.Product = *product
	}
	return nil
}

func (m *MockProductRepo) UpdateStock(ctx context.Context, productID string, newStock domain.StockQuantity) error {
	return nil
}

// MockTenantRepo implements interfaces.TenantRepository for tests.
type MockTenantRepo struct {
	Tenant  *domain.Tenant
	FindErr error
}

func (m *MockTenantRepo) FindByID(ctx context.Context, tenantID string) (*domain.Tenant, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	return m.Tenant, nil
}

// MockStockHistoryRepo implements interfaces.StockHistoryRepository for tests.
// Events records all Create calls for assertions.
type MockStockHistoryRepo struct {
	CreateErr error
	Events    []domain.StockAddedEvent
}

func (m *MockStockHistoryRepo) Create(ctx context.Context, event domain.StockAddedEvent) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	m.Events = append(m.Events, event)
	return nil
}
