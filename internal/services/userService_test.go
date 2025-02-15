package services

import (
	"errors"
	"merch-shop/internal/errs"
	"merch-shop/internal/mocks"
	"merch-shop/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendCoin(t *testing.T) {

	fromUser := &models.User{Username: "Andrey", Coins: 100}
	toUser := &models.User{Username: "Ivan", Coins: 50}

	t.Run("успешная отправка", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)
		mockRepo.On("GetUserByUsername", toUser.Username).Return(toUser, nil)
		mockRepo.On("SendCoin", fromUser, toUser, 50).Return(nil)

		err := service.SendCoin(fromUser.Username, models.SendCoinRequest{ToUser: toUser.Username, Amount: 50})
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", "Unknown").Return(nil, errs.ErrUserNotFound)

		err := service.SendCoin("Unknown", models.SendCoinRequest{ToUser: toUser.Username, Amount: 50})
		assert.ErrorIs(t, err, errs.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("отправка самому себе", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)

		err := service.SendCoin(fromUser.Username, models.SendCoinRequest{ToUser: fromUser.Username, Amount: 1})
		assert.ErrorIs(t, err, errs.ErrSendCoinsToYourself)
	})

	t.Run("недостаточно монет", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)
		mockRepo.On("GetUserByUsername", toUser.Username).Return(toUser, nil)

		err := service.SendCoin(fromUser.Username, models.SendCoinRequest{ToUser: toUser.Username, Amount: 2000})
		assert.ErrorIs(t, err, errs.ErrNotEnoughCoins)
	})

	t.Run("отрицательное количество монет", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)
		mockRepo.On("GetUserByUsername", toUser.Username).Return(toUser, nil)

		err := service.SendCoin(fromUser.Username, models.SendCoinRequest{ToUser: toUser.Username, Amount: -10})
		assert.ErrorIs(t, err, errs.ErrNegativeCoins)
	})

	t.Run("ошибка при транзакции", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", fromUser.Username).Return(fromUser, nil)
		mockRepo.On("GetUserByUsername", toUser.Username).Return(toUser, nil)
		mockRepo.On("SendCoin", fromUser, toUser, 50).Return(errors.New("db error"))

		err := service.SendCoin(fromUser.Username, models.SendCoinRequest{ToUser: toUser.Username, Amount: 50})
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestBuyMerch(t *testing.T) {
	user := &models.User{Username: "Andrey", Coins: 100}
	merch := &models.Merch{Name: "t-shirt", Price: 80}

	t.Run("успешная покупка", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)
		mockRepo.On("BuyMerch", user, merch).Return(nil)

		err := service.BuyMerch(user.Username, merch)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", "Unknown").Return(nil, errs.ErrUserNotFound)

		err := service.BuyMerch("Unknown", merch)
		assert.ErrorIs(t, err, errs.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("недостаточно монет", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		poorUser := &models.User{Username: "BrokeGuy", Coins: 10}

		mockRepo.On("GetUserByUsername", poorUser.Username).Return(poorUser, nil)

		err := service.BuyMerch(poorUser.Username, merch)
		assert.ErrorIs(t, err, errs.ErrNotEnoughCoins)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при покупке", func(t *testing.T) {
		mockRepo := mocks.NewUserRepository(t)
		service := UserService{userRepo: mockRepo}

		mockRepo.On("GetUserByUsername", user.Username).Return(user, nil)
		mockRepo.On("BuyMerch", user, merch).Return(errors.New("db error"))

		err := service.BuyMerch(user.Username, merch)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
