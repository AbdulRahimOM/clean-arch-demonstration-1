package usecases

import (
	"context"
	"errors"
	"myapp/internal/domain"
	"myapp/internal/testutil/mocks"
	"testing"
	"time"
)

var (
	errFindProduct   = errors.New("product not found")
	errFindTenant    = errors.New("tenant not found")
	errSaveProduct   = errors.New("save product failed")
	errCreateHistory = errors.New("create history failed")
)

func mustQuantity(n int) domain.StockQuantity {
	q, err := domain.NewStockQuantity(n)
	if err != nil {
		panic(err)
	}
	return q
}

func TestAddStockUseCase_Execute_Validation(t *testing.T) {
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{},
		TenantsRepo:   &mocks.MockTenantRepo{},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, &mocks.MockNotificationService{}, nil)
	ctx := context.Background()

	tests := []struct {
		name string
		req  AddStockRequest
		want error
	}{
		{
			name: "empty product id",
			req:  AddStockRequest{ProductID: "", TenantID: "t1", Quantity: 5},
			want: domain.ErrInvalidProductID,
		},
		{
			name: "empty tenant id",
			req:  AddStockRequest{ProductID: "p1", TenantID: "", Quantity: 5},
			want: domain.ErrTenantNotFound,
		},
		{
			name: "zero quantity",
			req:  AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 0},
			want: domain.ErrInvalidQuantity,
		},
		{
			name: "negative quantity",
			req:  AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: -1},
			want: domain.ErrInvalidQuantity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uc.Execute(ctx, tt.req)
			if got != nil {
				t.Fatalf("Execute() expected nil response on validation error, got %+v", got)
			}
			if err == nil || !errors.Is(err, tt.want) {
				t.Errorf("Execute() err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestAddStockUseCase_Execute_TenantNotFound(t *testing.T) {
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{},
		TenantsRepo:   &mocks.MockTenantRepo{FindErr: errFindTenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, &mocks.MockNotificationService{}, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 5, AddedBy: "u1"}
	got, err := uc.Execute(ctx, req)
	if got != nil {
		t.Fatalf("Execute() expected nil response, got %+v", got)
	}
	if err == nil || !errors.Is(err, errFindTenant) {
		t.Errorf("Execute() err = %v, want %v", err, errFindTenant)
	}
}

func TestAddStockUseCase_Execute_TenantInactive(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(100), IsActive: false}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, &mocks.MockNotificationService{}, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 5, AddedBy: "u1"}
	got, err := uc.Execute(ctx, req)
	if got != nil {
		t.Fatalf("Execute() expected nil response, got %+v", got)
	}
	if err == nil || !errors.Is(err, domain.ErrTenantInactive) {
		t.Errorf("Execute() err = %v, want %v", err, domain.ErrTenantInactive)
	}
}

func TestAddStockUseCase_Execute_ProductNotFound(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(100), IsActive: true}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{FindErr: errFindProduct},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, &mocks.MockNotificationService{}, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 5, AddedBy: "u1"}
	got, err := uc.Execute(ctx, req)
	if got != nil {
		t.Fatalf("Execute() expected nil response, got %+v", got)
	}
	if err == nil || !errors.Is(err, errFindProduct) {
		t.Errorf("Execute() err = %v, want %v", err, errFindProduct)
	}
}

func TestAddStockUseCase_Execute_StockExceedsLimit(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(10), IsActive: true}
	product := &domain.Product{
		ID: "p1", Name: "Product", CurrentStock: mustQuantity(8),
		LastUpdated: time.Now().Add(-1 * time.Hour), TenantID: "t1",
	}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{Product: product},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, &mocks.MockNotificationService{}, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 5, AddedBy: "u1"}
	got, err := uc.Execute(ctx, req)
	if got != nil {
		t.Fatalf("Execute() expected nil response, got %+v", got)
	}
	var limitErr domain.ErrStockExceedsLimit
	if err == nil || !errors.As(err, &limitErr) {
		t.Errorf("Execute() err = %v, want ErrStockExceedsLimit", err)
	}
	if product.CurrentStock.Value() != 8 {
		t.Errorf("product stock should remain 8, got %d", product.CurrentStock.Value())
	}
}

func TestAddStockUseCase_Execute_SaveProductFails(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(100), IsActive: true}
	product := &domain.Product{
		ID: "p1", Name: "Product", CurrentStock: mustQuantity(5),
		LastUpdated: time.Now().Add(-1 * time.Hour), TenantID: "t1",
	}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{Product: product, SaveErr: errSaveProduct},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, &mocks.MockNotificationService{}, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 3, AddedBy: "u1"}
	got, err := uc.Execute(ctx, req)
	if got != nil {
		t.Fatalf("Execute() expected nil response, got %+v", got)
	}
	if err == nil || !errors.Is(err, errSaveProduct) {
		t.Errorf("Execute() err = %v, want %v", err, errSaveProduct)
	}
}

func TestAddStockUseCase_Execute_StockHistoryCreateFails(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(100), IsActive: true}
	product := &domain.Product{
		ID: "p1", Name: "Product", CurrentStock: mustQuantity(5),
		LastUpdated: time.Now().Add(-1 * time.Hour), TenantID: "t1",
	}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{Product: product},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{CreateErr: errCreateHistory},
	}
	uc := NewAddStockUseCase(uow, &mocks.MockNotificationService{}, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 3, AddedBy: "u1"}
	got, err := uc.Execute(ctx, req)
	if got != nil {
		t.Fatalf("Execute() expected nil response, got %+v", got)
	}
	if err == nil || !errors.Is(err, errCreateHistory) {
		t.Errorf("Execute() err = %v, want %v", err, errCreateHistory)
	}
}

