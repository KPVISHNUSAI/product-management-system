package services

import (
	"encoding/json"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/pkg/messaging"
)

type ImageProcessingTask struct {
	ProductID uint     `json:"product_id"`
	Images    []string `json:"images"`
}

type ProductService struct {
	productRepo *postgres.ProductRepository
	mqPublisher messaging.Publisher
}

func NewProductService(repo *postgres.ProductRepository, publisher messaging.Publisher) *ProductService {
	return &ProductService{
		productRepo: repo,
		mqPublisher: publisher,
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
		ProductImages:      req.Images,
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

func (s *ProductService) GetProduct(id uint) (*models.Product, error) {
	return s.productRepo.GetByID(id)
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
