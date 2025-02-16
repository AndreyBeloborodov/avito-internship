package services

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"merch-shop/internal/errs"
	"merch-shop/internal/mocks"
	"merch-shop/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendCoin(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mockRepo *mocks.UserRepository) (string, models.SendCoinRequest)
		wantErr   error
	}{
		{
			name: "успешная отправка",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, models.SendCoinRequest) {
				fromUser := &models.User{Username: "Andrey", Coins: 100}
				toUser := &models.User{Username: "Ivan", Coins: 50}

				mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)
				mockRepo.On("GetUserByUsername", toUser.Username).Return(toUser, nil)
				mockRepo.On("SendCoin", fromUser, toUser, 50).Return(nil)

				return fromUser.Username, models.SendCoinRequest{ToUser: toUser.Username, Amount: 50}
			},
			wantErr: nil,
		},
		{
			name: "пользователь не найден",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, models.SendCoinRequest) {
				mockRepo.On("GetUserByUsername", "Unknown").Return(nil, gorm.ErrRecordNotFound)

				return "Unknown", models.SendCoinRequest{ToUser: "Ivan", Amount: 50}
			},
			wantErr: errs.ErrUserNotFound,
		},
		{
			name: "отправка самому себе",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, models.SendCoinRequest) {
				user := &models.User{Username: "Andrey", Coins: 100}

				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)

				return user.Username, models.SendCoinRequest{ToUser: user.Username, Amount: 1}
			},
			wantErr: errs.ErrSendCoinsToYourself,
		},
		{
			name: "недостаточно монет",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, models.SendCoinRequest) {
				fromUser := &models.User{Username: "Andrey", Coins: 100}
				toUser := &models.User{Username: "Ivan", Coins: 50}

				mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)
				mockRepo.On("GetUserByUsername", toUser.Username).Return(toUser, nil)

				return fromUser.Username, models.SendCoinRequest{ToUser: toUser.Username, Amount: 2000}
			},
			wantErr: errs.ErrNotEnoughCoins,
		},
		{
			name: "отрицательное количество монет",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, models.SendCoinRequest) {
				fromUser := &models.User{Username: "Andrey", Coins: 100}
				toUser := &models.User{Username: "Ivan", Coins: 50}

				mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)
				mockRepo.On("GetUserByUsername", toUser.Username).Return(toUser, nil)

				return fromUser.Username, models.SendCoinRequest{ToUser: toUser.Username, Amount: -10}
			},
			wantErr: errs.ErrNegativeCoins,
		},
		{
			name: "ошибка при транзакции",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, models.SendCoinRequest) {
				fromUser := &models.User{Username: "Andrey", Coins: 100}
				toUser := &models.User{Username: "Ivan", Coins: 50}

				mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)
				mockRepo.On("GetUserByUsername", toUser.Username).Return(toUser, nil)
				mockRepo.On("SendCoin", fromUser, toUser, 50).Return(errs.ErrInternalServer)

				return fromUser.Username, models.SendCoinRequest{ToUser: toUser.Username, Amount: 50}
			},
			wantErr: errs.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewUserRepository(t)
			service := UserService{userRepo: mockRepo}

			username, request := tt.mockSetup(mockRepo)

			err := service.SendCoin(username, request)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBuyMerch(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mockRepo *mocks.UserRepository) (string, *models.Merch)
		wantErr   error
	}{
		{
			name: "успешная покупка",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, *models.Merch) {
				user := &models.User{Username: "Andrey", Coins: 100}
				merch := &models.Merch{Name: "t-shirt", Price: 80}

				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)
				mockRepo.On("BuyMerch", user, merch).Return(nil)

				return user.Username, merch
			},
			wantErr: nil,
		},
		{
			name: "пользователь не найден",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, *models.Merch) {
				merch := &models.Merch{Name: "t-shirt", Price: 80}

				mockRepo.On("GetUserByUsername", "Unknown").Return(nil, gorm.ErrRecordNotFound)

				return "Unknown", merch
			},
			wantErr: errs.ErrUserNotFound,
		},
		{
			name: "недостаточно монет",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, *models.Merch) {
				user := &models.User{Username: "BrokeGuy", Coins: 10}
				merch := &models.Merch{Name: "t-shirt", Price: 80}

				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)

				return user.Username, merch
			},
			wantErr: errs.ErrNotEnoughCoins,
		},
		{
			name: "ошибка при покупке",
			mockSetup: func(mockRepo *mocks.UserRepository) (string, *models.Merch) {
				user := &models.User{Username: "Andrey", Coins: 100}
				merch := &models.Merch{Name: "t-shirt", Price: 80}

				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)
				mockRepo.On("BuyMerch", user, merch).Return(errs.ErrInternalServer)

				return user.Username, merch
			},
			wantErr: errs.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewUserRepository(t)
			service := UserService{userRepo: mockRepo}

			username, merch := tt.mockSetup(mockRepo)

			err := service.BuyMerch(username, merch)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserInfo(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		setupMocks    func(mockRepo *mocks.UserRepository)
		expectedError error
		expectedInfo  *models.InfoResponse
	}{
		{
			name:     "успешное получение информации",
			username: "Andrey",
			setupMocks: func(mockRepo *mocks.UserRepository) {
				user := &models.User{Username: "Andrey", Coins: 100}
				inventory := []models.Item{{Type: "t-shirt", Quantity: 1}}
				coinHistory := models.CoinHistory{
					Received: []models.CoinTransaction{{FromUser: "Ivan", Amount: 50}},
					Sent:     []models.CoinTransaction{{ToUser: "Alex", Amount: 20}},
				}

				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)
				mockRepo.On("GetUserInventory", user.ID).Return(inventory, nil)
				mockRepo.On("GetCoinHistory", user.ID).Return(coinHistory, nil)
			},
			expectedError: nil,
			expectedInfo: &models.InfoResponse{
				Coins:     100,
				Inventory: []models.Item{{Type: "t-shirt", Quantity: 1}},
				CoinHistory: models.CoinHistory{
					Received: []models.CoinTransaction{{FromUser: "Ivan", Amount: 50}},
					Sent:     []models.CoinTransaction{{ToUser: "Alex", Amount: 20}},
				},
			},
		},
		{
			name:     "пользователь не найден",
			username: "Unknown",
			setupMocks: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("GetUserByUsername", "Unknown").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: errs.ErrUserNotFound,
			expectedInfo:  nil,
		},
		{
			name:     "ошибка при получении инвентаря",
			username: "Andrey",
			setupMocks: func(mockRepo *mocks.UserRepository) {
				user := &models.User{Username: "Andrey", Coins: 100}
				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)
				mockRepo.On("GetUserInventory", user.ID).Return(nil, errs.ErrInternalServer)
			},
			expectedError: errs.ErrInternalServer,
			expectedInfo:  nil,
		},
		{
			name:     "ошибка при получении истории транзакций",
			username: "Andrey",
			setupMocks: func(mockRepo *mocks.UserRepository) {
				user := &models.User{Username: "Andrey", Coins: 100}
				inventory := []models.Item{{Type: "t-shirt", Quantity: 1}}

				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)
				mockRepo.On("GetUserInventory", user.ID).Return(inventory, nil)
				mockRepo.On("GetCoinHistory", user.ID).Return(models.CoinHistory{}, errs.ErrInternalServer)
			},
			expectedError: errs.ErrInternalServer,
			expectedInfo:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewUserRepository(t)
			service := UserService{userRepo: mockRepo}

			tt.setupMocks(mockRepo)

			info, err := service.GetUserInfo(tt.username)
			assert.ErrorIs(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedInfo, info)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mockRepo *mocks.UserRepository) (*models.AuthRequest, error)
		wantResp  *models.AuthResponse
		wantErr   error
	}{
		{
			name: "успешная аутентификация (пользователь существует)",
			mockSetup: func(mockRepo *mocks.UserRepository) (*models.AuthRequest, error) {
				password := "password123"
				hashPassword, _ := GetHashPassword(password)
				user := &models.User{
					Username: "Andrey",
					Password: hashPassword,
				}
				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)

				return &models.AuthRequest{Username: "Andrey", Password: password}, nil
			},
			wantResp: &models.AuthResponse{Token: "some-jwt-token"}, // ожидаемый JWT
			wantErr:  nil,
		},
		{
			name: "пользователь не найден, создаем нового",
			mockSetup: func(mockRepo *mocks.UserRepository) (*models.AuthRequest, error) {
				mockRepo.On("GetUserByUsername", "NewUser").Return(nil, gorm.ErrRecordNotFound)

				// Мокируем создание нового пользователя
				mockRepo.On("CreateUser", mock.Anything).Return(nil)

				return &models.AuthRequest{Username: "NewUser", Password: "newPassword123"}, nil
			},
			wantResp: &models.AuthResponse{Token: "some-jwt-token"}, // ожидаемый JWT
			wantErr:  nil,
		},
		{
			name: "неправильный пароль",
			mockSetup: func(mockRepo *mocks.UserRepository) (*models.AuthRequest, error) {
				user := &models.User{Username: "Andrey", Password: "$2a$10$Qh71lZRj8ix5brUBUoKlfe1sq5.nTkVffV6fwSTv.Hk1vwZZwP6Pi"}
				mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)

				// Здесь пароль неправильный
				return &models.AuthRequest{Username: "Andrey", Password: "wrongPassword"}, nil
			},
			wantResp: nil,
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name: "ошибка при создании пользователя",
			mockSetup: func(mockRepo *mocks.UserRepository) (*models.AuthRequest, error) {
				mockRepo.On("GetUserByUsername", "NewUser").Return(nil, gorm.ErrRecordNotFound)

				// Мокируем ошибку при создании пользователя
				mockRepo.On("CreateUser", mock.Anything).Return(errs.ErrCreateUser)

				return &models.AuthRequest{Username: "NewUser", Password: "newPassword123"}, nil
			},
			wantResp: nil,
			wantErr:  errs.ErrCreateUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewUserRepository(t)
			service := UserService{userRepo: mockRepo}

			req, err := tt.mockSetup(mockRepo)
			if err != nil {
				t.Fatal("mock setup error:", err)
			}

			resp, err := service.Authenticate(req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Token)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