func TestAddStockUseCase_Execute_Success(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(100), IsActive: true}
	product := &domain.Product{
		ID: "p1", Name: "Widget", CurrentStock: mustQuantity(10),
		LastUpdated: time.Now().Add(-1 * time.Hour), TenantID: "t1",
	}
	hist := &mocks.MockStockHistoryRepo{}
	notif := &mocks.MockNotificationService{}
	pub := &mocks.MockEventPublisher{}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{Product: product},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: hist,
	}
	uc := NewAddStockUseCase(uow, notif, pub)
	ctx := context.Background()

	req := AddStockRequest{
		ProductID: "p1", TenantID: "t1", Quantity: 15,
		AddedBy: "u1", Notes: "restock",
	}
	got, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() err = %v", err)
	}
	if got == nil {
		t.Fatal("Execute() expected non-nil response")
	}

	if got.ProductID != "p1" || got.ProductName != "Widget" {
		t.Errorf("response ProductID/Name = %q / %q, want p1 / Widget", got.ProductID, got.ProductName)
	}
	if got.PreviousStock != 10 || got.NewStock != 25 || got.Added != 15 {
		t.Errorf("response stock: previous=%d new=%d added=%d, want 10, 25, 15",
			got.PreviousStock, got.NewStock, got.Added)
	}
	if got.MaxAllowed != 100 {
		t.Errorf("MaxAllowed = %d, want 100", got.MaxAllowed)
	}
	utilWant := 25.0
	if got.Utilization != utilWant {
		t.Errorf("Utilization = %v, want %v", got.Utilization, utilWant)
	}

	if product.CurrentStock.Value() != 25 {
		t.Errorf("product.CurrentStock = %d, want 25", product.CurrentStock.Value())
	}
	if len(hist.Events) != 1 {
		t.Errorf("StockHistory.Create calls = %d, want 1", len(hist.Events))
	} else {
		e := hist.Events[0]
		if e.ProductID != "p1" || e.TenantID != "t1" || e.AddedBy != "u1" || e.Notes != "restock" {
			t.Errorf("StockAddedEvent: ProductID=%s TenantID=%s AddedBy=%s Notes=%s",
				e.ProductID, e.TenantID, e.AddedBy, e.Notes)
		}
		if e.Previous.Value() != 10 || e.Current.Value() != 25 {
			t.Errorf("StockAddedEvent: Previous=%d Current=%d", e.Previous.Value(), e.Current.Value())
		}
	}
	if len(pub.Published) != 1 {
		t.Errorf("EventPublisher.Publish calls = %d, want 1", len(pub.Published))
	}
}

func TestAddStockUseCase_Execute_Success_NoEventPublisher(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(100), IsActive: true}
	product := &domain.Product{
		ID: "p1", Name: "Widget", CurrentStock: mustQuantity(10),
		LastUpdated: time.Now().Add(-1 * time.Hour), TenantID: "t1",
	}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{Product: product},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, &mocks.MockNotificationService{}, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 5, AddedBy: "u1"}
	got, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() err = %v", err)
	}
	if got == nil || got.NewStock != 15 {
		t.Fatalf("Execute() response: got=%+v, want NewStock=15", got)
	}
}

func TestAddStockUseCase_Execute_Success_HighUtilizationSendsAlert(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(100), IsActive: true}
	product := &domain.Product{
		ID: "p1", Name: "Widget", CurrentStock: mustQuantity(75),
		LastUpdated: time.Now().Add(-1 * time.Hour), TenantID: "t1",
	}
	notif := &mocks.MockNotificationService{}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{Product: product},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, notif, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 10, AddedBy: "u1"}
	_, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() err = %v", err)
	}
	// Utilization 85/100 > 80% -> alert should be sent (async, give it a moment)
	time.Sleep(50 * time.Millisecond)
	if len(notif.StockAlerts) != 1 {
		t.Errorf("SendStockAlert calls = %d, want 1 (utilization > 80%%)", len(notif.StockAlerts))
	} else {
		a := notif.StockAlerts[0]
		if a.ProductID != "p1" || a.Utilization != 85 {
			t.Errorf("alert: ProductID=%s Utilization=%v, want p1 85", a.ProductID, a.Utilization)
		}
	}
}

func TestAddStockUseCase_Execute_Success_LowStockSendsAlert(t *testing.T) {
	tenant := &domain.Tenant{ID: "t1", Name: "Tenant", MaxStock: mustQuantity(100), IsActive: true}
	product := &domain.Product{
		ID: "p1", Name: "Widget", CurrentStock: mustQuantity(5),
		LastUpdated: time.Now().Add(-1 * time.Hour), TenantID: "t1",
	}
	notif := &mocks.MockNotificationService{}
	uow := &mocks.MockUnitOfWork{
		ProductsRepo:  &mocks.MockProductRepo{Product: product},
		TenantsRepo:   &mocks.MockTenantRepo{Tenant: tenant},
		StockHistRepo: &mocks.MockStockHistoryRepo{},
	}
	uc := NewAddStockUseCase(uow, notif, nil)
	ctx := context.Background()

	req := AddStockRequest{ProductID: "p1", TenantID: "t1", Quantity: 2, AddedBy: "u1"}
	_, err := uc.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() err = %v", err)
	}
	// New stock 7 < 10 -> IsLowStock(10) is true, low stock alert sent async
	time.Sleep(50 * time.Millisecond)
	if notif.LowStockCalls != 1 {
		t.Errorf("SendLowStockAlert calls = %d, want 1", notif.LowStockCalls)
	}
}
