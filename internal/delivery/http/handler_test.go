package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"wb_tech_level_zero/internal/config"
	httpapi "wb_tech_level_zero/internal/delivery/http"
	"wb_tech_level_zero/internal/dto"
	"wb_tech_level_zero/internal/orders"
	"wb_tech_level_zero/internal/service"

	"github.com/gorilla/mux"
)

type mockOrderService struct {
	GetOrderByUIDFunc func(ctx context.Context, orderUID string) (*orders.Order, error)
	GetOrdersFunc     func(ctx context.Context, params service.GetOrdersParams) ([]*orders.Order, int, error)
}

func (m *mockOrderService) GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error) {
	return m.GetOrderByUIDFunc(ctx, orderUID)
}

func (m *mockOrderService) GetOrders(ctx context.Context, params service.GetOrdersParams) ([]*orders.Order, int, error) {
	return m.GetOrdersFunc(ctx, params)
}

func TestGetOrderByUID(t *testing.T) {
	cfg := &config.Config{}

	t.Run("success - 200 OK", func(t *testing.T) {
		// Arrange
		mockService := &mockOrderService{
			GetOrderByUIDFunc: func(ctx context.Context, orderUID string) (*orders.Order, error) {
				return &orders.Order{OrderUID: "test-uid"}, nil
			},
		}
		handler := httpapi.NewHandlers(cfg, mockService)
		req := httptest.NewRequest(http.MethodGet, "/order/test-uid", nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/order/{order_uid}", handler.GetOrderByUID)

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}

		var resp dto.OrderDTO
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.OrderUID != "test-uid" {
			t.Errorf("expected order_uid 'test-uid', got '%s'", resp.OrderUID)
		}
	})

	t.Run("not found - 404", func(t *testing.T) {
		// Arrange
		mockService := &mockOrderService{
			GetOrderByUIDFunc: func(ctx context.Context, orderUID string) (*orders.Order, error) {
				return nil, orders.ErrOrderNotFound
			},
		}
		handler := httpapi.NewHandlers(cfg, mockService)
		req := httptest.NewRequest(http.MethodGet, "/order/not-found-uid", nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/order/{order_uid}", handler.GetOrderByUID)

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}

		var errResp dto.ErrorResponse
		if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Message != "Order not found" {
			t.Errorf("expected error message 'Order not found', got '%s'", errResp.Message)
		}
	})

	t.Run("internal server error - 500", func(t *testing.T) {
		// Arrange
		mockService := &mockOrderService{
			GetOrderByUIDFunc: func(ctx context.Context, orderUID string) (*orders.Order, error) {
				return nil, errors.New("database is down")
			},
		}
		handler := httpapi.NewHandlers(cfg, mockService)
		req := httptest.NewRequest(http.MethodGet, "/order/any-uid", nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/order/{order_uid}", handler.GetOrderByUID)

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		var errResp dto.ErrorResponse
		if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Message != "Internal server error" {
			t.Errorf("expected error message 'Internal server error', got '%s'", errResp.Message)
		}
	})
}

func TestGetOrders(t *testing.T) {
	cfg := &config.Config{DefaultPageLimit: 10}

	t.Run("success - 200 OK with params", func(t *testing.T) {
		// Arrange
		mockService := &mockOrderService{
			GetOrdersFunc: func(ctx context.Context, params service.GetOrdersParams) ([]*orders.Order, int, error) {
				if params.Page != 2 || params.Limit != 5 {
					t.Errorf("expected page 2, limit 5, got page %d, limit %d", params.Page, params.Limit)
				}
				return []*orders.Order{{OrderUID: "o1"}, {OrderUID: "o2"}}, 2, nil
			},
		}
		handler := httpapi.NewHandlers(cfg, mockService)
		req := httptest.NewRequest(http.MethodGet, "/orders?page=2&limit=5", nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/orders", handler.GetOrders)

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}

		var resp dto.OrdersResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Orders) != 2 {
			t.Errorf("expected 2 orders, got %d", len(resp.Orders))
		}
		if resp.Total != 2 {
			t.Errorf("expected total 2, got %d", resp.Total)
		}
		if resp.Page != 2 {
			t.Errorf("expected page 2, got %d", resp.Page)
		}
	})

	t.Run("success - 200 OK with default params", func(t *testing.T) {
		// Arrange
		mockService := &mockOrderService{
			GetOrdersFunc: func(ctx context.Context, params service.GetOrdersParams) ([]*orders.Order, int, error) {
				if params.Page != 1 || params.Limit != 10 {
					t.Errorf("expected default page 1, limit 10, got page %d, limit %d", params.Page, params.Limit)
				}
				return []*orders.Order{}, 0, nil
			},
		}
		handler := httpapi.NewHandlers(cfg, mockService)
		req := httptest.NewRequest(http.MethodGet, "/orders?page=invalid&limit=0", nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/orders", handler.GetOrders)

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("internal server error - 500", func(t *testing.T) {
		mockService := &mockOrderService{
			GetOrdersFunc: func(ctx context.Context, params service.GetOrdersParams) ([]*orders.Order, int, error) {
				return nil, 0, errors.New("db is down")
			},
		}
		handler := httpapi.NewHandlers(cfg, mockService)
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/orders", handler.GetOrders)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
	})
}
