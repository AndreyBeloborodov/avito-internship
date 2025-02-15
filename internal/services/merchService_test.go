package services

import (
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"merch-shop/internal/errs"
	"merch-shop/internal/mocks"
	"merch-shop/internal/models"
	"testing"
)

func TestGetMerchByName(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mockRepo *mocks.MerchRepository) string
		wantMerch *models.Merch
		wantErr   error
	}{
		{
			name: "успешное получение товара",
			mockSetup: func(mockRepo *mocks.MerchRepository) string {
				merch := &models.Merch{Name: "t-shirt", Price: 80}
				mockRepo.On("GetMerchByName", "t-shirt").Return(merch, nil)
				return "t-shirt"
			},
			wantMerch: &models.Merch{Name: "t-shirt", Price: 80},
			wantErr:   nil,
		},
		{
			name: "товар не найден",
			mockSetup: func(mockRepo *mocks.MerchRepository) string {
				mockRepo.On("GetMerchByName", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
				return "nonexistent"
			},
			wantMerch: nil,
			wantErr:   errs.ErrMerchNotFound,
		},
		{
			name: "ошибка в репозитории",
			mockSetup: func(mockRepo *mocks.MerchRepository) string {
				mockRepo.On("GetMerchByName", "t-shirt").Return(nil, errs.ErrInternalServer)
				return "t-shirt"
			},
			wantMerch: nil,
			wantErr:   errs.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMerchRepository(t)
			service := MerchService{merchRepo: mockRepo}

			name := tt.mockSetup(mockRepo)
			merch, err := service.GetMerchByName(name)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMerch, merch)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
