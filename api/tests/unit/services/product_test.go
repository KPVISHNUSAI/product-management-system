package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockProductRepo struct {
	mock.Mock
}

func (m *MockProductRepo) Create(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepo) GetByID(id uint) (*models.Product, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepo) GetByUserID(userID uint) ([]models.Product, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepo) GetFilteredProducts(userID uint, minPrice, maxPrice float64, productName string) ([]models.Product, error) {
	args := m.Called(userID, minPrice, maxPrice, productName)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepo) Update(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepo) UpdateProcessingStatus(id uint, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockProductRepo) UpdateCompressedImages(id uint, images pq.StringArray) error {
	args := m.Called(id, images)
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(queue string, data []byte) error {
	args := m.Called(queue, data)
	return args.Error(0)
}

// Test cases
func TestCreateProduct(t *testing.T) {
	mockRepo := new(MockProductRepo)
	mockPublisher := new(MockPublisher)
	mockCache := new(MockCache)
	service := services.NewProductService(mockRepo, mockPublisher, mockCache)

	req := &services.CreateProductRequest{
		UserID:      1,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Images:      []string{"test.jpg"},
	}

	expectedProduct := &models.Product{
		UserID:             req.UserID,
		ProductName:        req.Name,
		ProductDescription: req.Description,
		ProductPrice:       req.Price,
		ProductImages:      pq.StringArray(req.Images),
		ProcessingStatus:   "pending",
	}

	mockRepo.On("Create", mock.AnythingOfType("*models.Product")).Return(nil)
	mockPublisher.On("Publish", "image_processing", mock.Anything).Return(nil)

	product, err := service.CreateProduct(req)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, expectedProduct.ProductName, product.ProductName)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestGetProduct(t *testing.T) {
	mockRepo := new(MockProductRepo)
	mockPublisher := new(MockPublisher)
	mockCache := new(MockCache)
	service := services.NewProductService(mockRepo, mockPublisher, mockCache)

	expectedProduct := &models.Product{
		ID:          1,
		ProductName: "Test Product",
	}

	t.Run("Cache Hit", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, "product:1", mock.AnythingOfType("**models.Product")).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(**models.Product) // Use **models.Product
				*arg = expectedProduct
			}).
			Return(nil)

		product, err := service.GetProduct(1)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, expectedProduct.ProductName, product.ProductName)
	})

	t.Run("Cache Miss", func(t *testing.T) {
		mockCache.On("Get", mock.Anything, "product:1", mock.AnythingOfType("*models.Product")).
			Return(fmt.Errorf("cache miss"))
		mockRepo.On("GetByID", uint(1)).Return(expectedProduct, nil)
		mockCache.On("Set", mock.Anything, "product:1", expectedProduct, mock.Anything).
			Return(nil)

		product, err := service.GetProduct(1)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, expectedProduct.ProductName, product.ProductName)
	})
}

func TestGetFilteredProducts(t *testing.T) {
	mockRepo := new(MockProductRepo)
	mockPublisher := new(MockPublisher)
	mockCache := new(MockCache)
	service := services.NewProductService(mockRepo, mockPublisher, mockCache)

	req := &services.FilterProductsRequest{
		UserID:      1,
		MinPrice:    10.0,
		MaxPrice:    100.0,
		ProductName: "test",
	}

	expectedProducts := []models.Product{
		{ID: 1, ProductName: "Test Product", ProductPrice: 50.0},
	}

	// Cache key for the request
	cacheKey := fmt.Sprintf("list:%d:minPrice:%f:maxPrice:%f:productName:%s",
		req.UserID, req.MinPrice, req.MaxPrice, req.ProductName)

	// Simulate a cache miss
	mockCache.On("Get", mock.Anything, cacheKey, mock.AnythingOfType("*[]models.Product")).
		Return(fmt.Errorf("cache miss"))

	// Simulate database fetch after cache miss
	mockRepo.On("GetFilteredProducts", req.UserID, req.MinPrice, req.MaxPrice, req.ProductName).
		Return(expectedProducts, nil)

	// Simulate setting the cache after database fetch
	mockCache.On("Set", mock.Anything, cacheKey, expectedProducts, mock.Anything).
		Return(nil)

	products, err := service.GetFilteredProducts(req)

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, expectedProducts[0].ProductName, products[0].ProductName)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
