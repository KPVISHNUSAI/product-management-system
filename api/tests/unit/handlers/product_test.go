// api/tests/unit/handlers/product_test.go
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KPVISHNUSAI/product-management-system/api/handlers"
	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(req *services.CreateProductRequest) (*models.Product, error) {
	args := m.Called(req)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) GetProduct(id uint) (*models.Product, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) GetUserProducts(userID uint) ([]models.Product, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductService) GetFilteredProducts(req *services.FilterProductsRequest) ([]models.Product, error) {
	args := m.Called(req)
	return args.Get(0).([]models.Product), args.Error(1)
}

func setupTestRouter() (*gin.Engine, *MockProductService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockProductService)
	handler := handlers.NewProductHandler(mockService)

	// Setup routes
	products := router.Group("/api/products")
	{
		products.POST("/", handler.CreateProduct)
		products.GET("/:id", handler.GetProduct)
		products.GET("/", handler.GetUserProducts)
		products.GET("/filter", handler.GetFilteredProducts)
	}

	return router, mockService
}

func TestCreateProduct(t *testing.T) {
	router, mockService := setupTestRouter()

	t.Run("Successful Product Creation", func(t *testing.T) {
		req := services.CreateProductRequest{
			UserID:      1,
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			Images:      []string{"test.jpg"},
		}

		expectedProduct := &models.Product{
			ID:                 1,
			UserID:             req.UserID,
			ProductName:        req.Name,
			ProductDescription: req.Description,
			ProductPrice:       req.Price,
		}

		mockService.On("CreateProduct", &req).Return(expectedProduct, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/products/", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Product
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.ID, response.ID)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/products/", bytes.NewBuffer([]byte("invalid json")))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetProduct(t *testing.T) {
	router, mockService := setupTestRouter()

	t.Run("Successful Product Retrieval", func(t *testing.T) {
		product := &models.Product{
			ID:          1,
			ProductName: "Test Product",
		}

		mockService.On("GetProduct", uint(1)).Return(product, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/products/1", nil)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Product
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, product.ID, response.ID)
		mockService.AssertExpectations(t)
	})

	t.Run("Product Not Found", func(t *testing.T) {
		mockService.On("GetProduct", uint(999)).Return((*models.Product)(nil),
			fmt.Errorf("not found"))

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/products/999", nil)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetFilteredProducts(t *testing.T) {
	router, mockService := setupTestRouter()

	t.Run("Successful Filtered Products", func(t *testing.T) {
		req := &services.FilterProductsRequest{
			UserID:      1,
			MinPrice:    10.0,
			MaxPrice:    100.0,
			ProductName: "test",
		}

		expectedProducts := []models.Product{
			{ID: 1, ProductName: "Test Product", ProductPrice: 50.0},
		}

		mockService.On("GetFilteredProducts", req).Return(expectedProducts, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/products/filter?user_id=1&min_price=10.0&max_price=100.0&product_name=test", nil)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []models.Product
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 1)
		assert.Equal(t, expectedProducts[0].ID, response[0].ID)
		mockService.AssertExpectations(t)
	})
}
