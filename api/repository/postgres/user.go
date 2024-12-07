package postgres

import (
	"github.com/KPVISHNUSAI/product-management-system/api/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.AppUser) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id uint) (*models.AppUser, error) {
	var user models.AppUser
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Update(user *models.AppUser) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.AppUser{}, id).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.AppUser, error) {
	var user models.AppUser
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}
