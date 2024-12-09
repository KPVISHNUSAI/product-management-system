// api/tests/unit/handlers/auth_test.go
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

type MockUserService struct {
	mock.Mock
}

// Implement the UserService interface
func (m *MockUserService) CreateUser(req *services.CreateUserRequest) (*models.AppUser, error) {
	args := m.Called(req)
	return args.Get(0).(*models.AppUser), args.Error(1)
}

func (m *MockUserService) ValidateCredentials(email, password string) (*models.AppUser, error) {
	args := m.Called(email, password)
	return args.Get(0).(*models.AppUser), args.Error(1)
}

func (m *MockUserService) GenerateToken(user *models.AppUser) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func setupAuthTestRouter() (*gin.Engine, *MockUserService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockUserService)
	handler := handlers.NewAuthHandler(mockService)

	auth := router.Group("/api/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
	}

	return router, mockService
}

func TestRegister(t *testing.T) {
	router, mockService := setupAuthTestRouter()

	t.Run("Successful Registration", func(t *testing.T) {
		req := services.CreateUserRequest{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		expectedUser := &models.AppUser{
			ID:       1,
			Email:    req.Email,
			Name:     req.Name,
			Password: "hashed_password",
		}

		mockService.On("CreateUser", &req).Return(expectedUser, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.AppUser
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.Email, response.Email)
		assert.Equal(t, expectedUser.Name, response.Name)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer([]byte("invalid json")))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		req := services.CreateUserRequest{
			Email:    "existing@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		mockService.On("CreateUser", &req).Return(
			(*models.AppUser)(nil),
			fmt.Errorf("user with email already exists"),
		)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestLogin(t *testing.T) {
	router, mockService := setupAuthTestRouter()

	t.Run("Successful Login", func(t *testing.T) {
		req := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{
			Email:    "test@example.com",
			Password: "password123",
		}

		user := &models.AppUser{
			ID:    1,
			Email: req.Email,
		}

		expectedToken := "jwt.token.here"

		mockService.On("ValidateCredentials", req.Email, req.Password).Return(user, nil)
		mockService.On("GenerateToken", user).Return(expectedToken, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expectedToken, response["token"])
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		req := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		mockService.On("ValidateCredentials", req.Email, req.Password).Return(
			(*models.AppUser)(nil),
			fmt.Errorf("invalid credentials"),
		)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Token Generation Failure", func(t *testing.T) {
		// Create fresh instances for this test case
		router := gin.New()
		mockService := new(MockUserService)
		handler := handlers.NewAuthHandler(mockService)

		// Setup route
		auth := router.Group("/api/auth")
		auth.POST("/login", handler.Login)

		// Test request data
		loginRequest := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{
			Email:    "test@example.com",
			Password: "password123",
		}

		// Test user
		user := &models.AppUser{
			ID:    1,
			Email: loginRequest.Email,
		}

		// Setup mock expectations
		mockService.On("ValidateCredentials", loginRequest.Email, loginRequest.Password).
			Return(user, nil).Once()

		mockService.On("GenerateToken", mock.MatchedBy(func(u *models.AppUser) bool {
			return u.ID == user.ID && u.Email == user.Email
		})).Return("", fmt.Errorf("token generation failed")).Once()

		// Perform request
		body, _ := json.Marshal(loginRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, r)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected internal server error status code")

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should return valid JSON response")
		assert.Equal(t, "failed to generate token", response["error"], "Should return correct error message")

		// Verify mock expectations
		mockService.AssertExpectations(t)
	})

}
