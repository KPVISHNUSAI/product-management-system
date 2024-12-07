package models

import (
	"time"
)

type Product struct {
	ID                      uint   `gorm:"primaryKey"`
	UserID                  uint   `gorm:"not null"`
	ProductName             string `gorm:"not null"`
	ProductDescription      string
	ProductPrice            float64   `gorm:"not null"`
	ProductImages           []string  `gorm:"type:text[]"`
	CompressedProductImages []string  `gorm:"type:text[]"`
	ProcessingStatus        string    `gorm:"default:pending"`
	CreatedAt               time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt               time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	User                    AppUser   `gorm:"foreignKey:UserID"`
}
