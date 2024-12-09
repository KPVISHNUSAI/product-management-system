package handlers

import (
	"log"
	"net/http"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	CreateUser(req *services.CreateUserRequest) (*models.AppUser, error)
	ValidateCredentials(email, password string) (*models.AppUser, error)
	GenerateToken(user *models.AppUser) (string, error)
}

type AuthHandler struct {
	userService UserService
}

func NewAuthHandler(service UserService) *AuthHandler {
	return &AuthHandler{userService: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Bind the JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate credentials
	user, err := h.userService.ValidateCredentials(req.Email, req.Password)
	if err != nil {
		log.Println("ValidateCredentials failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate token
	token, err := h.userService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Return the token
	c.JSON(http.StatusOK, gin.H{"token": token})
}
