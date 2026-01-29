// internal/domain/errors.go
package domain

import (
	"errors"
	"fmt"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrTenantNotFound   = errors.New("tenant not found")
	ErrTenantInactive   = errors.New("tenant is inactive")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrInvalidProductID = errors.New("invalid product id")
)

type ErrStockExceedsLimit struct {
	Current    int
	Adding     int
	WouldBe    int
	MaxAllowed int
}

func (e ErrStockExceedsLimit) Error() string {
	return fmt.Sprintf(
		"cannot exceed max stock of %d. Current: %d, Adding: %d, Would be: %d",
		e.MaxAllowed, e.Current, e.Adding, e.WouldBe,
	)
}
