package main

import (
	"fmt"

	"github.com/KPVISHNUSAI/product-management-system/api/config"
	"github.com/KPVISHNUSAI/product-management-system/api/handlers"
	"github.com/KPVISHNUSAI/product-management-system/api/middleware"
	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/KPVISHNUSAI/product-management-system/pkg/cache"
	"github.com/KPVISHNUSAI/product-management-system/pkg/database"
	"github.com/KPVISHNUSAI/product-management-system/pkg/messaging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize RabbitMQ
	mqClient, err := messaging.NewRabbitMQClient(cfg.RabbitMQ.URL)
	if err != nil {
		logger.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}
	defer mqClient.Close()

	// Initialize database
	db, err := database.NewPostgresDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	// Initialize Redis
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	logger.Info("Connecting to Redis", zap.String("addr", redisAddr)) // Add logging
	redisClient, err := cache.NewRedisCache(redisAddr, cfg.Redis.Password)
	if err != nil {
		logger.Fatal("failed to connect to Redis",
			zap.Error(err),
			zap.String("addr", redisAddr),
			zap.String("host", cfg.Redis.Host),
			zap.String("port", cfg.Redis.Port))
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	productRepo := postgres.NewProductRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo, cfg.Server.JWTSecret)
	// Initialize services with MQ
	productService := services.NewProductService(productRepo, mqClient, redisClient)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	productHandler := handlers.NewProductHandler(productService)

	// Initialize router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware(logger))

	// Routes
	api := r.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		products := api.Group("/products")
		products.Use(middleware.AuthMiddleware(cfg.Server.JWTSecret))
		{
			products.POST("/", productHandler.CreateProduct)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("/", productHandler.GetUserProducts)
		}
	}

	// Start server
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
