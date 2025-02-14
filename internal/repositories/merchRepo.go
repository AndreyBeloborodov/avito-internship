package repositories

import (
	"gorm.io/gorm"
)

// MerchRepo - структура для работы с базой данных
type MerchRepo struct {
	db *gorm.DB
}

func NewMerchRepo(db *gorm.DB) *MerchRepo {
	return &MerchRepo{db: db}
}
