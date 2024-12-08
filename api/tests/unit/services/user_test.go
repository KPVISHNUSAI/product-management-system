package tests

import (
	"testing"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
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

func TestCreateUser(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := services.NewUserService(mockRepo, "test-secret")

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
	mockRepo.AssertExpectations(t)
}

func TestValidateCredentials(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := services.NewUserService(mockRepo, "test-secret")

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
}
