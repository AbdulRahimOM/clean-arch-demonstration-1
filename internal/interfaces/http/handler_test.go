package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"myapp/internal/application/usecases"
	"myapp/internal/domain"
	httphandler "myapp/internal/interfaces/http"
	"myapp/internal/testutil/httputil"
)

var errInternal = errors.New("internal failure")

// mockAddStockUseCase implements usecases.AddStockUseCase for handler tests.
// Defined here to avoid import cycle (testutil/mocks cannot import usecases).
type mockAddStockUseCase struct {
	response *usecases.AddStockResponse
	err      error
}

func (m *mockAddStockUseCase) Execute(ctx context.Context, req usecases.AddStockRequest) (*usecases.AddStockResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

const testUserID = "test-user-123"

func setupAddStockApp(uc usecases.AddStockUseCase) *fiber.App {
	app := fiber.New()
	app.Use(httputil.UserIDMiddleware(testUserID))
	handler := httphandler.NewStockHandler(uc)
	app.Post("/api/v1/stock/add", handler.AddStock)
	return app
}

func TestStockHandler_AddStock_Success(t *testing.T) {
	uc := &mockAddStockUseCase{
		response: &usecases.AddStockResponse{
			ProductID:     "p1",
			ProductName:   "Widget",
			PreviousStock: 10,
			NewStock:      25,
			Added:         15,
			MaxAllowed:    100,
			Utilization:   25,
		},
	}
	app := setupAddStockApp(uc)

	body := map[string]interface{}{
		"product_id": "p1",
		"quantity":   15,
		"tenant_id":  "t1",
		"notes":      "restock",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var result httphandler.AddStockResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !result.Success || result.ProductID != "p1" || result.ProductName != "Widget" {
		t.Errorf("response: success=%v product_id=%s product_name=%s", result.Success, result.ProductID, result.ProductName)
	}
	if result.Previous != 10 || result.NewStock != 25 || result.Added != 15 {
		t.Errorf("response: previous=%d new_stock=%d added=%d", result.Previous, result.NewStock, result.Added)
	}
}

func TestStockHandler_AddStock_InvalidBody(t *testing.T) {
	uc := &mockAddStockUseCase{}
	app := setupAddStockApp(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/add", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var errResp httphandler.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if errResp.Error != "Invalid request format" {
		t.Errorf("error = %q", errResp.Error)
	}
}

func TestStockHandler_AddStock_ErrProductNotFound(t *testing.T) {
	uc := &mockAddStockUseCase{err: domain.ErrProductNotFound}
	app := setupAddStockApp(uc)

	body := map[string]interface{}{"product_id": "p1", "quantity": 5, "tenant_id": "t1"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Code != "PRODUCT_NOT_FOUND" {
		t.Errorf("code = %q", errResp.Code)
	}
}

func TestStockHandler_AddStock_ErrTenantNotFound(t *testing.T) {
	uc := &mockAddStockUseCase{err: domain.ErrTenantNotFound}
	app := setupAddStockApp(uc)

	body := map[string]interface{}{"product_id": "p1", "quantity": 5, "tenant_id": "t1"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Code != "TENANT_NOT_FOUND" {
		t.Errorf("code = %q", errResp.Code)
	}
}

func TestStockHandler_AddStock_ErrTenantInactive(t *testing.T) {
	uc := &mockAddStockUseCase{err: domain.ErrTenantInactive}
	app := setupAddStockApp(uc)

	body := map[string]interface{}{"product_id": "p1", "quantity": 5, "tenant_id": "t1"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Code != "TENANT_INACTIVE" {
		t.Errorf("code = %q", errResp.Code)
	}
}

func TestStockHandler_AddStock_ErrInvalidQuantity(t *testing.T) {
	uc := &mockAddStockUseCase{err: domain.ErrInvalidQuantity}
	app := setupAddStockApp(uc)

	body := map[string]interface{}{"product_id": "p1", "quantity": 5, "tenant_id": "t1"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Code != "INVALID_QUANTITY" {
		t.Errorf("code = %q", errResp.Code)
	}
}

func TestStockHandler_AddStock_ErrStockExceedsLimit(t *testing.T) {
	uc := &mockAddStockUseCase{
		err: domain.ErrStockExceedsLimit{
			Current:    10,
			Adding:     5,
			WouldBe:    15,
			MaxAllowed: 12,
		},
	}
	app := setupAddStockApp(uc)

	body := map[string]interface{}{"product_id": "p1", "quantity": 5, "tenant_id": "t1"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Code != "STOCK_LIMIT_EXCEEDED" {
		t.Errorf("code = %q", errResp.Code)
	}
}

func TestStockHandler_AddStock_InternalError(t *testing.T) {
	uc := &mockAddStockUseCase{err: errInternal}
	app := setupAddStockApp(uc)

	body := map[string]interface{}{"product_id": "p1", "quantity": 5, "tenant_id": "t1"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/add", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Error != "Internal server error" {
		t.Errorf("error = %q", errResp.Error)
	}
}
