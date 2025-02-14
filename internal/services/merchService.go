package services

import (
	"merch-shop/internal/models"
	"merch-shop/internal/repositories"
)

// MerchService - сервис для работы с пользователями
type MerchService struct {
	merchRepo *repositories.MerchRepo
}

func NewMerchService(repo *repositories.MerchRepo) *MerchService {
	return &MerchService{merchRepo: repo}
}

func (s *MerchService) GetMerchByName(name string) (*models.Merch, error) {
	return s.merchRepo.GetMerchByName(name)
}
