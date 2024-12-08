package models

import (
	"time"

	"github.com/lib/pq"
)

type Product struct {
	ID                      uint           `gorm:"primaryKey"`
	UserID                  uint           `gorm:"not null"`
	ProductName             string         `gorm:"not null"`
	ProductDescription      string         `gorm:"column:product_description"`
	ProductPrice            float64        `gorm:"not null"`
	ProductImages           pq.StringArray `gorm:"type:text[]"`
	CompressedProductImages pq.StringArray `gorm:"type:text[]"`
	ProcessingStatus        string         `gorm:"default:pending"`
	CreatedAt               time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt               time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	User                    AppUser        `gorm:"foreignKey:UserID"`
}

func (Product) TableName() string {
	return "app_products"
}
