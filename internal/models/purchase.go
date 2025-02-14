package models

import "gorm.io/gorm"

// Purchase - структура для хранения информации о покупке
type Purchase struct {
	gorm.Model
	UserID  uint `gorm:"not null" json:"userId"`
	MerchID uint `gorm:"not null" json:"merchId"`
}
