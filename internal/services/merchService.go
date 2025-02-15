package services

import (
	"errors"
	"gorm.io/gorm"
	"merch-shop/internal/errs"
	"merch-shop/internal/models"
	"merch-shop/internal/repositories"
)

// MerchService - сервис для работы с пользователями
type MerchService struct {
	merchRepo repositories.MerchRepository
}

func NewMerchService(repo repositories.MerchRepository) *MerchService {
	return &MerchService{merchRepo: repo}
}

func (s *MerchService) GetMerchByName(name string) (*models.Merch, error) {
	merch, err := s.merchRepo.GetMerchByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrMerchNotFound
		}
		return nil, err
	}
	return merch, nil
}
