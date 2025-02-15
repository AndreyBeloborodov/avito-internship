package repositories

import (
	"gorm.io/gorm"
	"merch-shop/internal/models"
)

type MerchRepository interface {
	GetMerchByName(name string) (*models.Merch, error)
}

// MerchRepo - структура для работы с базой данных
type MerchRepo struct {
	db *gorm.DB
}

func NewMerchRepo(db *gorm.DB) *MerchRepo {
	return &MerchRepo{db: db}
}

func (r *MerchRepo) GetMerchByName(name string) (*models.Merch, error) {
	var merch models.Merch
	if err := r.db.Where("name = ?", name).First(&merch).Error; err != nil {
		return nil, err
	}
	return &merch, nil
}
