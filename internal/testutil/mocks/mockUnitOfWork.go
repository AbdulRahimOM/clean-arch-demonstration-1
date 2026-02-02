// Package mocks provides test doubles for application-layer interfaces.
// Shared by usecase tests, handler tests, and other tests that need to stub
// UnitOfWork, repositories, and external services.
package mocks

import "myapp/internal/application/interfaces"

// MockUnitOfWork implements interfaces.UnitOfWork for tests.
type MockUnitOfWork struct {
	ProductsRepo  *MockProductRepo
	TenantsRepo   *MockTenantRepo
	StockHistRepo *MockStockHistoryRepo
}

func (m *MockUnitOfWork) Products() interfaces.ProductRepository {
	return m.ProductsRepo
}
func (m *MockUnitOfWork) Tenants() interfaces.TenantRepository {
	return m.TenantsRepo
}
func (m *MockUnitOfWork) StockHistory() interfaces.StockHistoryRepository {
	return m.StockHistRepo
}
