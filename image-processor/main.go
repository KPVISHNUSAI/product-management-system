package main

import (
	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/image-processor/config"
	"github.com/KPVISHNUSAI/product-management-system/image-processor/processor"
	"github.com/KPVISHNUSAI/product-management-system/image-processor/queue"
	"github.com/KPVISHNUSAI/product-management-system/pkg/database"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initialize AWS session
	sess := session.Must(session.NewSession())
	s3Client := s3.New(sess)

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	if err != nil {
		panic(err)
	}

	// Initialize components
	imageProcessor := processor.NewImageProcessor(s3Client, cfg.AWS.Bucket)
	productRepo := postgres.NewProductRepository(db)

	// Initialize consumer
	consumer, err := queue.NewConsumer(
		cfg.RabbitMQ.URL,
		imageProcessor,
		productRepo,
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
