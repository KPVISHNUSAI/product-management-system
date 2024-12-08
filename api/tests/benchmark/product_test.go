// api/tests/benchmark/product_test.go
package benchmark

import (
	"fmt"
	"testing"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/KPVISHNUSAI/product-management-system/pkg/cache"
	"github.com/KPVISHNUSAI/product-management-system/pkg/database"
	"gorm.io/gorm"
)

type ProductService interface {
	GetProduct(id uint) (*models.Product, error)
	GetUserProducts(userID uint) ([]models.Product, error)
	CreateProduct(req *services.CreateProductRequest) (*models.Product, error)
}

func setupBenchmarkDB() *gorm.DB {
	db, err := database.NewPostgresDB(
		"localhost",
		"5431",
		"postgres",
		"postgres",
		"product_management_test",
	)
	if err != nil {
		panic(err)
	}

	// Seed test data
	seedTestData(db)
	return db
}

func setupBenchmarkCache() *cache.RedisCache {
	redisCache, err := cache.NewRedisCache("localhost:6379", "redis")
	if err != nil {
		panic(err)
	}
	return redisCache
}

func seedTestData(db *gorm.DB) {
	// Create test user
	user := &models.AppUser{
		Email: "test@example.com",
		Name:  "Test User",
	}
	db.Create(user)

	// Create 100 test products
	for i := 0; i < 100; i++ {
		product := &models.Product{
			UserID:       user.ID,
			ProductName:  fmt.Sprintf("Product %d", i),
			ProductPrice: 99.99,
		}
		db.Create(product)
	}
}

func BenchmarkGetProduct(b *testing.B) {
	db := setupBenchmarkDB()
	repo := postgres.NewProductRepository(db)
	service := setupBenchmarkService()

	b.Run("With Cache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = service.GetProduct(1)
		}
	})

	b.Run("Without Cache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = repo.GetByID(1)
		}
	})
}

func BenchmarkCreateProduct(b *testing.B) {
	service := setupBenchmarkService()
	req := &services.CreateProductRequest{
		UserID: 1,
		Name:   "Benchmark Product",
		Price:  99.99,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateProduct(req)
	}
}

func BenchmarkUserProductsList(b *testing.B) {
	service := setupBenchmarkService()

	b.Run("Small List (10 items)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.GetUserProducts(1) // User with 10 products
		}
	})

	b.Run("Large List (100 items)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.GetUserProducts(2) // User with 100 products
		}
	})
}

func setupBenchmarkService() ProductService {
	db := setupBenchmarkDB()
	repo := postgres.NewProductRepository(db)
	cache := setupBenchmarkCache()
	return services.NewProductService(repo, nil, cache)
}
