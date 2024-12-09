// api/tests/benchmark/product_test.go
package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/KPVISHNUSAI/product-management-system/pkg/cache"
	"github.com/KPVISHNUSAI/product-management-system/pkg/database"
)

func setupBenchmarkEnvironment(b *testing.B) (*services.ProductService, *postgres.ProductRepository, *cache.RedisCache) {
	// Initialize database connection
	db, err := database.NewPostgresDB(
		"localhost",
		"5431",
		"postgres",
		"postgres",
		"product_management",
	)
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache("localhost:6379", "")
	if err != nil {
		b.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repository and service
	repo := postgres.NewProductRepository(db)
	service := services.NewProductService(repo, nil, redisCache)

	return service, repo, redisCache
}

func seedBenchmarkData(b *testing.B, repo *postgres.ProductRepository) {
	// Create test user
	user := &models.AppUser{
		Email: "benchmark@example.com",
		Name:  "Benchmark User",
	}
	db := repo.GetDB()
	db.Create(user)

	// Create test products
	for i := 0; i < 100; i++ {
		product := &models.Product{
			UserID:             user.ID,
			ProductName:        fmt.Sprintf("Product %d", i),
			ProductDescription: fmt.Sprintf("Description for product %d", i),
			ProductPrice:       float64(50 + i),
			ProcessingStatus:   "completed",
		}
		db.Create(product)
	}
}

func BenchmarkGetProduct(b *testing.B) {
	service, repo, _ := setupBenchmarkEnvironment(b)
	seedBenchmarkData(b, repo)

	b.Run("With Cache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service.GetProduct(1)
		}
	})

	b.Run("Without Cache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			repo.GetByID(1)
		}
	})
}

func BenchmarkCreateProduct(b *testing.B) {
	service, _, _ := setupBenchmarkEnvironment(b)

	req := &services.CreateProductRequest{
		UserID:      1,
		Name:        "Benchmark Product",
		Description: "Product for benchmark testing",
		Price:       99.99,
		Images:      []string{"test.jpg"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateProduct(req)
	}
}

func BenchmarkGetFilteredProducts(b *testing.B) {
	service, repo, _ := setupBenchmarkEnvironment(b)
	seedBenchmarkData(b, repo)

	b.Run("Small Result Set (10 products)", func(b *testing.B) {
		req := &services.FilterProductsRequest{
			UserID:      1,
			MinPrice:    50,
			MaxPrice:    60,
			ProductName: "Product",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service.GetFilteredProducts(req)
		}
	})

	b.Run("Large Result Set (50 products)", func(b *testing.B) {
		req := &services.FilterProductsRequest{
			UserID:      1,
			MinPrice:    50,
			MaxPrice:    100,
			ProductName: "Product",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service.GetFilteredProducts(req)
		}
	})
}

func BenchmarkCacheOperations(b *testing.B) {
	_, _, cache := setupBenchmarkEnvironment(b)
	ctx := context.Background()

	product := &models.Product{
		ID:          1,
		ProductName: "Cache Test Product",
	}

	b.Run("Cache Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Set(ctx, fmt.Sprintf("product:%d", i), product, time.Hour)
		}
	})

	b.Run("Cache Get", func(b *testing.B) {
		// Pre-populate cache
		cache.Set(ctx, "product:1", product, time.Hour)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var p models.Product
			cache.Get(ctx, "product:1", &p)
		}
	})
}
