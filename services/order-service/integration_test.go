// +build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/shopsphere/order-service/internal/handlers"
	"github.com/shopsphere/order-service/internal/repository"
	"github.com/shopsphere/order-service/internal/service"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// MockProductService for integration testing
type MockProductService struct{}

func (m *MockProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	return &models.Product{
		ID:          id,
		SKU:         "TEST-SKU",
		Name:        "Test Product",
		Description: "Test product description",
		Price:       decimal.NewFromFloat(99.99),
		Stock:       10,
		Images:      []string{"test-image.jpg"},
		Attributes: models.ProductAttributes{
			Brand: "TestBrand",
			Color: "Red",
			Size:  "M",
			Custom: map[string]interface{}{
				"material": "cotton",
			},
		},
	}, nil
}

func (m *MockProductService) ValidateStock(ctx context.Context, productID string, quantity int) error {
	if quantity > 10 {
		return fmt.Errorf("insufficient stock")
	}
	return nil
}

// MockInventoryService for integration testing
type MockInventoryService struct{}

func (m *MockInventoryService) ReserveStock(ctx context.Context, productID string, quantity int) error {
	return nil
}

func (m *MockInventoryService) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	return nil
}

func TestOrderServiceIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup in-memory database for testing
	db := setupTestDB(t)
	defer db.Close()

	// Initialize components
	repo := repository.NewPostgresOrderRepository(db)
	productService := &MockProductService{}
	inventoryService := &MockInventoryService{}
	orderService := service.NewOrderService(repo, productService, inventoryService)
	handler := handlers.NewOrderHandler(orderService)

	// Setup router
	router := mux.NewRouter()
	router.HandleFunc("/orders", handler.CreateOrder).Methods("POST")
	router.HandleFunc("/orders/{id}", handler.GetOrder).Methods("GET")
	router.HandleFunc("/orders/{id}/status", handler.UpdateOrderStatus).Methods("PUT")
	router.HandleFunc("/orders/{id}/cancel", handler.CancelOrder).Methods("PUT")

	// Test creating an order
	t.Run("CreateOrder", func(t *testing.T) {
		createReq := service.CreateOrderRequest{
			UserID: "user123",
			Items: []service.OrderItemRequest{
				{
					ProductID: "prod1",
					Quantity:  2,
					Price:     decimal.NewFromFloat(99.99),
				},
			},
			ShippingAddress: models.Address{
				Street:     "123 Test St",
				City:       "Test City",
				State:      "TS",
				PostalCode: "12345",
				Country:    "US",
			},
			BillingAddress: models.Address{
				Street:     "123 Test St",
				City:       "Test City",
				State:      "TS",
				PostalCode: "12345",
				Country:    "US",
			},
			PaymentMethod: models.PaymentMethod{
				Type:  "card",
				Last4: "1234",
				Brand: "visa",
			},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", rr.Code)
		}

		var order models.Order
		if err := json.Unmarshal(rr.Body.Bytes(), &order); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if order.ID == "" {
			t.Error("Expected order ID to be set")
		}

		if order.Status != models.OrderPending {
			t.Errorf("Expected status pending, got %s", order.Status)
		}

		// Test getting the order
		t.Run("GetOrder", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/orders/"+order.ID, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
			}

			var retrievedOrder models.Order
			if err := json.Unmarshal(rr.Body.Bytes(), &retrievedOrder); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if retrievedOrder.ID != order.ID {
				t.Errorf("Expected order ID %s, got %s", order.ID, retrievedOrder.ID)
			}
		})

		// Test updating order status
		t.Run("UpdateOrderStatus", func(t *testing.T) {
			statusReq := map[string]interface{}{
				"status":     "confirmed",
				"reason":     "Payment confirmed",
				"changed_by": "system",
			}

			body, _ := json.Marshal(statusReq)
			req := httptest.NewRequest("PUT", "/orders/"+order.ID+"/status", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
			}
		})

		// Test cancelling order
		t.Run("CancelOrder", func(t *testing.T) {
			cancelReq := map[string]interface{}{
				"reason":     "Customer request",
				"cancelled_by": "user123",
			}

			body, _ := json.Marshal(cancelReq)
			req := httptest.NewRequest("PUT", "/orders/"+order.ID+"/cancel", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
			}
		})
	})
}

func setupTestDB(t *testing.T) *utils.Database {
	// This would typically set up an in-memory database or test database
	// For now, we'll return nil and skip actual database operations
	// In a real integration test, you'd set up a test PostgreSQL instance
	t.Skip("Database setup not implemented for integration test")
	return nil
}
