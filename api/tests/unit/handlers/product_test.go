package tests

import (
	"bytes"
	"encoding/json"
	"errors"
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

func TestCreateProductHandler(t *testing.T) {
	mockService := new(MockProductService)
	handler := handlers.NewProductHandler(mockService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/products", handler.CreateProduct)

	t.Run("Success", func(t *testing.T) {
		req := services.CreateProductRequest{
			UserID: 1,
			Name:   "Test Product",
			Price:  99.99,
		}

		expectedProduct := &models.Product{
			UserID:       req.UserID,
			ProductName:  req.Name,
			ProductPrice: req.Price,
		}

		mockService.On("CreateProduct", &req).Return(expectedProduct, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/products", bytes.NewBuffer([]byte("invalid json")))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetProductHandler(t *testing.T) {
	mockService := new(MockProductService)
	handler := handlers.NewProductHandler(mockService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/products/:id", handler.GetProduct)

	t.Run("Success", func(t *testing.T) {
		product := &models.Product{
			ID:          1,
			ProductName: "Test Product",
		}

		mockService.On("GetProduct", uint(1)).Return(product, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/products/1", nil)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("GetProduct", uint(999)).Return(nil, errors.New("not found"))

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/products/999", nil)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
