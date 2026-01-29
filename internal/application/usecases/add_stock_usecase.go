// internal/application/usecases/add_stock_usecase.go
package usecases

import (
	"context"
	"myapp/internal/application/interfaces"
	"myapp/internal/domain"
	"time"
)

// Input DTO (Application-specific, not HTTP-specific)
type AddStockRequest struct {
	ProductID string
	Quantity  int
	TenantID  string
	Notes     string
	AddedBy   string
}

// Output DTO
type AddStockResponse struct {
	ProductID     string
	ProductName   string
	PreviousStock int
	NewStock      int
	Added         int
	MaxAllowed    int
	Utilization   float64
}

// Use Case interface (what handlers depend on)
type AddStockUseCase interface {
	Execute(ctx context.Context, req AddStockRequest) (*AddStockResponse, error)
}

// Implementation
type addStockUseCase struct {
	uow                   interfaces.UnitOfWork
	notificationSvc       interfaces.NotificationService
	eventPublisher        interfaces.EventPublisher
	recentUpdateThreshold time.Duration
}

func NewAddStockUseCase(
	uow interfaces.UnitOfWork,
	notificationSvc interfaces.NotificationService,
	eventPublisher interfaces.EventPublisher,
) AddStockUseCase {
	return &addStockUseCase{
		uow:                   uow,
		notificationSvc:       notificationSvc,
		eventPublisher:        eventPublisher,
		recentUpdateThreshold: 5 * time.Minute,
	}
}

func (uc *addStockUseCase) Execute(ctx context.Context, req AddStockRequest) (*AddStockResponse, error) {
	// 1. Validate input
	if err := uc.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. Begin transaction
	if err := uc.uow.Begin(ctx); err != nil {
		return nil, err
	}
	defer uc.uow.Rollback(ctx) // Safe rollback if not committed

	// 3. Get tenant
	tenant, err := uc.uow.Tenants().FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	// 4. Validate tenant
	if err := tenant.CanReceiveStock(); err != nil {
		return nil, err
	}

	// 5. Get product
	product, err := uc.uow.Products().FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	// 6. Create quantity value object
	quantity, err := domain.NewStockQuantity(req.Quantity)
	if err != nil {
		return nil, err
	}

	// 7. Business rule: Check if product was recently updated
	if product.IsRecentlyUpdated(uc.recentUpdateThreshold) {
		// Could log or handle as needed
		// domain event could be published
	}

	// 8. Add stock with business logic
	previousStock := product.CurrentStock
	if err := product.AddStock(quantity, tenant.MaxStock); err != nil {
		return nil, err
	}

	// 9. Save updated product
	if err := uc.uow.Products().Save(ctx, product); err != nil {
		return nil, err
	}

	// 10. Create audit log
	stockEvent := domain.StockAddedEvent{
		ProductID: product.ID,
		TenantID:  req.TenantID,
		Quantity:  quantity,
		Previous:  previousStock,
		Current:   product.CurrentStock,
		AddedBy:   req.AddedBy,
		Timestamp: time.Now(),
		Notes:     req.Notes,
	}

	if err := uc.uow.StockHistory().Create(ctx, stockEvent); err != nil {
		return nil, err
	}

	// 11. Check if stock limit alert needed
	utilization := product.UtilizationPercentage(tenant.MaxStock)
	if utilization > 80 {
		alertEvent := domain.StockLimitAlertEvent{
			ProductID:   product.ID,
			ProductName: product.Name,
			Current:     product.CurrentStock,
			MaxLimit:    tenant.MaxStock,
			Utilization: utilization,
			TenantID:    req.TenantID,
			Timestamp:   time.Now(),
		}

		// Async notification (fire and forget in background)
		go func() {
			ctx := context.Background()
			_ = uc.notificationSvc.SendStockAlert(ctx, alertEvent)
		}()
	}

	// 12. Check for low stock
	if product.IsLowStock(10) {
		go func() {
			ctx := context.Background()
			_ = uc.notificationSvc.SendLowStockAlert(ctx, product, 10)
		}()
	}

	// 13. Publish domain event
	if uc.eventPublisher != nil {
		_ = uc.eventPublisher.Publish(ctx, stockEvent)
	}

	// 14. Commit transaction
	if err := uc.uow.Commit(ctx); err != nil {
		return nil, err
	}

	// 15. Return response
	return &AddStockResponse{
		ProductID:     product.ID,
		ProductName:   product.Name,
		PreviousStock: previousStock.Value(),
		NewStock:      product.CurrentStock.Value(),
		Added:         quantity.Value(),
		MaxAllowed:    tenant.MaxStock.Value(),
		Utilization:   utilization,
	}, nil
}

func (uc *addStockUseCase) validateRequest(req AddStockRequest) error {
	if req.ProductID == "" {
		return domain.ErrInvalidProductID
	}
	if req.TenantID == "" {
		return domain.ErrTenantNotFound
	}
	if req.Quantity <= 0 {
		return domain.ErrInvalidQuantity
	}
	return nil
}
