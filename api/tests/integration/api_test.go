// api/tests/integration/api_test.go
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/KPVISHNUSAI/product-management-system/api/config"
	"github.com/KPVISHNUSAI/product-management-system/api/handlers"
	"github.com/KPVISHNUSAI/product-management-system/api/middleware"
	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/KPVISHNUSAI/product-management-system/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type IntegrationTestSuite struct {
	db     *gorm.DB
	router *gin.Engine
	token  string
}

type TestCache struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func (c *TestCache) Get(ctx context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if data, ok := c.data[key]; ok {
		return json.Unmarshal(data, dest)
	}
	return fmt.Errorf("cache miss")
}

func (c *TestCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.data[key] = data
	return nil
}

func (c *TestCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

func setupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Ensure test database configuration
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == "" {
		cfg.Database.Port = "5431"
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "postgres"
	}
	if cfg.Database.Password == "" {
		cfg.Database.Password = "postgres"
	}

	db, err := database.NewPostgresDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		"product_management_test",
	)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Ensure tables are created
	db.AutoMigrate(&models.AppUser{}, &models.Product{})

	gin.SetMode(gin.TestMode)
	router := setupRouter(db, cfg)

	return &IntegrationTestSuite{
		db:     db,
		router: router,
	}
}

func (s *IntegrationTestSuite) cleanup() {
	// Clean up test data
	s.db.Exec("DELETE FROM app_products")
	s.db.Exec("DELETE FROM app_users")
}

func setupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	productRepo := postgres.NewProductRepository(db)

	// Initialize mock message queue publisher
	mockPublisher := &MockPublisher{
		mock.Mock{},
	}

	// Initialize test cache
	testCache := &TestCache{
		data: make(map[string][]byte),
	}

	// Initialize services with mocks
	userService := services.NewUserService(userRepo, "test-secret")
	productService := services.NewProductService(productRepo, mockPublisher, testCache)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	productHandler := handlers.NewProductHandler(productService)

	// Setup routes
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		products := api.Group("/products")
		products.Use(middleware.AuthMiddleware("test-secret"))
		{
			products.POST("", productHandler.CreateProduct)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("", productHandler.GetUserProducts)
			products.GET("/filter", productHandler.GetFilteredProducts)
		}
	}

	return router
}

// Add MockPublisher implementation
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(queue string, data []byte) error {
	args := m.Called(queue, data)
	return args.Error(0)
}

func TestUserRegistrationAndAuthentication(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	t.Run("User Registration Flow", func(t *testing.T) {
		// Register user
		registerBody := map[string]interface{}{
			"email":    "test@example.com",
			"name":     "Test User",
			"password": "password123",
		}
		// Perform the registration request
		resp := performRequest(suite.router, "POST", "/api/auth/register", registerBody)

		// Log the response to check if email is returned correctly
		fmt.Println("Registration response:", resp)

		// Extract the user email from the response and assert it
		var registrationResponse struct {
			Email string `json:"email"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &registrationResponse); err != nil {
			t.Fatal("Failed to unmarshal registration response:", err)
		}

		// Check if the email is as expected
		if registrationResponse.Email != "test@example.com" {
			t.Errorf("Expected email to be 'test@example.com', got %v", registrationResponse.Email)
		}
	})

	t.Run("User Login Flow", func(t *testing.T) {
		loginBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}
		w := performRequest(suite.router, "POST", "/api/auth/login", loginBody)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["token"])

		suite.token = response["token"] // Save token for subsequent tests
	})
}

// TestProductOperations
func TestProductOperations(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Create test user and get token
	suite.createTestUser(t)

	// Get user ID from database
	var user models.AppUser
	err := suite.db.Where("email = ?", "test@example.com").First(&user).Error
	assert.NoError(t, err, "Failed to retrieve test user")

	t.Run("Product Creation and Retrieval", func(t *testing.T) {
		createBody := map[string]interface{}{
			"user_id":             user.ID, // Use actual user ID
			"product_name":        "Test Product",
			"product_description": "Test Description",
			"product_price":       99.99,
			"product_images":      []string{"test.jpg"},
		}

		w := performAuthorizedRequest(suite.router, "POST", "/api/products", createBody, suite.token)
		assert.Equal(t, http.StatusCreated, w.Code)

		var product models.Product
		err := json.Unmarshal(w.Body.Bytes(), &product)
		assert.NoError(t, err)
		assert.Equal(t, "Test Product", product.ProductName)
		assert.Equal(t, user.ID, product.UserID)

		// Test retrieval
		w = performAuthorizedRequest(suite.router, "GET", fmt.Sprintf("/api/products/%d", product.ID), nil, suite.token)
		assert.Equal(t, http.StatusOK, w.Code)

		var retrievedProduct models.Product
		err = json.Unmarshal(w.Body.Bytes(), &retrievedProduct)
		assert.NoError(t, err)
		assert.Equal(t, product.ID, retrievedProduct.ID)
	})

	t.Run("Product Filtering", func(t *testing.T) {
		// Create test products with correct user ID
		products := []models.Product{
			{
				UserID:       user.ID,
				ProductName:  "Test Product 1",
				ProductPrice: 75.99,
			},
			{
				UserID:       user.ID,
				ProductName:  "Test Product 2",
				ProductPrice: 125.99,
			},
		}

		for _, p := range products {
			result := suite.db.Create(&p)
			assert.NoError(t, result.Error, "Failed to create test product")
		}

		w := performAuthorizedRequest(
			suite.router,
			"GET",
			fmt.Sprintf("/api/products/filter?user_id=%d&min_price=50&max_price=150&product_name=Test", user.ID),
			nil,
			suite.token,
		)
		assert.Equal(t, http.StatusOK, w.Code)

		var filteredProducts []models.Product
		err := json.Unmarshal(w.Body.Bytes(), &filteredProducts)
		assert.NoError(t, err)
		assert.NotEmpty(t, filteredProducts)

		for _, p := range filteredProducts {
			assert.Equal(t, user.ID, p.UserID)
			assert.True(t, p.ProductPrice >= 50 && p.ProductPrice <= 150)
			assert.Contains(t, p.ProductName, "Test")
		}
	})
}

// Helper functions

func performRequest(r http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performAuthorizedRequest(r http.Handler, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func (s *IntegrationTestSuite) createTestUser(t *testing.T) {
	// Clean existing data
	s.cleanup()

	// Create user through the registration endpoint
	registerBody := map[string]interface{}{
		"email":    "test@example.com",
		"name":     "Test User",
		"password": "password123",
	}

	w := performRequest(s.router, "POST", "/api/auth/register", registerBody)
	assert.Equal(t, http.StatusCreated, w.Code, "Failed to create test user")

	// Verify user was created
	var user models.AppUser
	result := s.db.Where("email = ?", registerBody["email"]).First(&user)
	assert.NoError(t, result.Error, "User should exist in database")

	// Login to get token
	loginBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}

	w = performRequest(s.router, "POST", "/api/auth/login", loginBody)
	assert.Equal(t, http.StatusOK, w.Code, "Failed to login test user")

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal login response")
	assert.NotEmpty(t, response["token"], "Token should not be empty")

	s.token = response["token"]
}

func (s *IntegrationTestSuite) createTestProducts(t *testing.T) {
	products := []models.Product{
		{
			UserID:       1,
			ProductName:  "Test Product 1",
			ProductPrice: 75.99,
		},
		{
			UserID:       1,
			ProductName:  "Test Product 2",
			ProductPrice: 125.99,
		},
	}

	for _, p := range products {
		s.db.Create(&p)
	}
}
