package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/KPVISHNUSAI/product-management-system/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestProductAPI(t *testing.T) {
	router := setupTestRouter()
	db := setupTestDB()
	defer cleanupTestDB(db)

	t.Run("Create Product", func(t *testing.T) {
		req := services.CreateProductRequest{
			UserID: 1,
			Name:   "Test Product",
			Price:  99.99,
		}
		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/products", bytes.NewBuffer(body))
		r.Header.Set("Authorization", "Bearer "+createTestToken(1))

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Product
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, req.Name, response.ProductName)
	})

	t.Run("Get Product", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/products/1", nil)
		r.Header.Set("Authorization", "Bearer "+createTestToken(1))

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func setupTestRouter() *gin.Engine {
	router := gin.New()
	// Setup test configuration
	return router
}

func setupTestDB() *gorm.DB {
	// Setup test database
	db, _ := database.NewPostgresDB("localhost", "5431", "postgres", "postgres", "product_management_test")
	db.AutoMigrate(&models.AppUser{}, &models.Product{})
	return db
}

func createTestToken(userID uint) string {
	// Create JWT token for testing
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}

func cleanupTestDB(db *gorm.DB) {
	db.Exec("DROP TABLE IF EXISTS products")
	db.Exec("DROP TABLE IF EXISTS app_users")
}
