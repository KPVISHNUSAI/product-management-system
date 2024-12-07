package postgres

import (
	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("User").First(&product, id).Error
	return &product, err
}

func (r *ProductRepository) GetByUserID(userID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("user_id = ?", userID).Find(&products).Error
	return products, err
}

func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

func (r *ProductRepository) UpdateProcessingStatus(id uint, status string) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).
		Update("processing_status", status).Error
}

func (r *ProductRepository) UpdateCompressedImages(id uint, images []string) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).
		Update("compressed_product_images", images).Error
}
