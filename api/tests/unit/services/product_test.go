package tests

import (
	"context"
	"testing"
	"time"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ProductRepository interface {
	Create(product *models.Product) error
	GetByID(id uint) (*models.Product, error)
	GetByUserID(userID uint) ([]models.Product, error)
	Update(product *models.Product) error
	UpdateProcessingStatus(id uint, status string) error
	UpdateCompressedImages(id uint, images pq.StringArray) error
}

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

type MockProductRepo struct {
	mock.Mock
}

type MockPublisher struct {
	mock.Mock
}

type MockCache struct {
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

// Implement Cache interface
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

// Implement Publisher interface
func (m *MockPublisher) Publish(queue string, data []byte) error {
	args := m.Called(queue, data)
	return args.Error(0)
}

func TestCreateProduct(t *testing.T) {
	mockRepo := new(MockProductRepo)
	mockPublisher := new(MockPublisher)
	mockCache := new(MockCache)

	service := services.NewProductService(mockRepo, mockPublisher, mockCache)

	req := &services.CreateProductRequest{
		UserID: 1,
		Name:   "Test Product",
		Price:  99.99,
	}

	mockRepo.On("Create", mock.AnythingOfType("*models.Product")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	product, err := service.CreateProduct(req)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, req.Name, product.ProductName)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
