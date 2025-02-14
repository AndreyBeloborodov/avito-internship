package services

import (
	"merch-shop/internal/repositories"
)

// MerchService - сервис для работы с пользователями
type MerchService struct {
	repo *repositories.MerchRepo
}

func NewMerchService(repo *repositories.MerchRepo) *MerchService {
	return &MerchService{repo: repo}
}
