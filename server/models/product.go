package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Id          int `gorm:"uniqueIndex"`
	Name        string
	Price       int
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
