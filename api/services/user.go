package services

import (
	"fmt"
	"time"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(user *models.AppUser) error
	GetByEmail(email string) (*models.AppUser, error)
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type UserService struct {
	userRepo  UserRepository
	jwtSecret string
}

func NewUserService(repo UserRepository, jwtSecret string) *UserService {
	return &UserService{
		userRepo:  repo,
		jwtSecret: jwtSecret,
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(bytes), nil
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*models.AppUser, error) {
	// Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to hash password: %v", err)
	}

	// Create user object
	user := &models.AppUser{
		Email:    req.Email,
		Name:     req.Name,
		Password: hashedPassword,
	}

	// Store the user in the database
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Return the user without the password field
	user.Password = "" // Ensure password is not exposed
	return user, nil
}

func (s *UserService) GenerateToken(user *models.AppUser) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *UserService) ValidateCredentials(email, password string) (*models.AppUser, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}

	return user, nil
}
