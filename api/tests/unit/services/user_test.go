// api/tests/unit/services/user_test.go
package tests

import (
	"fmt"
	"testing"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(user *models.AppUser) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByEmail(email string) (*models.AppUser, error) {
	args := m.Called(email)
	return args.Get(0).(*models.AppUser), args.Error(1)
}

func (m *MockUserRepo) GetByID(id uint) (*models.AppUser, error) {
	args := m.Called(id)
	return args.Get(0).(*models.AppUser), args.Error(1)
}

func TestCreateUser(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := services.NewUserService(mockRepo, "test-secret")

	t.Run("Successful User Creation", func(t *testing.T) {
		req := &services.CreateUserRequest{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		mockRepo.On("Create", mock.AnythingOfType("*models.AppUser")).Return(nil)

		user, err := service.CreateUser(req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Name, user.Name)
		assert.NotEqual(t, req.Password, user.Password) // Password should be hashed
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed User Creation", func(t *testing.T) {
		req := &services.CreateUserRequest{
			Email:    "existing@example.com",
			Name:     "Test User",
			Password: "password123",
		}

		mockRepo.On("Create", mock.AnythingOfType("*models.AppUser")).
			Return(fmt.Errorf("email already exists"))

		user, err := service.CreateUser(req)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestValidateCredentials(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := services.NewUserService(mockRepo, "test-secret")

	t.Run("Valid Credentials", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		mockUser := &models.AppUser{
			Email:    email,
			Password: string(hashedPassword),
		}

		mockRepo.On("GetByEmail", email).Return(mockUser, nil)

		user, err := service.ValidateCredentials(email, password)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		email := "test@example.com"
		password := "wrongpassword"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

		mockUser := &models.AppUser{
			Email:    email,
			Password: string(hashedPassword),
		}

		mockRepo.On("GetByEmail", email).Return(mockUser, nil)

		user, err := service.ValidateCredentials(email, password)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		email := "nonexistent@example.com"
		password := "password123"

		mockRepo.On("GetByEmail", email).
			Return((*models.AppUser)(nil), fmt.Errorf("user not found"))

		user, err := service.ValidateCredentials(email, password)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestGenerateToken(t *testing.T) {
	service := services.NewUserService(new(MockUserRepo), "test-secret")

	t.Run("Successful Token Generation", func(t *testing.T) {
		user := &models.AppUser{
			ID:    1,
			Email: "test@example.com",
		}

		token, err := service.GenerateToken(user)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token content
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, float64(user.ID), claims["user_id"])
		assert.Equal(t, user.Email, claims["email"])
	})
}
