package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/pkg/messaging"
	"github.com/go-redis/redis"
	"github.com/lib/pq"
)

type ImageProcessingTask struct {
	ProductID uint     `json:"product_id"`
	Images    []string `json:"images"`
}

type ProductRepository interface {
	Create(product *models.Product) error
	GetByID(id uint) (*models.Product, error)
	GetByUserID(userID uint) ([]models.Product, error)
	Update(product *models.Product) error
	UpdateProcessingStatus(id uint, status string) error
	UpdateCompressedImages(id uint, images pq.StringArray) error
}

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

type ProductService struct {
	productRepo ProductRepository
	mqPublisher messaging.Publisher
	cache       Cache
}

const (
	defaultCacheDuration = 1 * time.Hour
	shortCacheDuration   = 5 * time.Minute
	longCacheDuration    = 24 * time.Hour

	productCachePrefix = "product:"
	userCachePrefix    = "user:"
	listCachePrefix    = "list:"
)

func (s *ProductService) getCacheDuration(dataType string) time.Duration {
	switch dataType {
	case "product":
		return defaultCacheDuration
	case "list":
		return shortCacheDuration
	case "static":
		return longCacheDuration
	default:
		return defaultCacheDuration
	}
}

func (s *ProductService) getCacheKey(prefix string, id interface{}) string {
	return fmt.Sprintf("%s%v", prefix, id)
}

func NewProductService(repo ProductRepository, publisher messaging.Publisher, cache Cache) *ProductService {
	return &ProductService{
		productRepo: repo,
		mqPublisher: publisher,
		cache:       cache,
	}
}

type CreateProductRequest struct {
	UserID      uint     `json:"user_id"`
	Name        string   `json:"product_name"`
	Description string   `json:"product_description"`
	Price       float64  `json:"product_price"`
	Images      []string `json:"product_images"`
}

func (s *ProductService) CreateProduct(req *CreateProductRequest) (*models.Product, error) {
	product := &models.Product{
		UserID:             req.UserID,
		ProductName:        req.Name,
		ProductDescription: req.Description,
		ProductPrice:       req.Price,
		ProductImages:      pq.StringArray(req.Images), // Ensure correct array type
		ProcessingStatus:   "pending",
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	// Queue image processing task
	task := ImageProcessingTask{
		ProductID: product.ID,
		Images:    req.Images,
	}

	if err := s.queueImageProcessing(task); err != nil {
		return product, err
	}

	return product, nil
}

func (s *ProductService) handleCacheError(err error, operation string) {
	if err != redis.Nil {
		log.Printf("Cache %s error: %v", operation, err)
	}
}

func (s *ProductService) GetProduct(id uint) (*models.Product, error) {
	ctx := context.Background()
	cacheKey := s.getCacheKey(productCachePrefix, id)

	var product *models.Product
	err := s.cache.Get(ctx, cacheKey, &product)
	if err == nil {
		return product, nil
	}
	s.handleCacheError(err, "get")

	// Use existing err variable
	product, err = s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, cacheKey, product, s.getCacheDuration("product")); err != nil {
		s.handleCacheError(err, "set")
	}

	return product, nil
}

func (s *ProductService) GetUserProducts(userID uint) ([]models.Product, error) {
	return s.productRepo.GetByUserID(userID)
}

func (s *ProductService) queueImageProcessing(task ImageProcessingTask) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return s.mqPublisher.Publish("image_processing", taskBytes)
}

func (s *ProductService) InvalidateCache(id uint) error {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("product:%d", id)
	return s.cache.Delete(ctx, cacheKey)
}

func (s *ProductService) UpdateProduct(product *models.Product) error {
	err := s.productRepo.Update(product)
	if err != nil {
		return err
	}
	return s.InvalidateCache(product.ID)
}
