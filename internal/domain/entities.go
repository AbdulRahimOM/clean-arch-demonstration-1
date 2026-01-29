// internal/domain/entities.go
package domain

import (
	"errors"
	"time"
)

// Value Objects
type StockQuantity struct {
	value int
}

func NewStockQuantity(q int) (StockQuantity, error) {
	if q < 0 {
		return StockQuantity{}, errors.New("quantity cannot be negative")
	}
	return StockQuantity{value: q}, nil
}

func (q StockQuantity) Value() int {
	return q.value
}

func (q StockQuantity) Add(other StockQuantity) StockQuantity {
	return StockQuantity{value: q.value + other.value}
}

func (q StockQuantity) Exceeds(limit StockQuantity) bool {
	return q.value > limit.value
}

// Domain Entity with business behavior
type Product struct {
	ID           string
	Name         string
	CurrentStock StockQuantity
	LastUpdated  time.Time
	TenantID     string
}

func (p *Product) AddStock(quantity StockQuantity, maxLimit StockQuantity) error {
	newStock := p.CurrentStock.Add(quantity)
	
	if newStock.Exceeds(maxLimit) {
		return ErrStockExceedsLimit{
			Current:    p.CurrentStock.Value(),
			Adding:     quantity.Value(),
			WouldBe:    newStock.Value(),
			MaxAllowed: maxLimit.Value(),
		}
	}
	
	p.CurrentStock = newStock
	p.LastUpdated = time.Now()
	return nil
}

func (p *Product) IsRecentlyUpdated(threshold time.Duration) bool {
	return time.Since(p.LastUpdated) < threshold
}

func (p *Product) IsLowStock(threshold int) bool {
	return p.CurrentStock.Value() < threshold
}

func (p *Product) UtilizationPercentage(maxLimit StockQuantity) float64 {
	if maxLimit.Value() == 0 {
		return 0
	}
	return float64(p.CurrentStock.Value()) / float64(maxLimit.Value()) * 100
}

type Tenant struct {
	ID       string
	Name     string
	MaxStock StockQuantity
	IsActive bool
}

func (t *Tenant) CanReceiveStock() error {
	if !t.IsActive {
		return ErrTenantInactive
	}
	return nil
}

// Domain Events
type StockAddedEvent struct {
	ProductID    string
	TenantID     string
	Quantity     StockQuantity
	Previous     StockQuantity
	Current      StockQuantity
	AddedBy      string
	Timestamp    time.Time
	Notes        string
}

type StockLimitAlertEvent struct {
	ProductID   string
	ProductName string
	Current     StockQuantity
	MaxLimit    StockQuantity
	Utilization float64
	TenantID    string
	Timestamp   time.Time
}