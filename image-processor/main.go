package main

import (
	"fmt"

	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/KPVISHNUSAI/product-management-system/image-processor/config"
	"github.com/KPVISHNUSAI/product-management-system/image-processor/processor"
	"github.com/KPVISHNUSAI/product-management-system/image-processor/queue"
	"github.com/KPVISHNUSAI/product-management-system/pkg/cache"
	"github.com/KPVISHNUSAI/product-management-system/pkg/database"
	"github.com/KPVISHNUSAI/product-management-system/pkg/messaging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initialize AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2")},
	)
	s3Client := s3.New(sess)

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	if err != nil {
		panic(err)
	}

	// Initialize components
	imageProcessor := processor.NewImageProcessor(s3Client, cfg.AWS.Bucket)
	productRepo := postgres.NewProductRepository(db)

	// Add these lines
	redisClient, err := cache.NewRedisCache(
		fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		cfg.Redis.Password,
	)
	if err != nil {
		panic(err)
	}

	mqClient, err := messaging.NewRabbitMQClient(cfg.RabbitMQ.URL)
	if err != nil {
		panic(err)
	}

	productService := services.NewProductService(productRepo, mqClient, redisClient)

	// Initialize consumer
	consumer, err := queue.NewConsumer(
		cfg.RabbitMQ.URL,
		imageProcessor,
		productRepo,
		productService,
	)
	if err != nil {
		panic(err)
	}

	// Start consuming messages
	if err := consumer.Start(); err != nil {
		panic(err)
	}

	select {}
}
