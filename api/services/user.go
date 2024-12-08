package services

import (
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
	return string(bytes), err
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*models.AppUser, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.AppUser{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

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
