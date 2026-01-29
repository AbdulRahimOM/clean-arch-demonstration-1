// internal/interfaces/http/handler.go
package http

import (
	"context"
	"log"
	"time"

	"myapp/internal/application/usecases"
	"myapp/internal/domain"

	"github.com/gofiber/fiber/v2"
)

type StockHandler struct {
	addStockUseCase usecases.AddStockUseCase
}

func NewStockHandler(addStockUseCase usecases.AddStockUseCase) *StockHandler {
	return &StockHandler{
		addStockUseCase: addStockUseCase,
	}
}

// Clean Architecture Handler - Only deals with HTTP concerns
func (h *StockHandler) AddStock(c *fiber.Ctx) error {
	// 1. Parse HTTP request
	var req AddStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Invalid request format",
		})
	}

	// 2. Get user from context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	// 3. Convert HTTP DTO to Application DTO
	appReq := usecases.AddStockRequest{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		TenantID:  req.TenantID,
		Notes:     req.Notes,
		AddedBy:   userID,
	}

	// 4. Call use case (business logic)
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	response, err := h.addStockUseCase.Execute(ctx, appReq)
	if err != nil {
		return h.handleError(c, err)
	}

	// 5. Convert Application Response to HTTP Response
	resp := AddStockResponse{
		Success:     true,
		ProductID:   response.ProductID,
		ProductName: response.ProductName,
		Previous:    response.PreviousStock,
		NewStock:    response.NewStock,
		Added:       response.Added,
		MaxAllowed:  response.MaxAllowed,
		Utilization: response.Utilization,
		Message:     "Stock updated successfully",
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	// 6. Return HTTP response
	return c.Status(200).JSON(resp)
}

func (h *StockHandler) handleError(c *fiber.Ctx, err error) error {
	// Map domain errors to HTTP status codes
	switch err.(type) {
	case domain.ErrStockExceedsLimit:
		return c.Status(400).JSON(ErrorResponse{
			Error: err.Error(),
			Code:  "STOCK_LIMIT_EXCEEDED",
		})
	}

	// Map other domain errors
	switch err {
	case domain.ErrProductNotFound:
		return c.Status(404).JSON(ErrorResponse{
			Error: "Product not found",
			Code:  "PRODUCT_NOT_FOUND",
		})
	case domain.ErrTenantNotFound:
		return c.Status(404).JSON(ErrorResponse{
			Error: "Tenant not found",
			Code:  "TENANT_NOT_FOUND",
		})
	case domain.ErrTenantInactive:
		return c.Status(400).JSON(ErrorResponse{
			Error: "Tenant is inactive",
			Code:  "TENANT_INACTIVE",
		})
	case domain.ErrInvalidQuantity:
		return c.Status(400).JSON(ErrorResponse{
			Error: "Quantity must be positive",
			Code:  "INVALID_QUANTITY",
		})
	default:
		// Log internal errors but don't expose details
		log.Printf("Internal error: %v", err)
		return c.Status(500).JSON(ErrorResponse{
			Error: "Internal server error",
		})
	}
}
