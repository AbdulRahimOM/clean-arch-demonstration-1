// internal/interfaces/http/dto.go
package http

// HTTP Request DTO
type AddStockRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
	TenantID  string `json:"tenant_id" validate:"required"`
	Notes     string `json:"notes"`
}

// HTTP Response DTO
type AddStockResponse struct {
	Success      bool    `json:"success"`
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Previous     int     `json:"previous_stock"`
	NewStock     int     `json:"new_stock"`
	Added        int     `json:"added"`
	MaxAllowed   int     `json:"max_allowed"`
	Utilization  float64 `json:"utilization_percentage"`
	Message      string  `json:"message"`
	Timestamp    string  `json:"timestamp"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}