package postgres

import (
	"fmt"

	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	// Register the model with table name
	db.AutoMigrate(&models.Product{})
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *models.Product) error {
	// First verify user exists
	var user models.AppUser
	if err := r.db.First(&user, product.UserID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := r.db.Create(product).Error; err != nil {
		return err
	}

	// Load user data
	return r.db.Preload("User").First(product, product.ID).Error
}

func (r *ProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.Table("app_products").Preload("User").First(&product, id).Error
	return &product, err
}

func (r *ProductRepository) GetByUserID(userID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Table("app_products").Where("user_id = ?", userID).Find(&products).Error
	return products, err
}

func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Table("app_products").Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Table("app_products").Delete(&models.Product{}, id).Error
}

func (r *ProductRepository) UpdateProcessingStatus(id uint, status string) error {
	return r.db.Table("app_products").Model(&models.Product{}).Where("id = ?", id).
		Update("processing_status", status).Error
}

func (r *ProductRepository) UpdateCompressedImages(id uint, images pq.StringArray) error {
	return r.db.Table("app_products").Model(&models.Product{}).Where("id = ?", id).
		Update("compressed_product_images", images).Error
}
