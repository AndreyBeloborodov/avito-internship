package services

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"merch-shop/internal/errs"
	"merch-shop/internal/models"
	"merch-shop/internal/repositories"
	"time"
)

var jwtSecret = []byte("key-1848237283829139213")

// UserService - сервис для работы с пользователями
type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{userRepo: repo}
}

// Authenticate - метод для аутентификации и создания пользователя
func (s *UserService) Authenticate(req *models.AuthRequest) (*models.AuthResponse, error) {
	// Проверяем, есть ли пользователь в базе
	user, err := s.userRepo.GetUserByUsername(req.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Если пользователя нет в базе, создаём нового
		hashedPassword, _ := GetHashPassword(req.Password)
		user = &models.User{
			Username: req.Username,
			Password: hashedPassword,
			Coins:    1000, // Начальные монеты
		}
		if err = s.userRepo.CreateUser(user); err != nil {
			return nil, errs.ErrCreateUser
		}
	} else if err == nil && user != nil {
		// Если пользователь найден, проверяем пароль
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return nil, errs.ErrInvalidPassword
		}
	} else {
		return nil, errs.ErrInternalServer
	}

	// Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, errors.New("could not create JWT token")
	}

	return &models.AuthResponse{Token: tokenString}, nil
}

func GetHashPassword(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(hash), err
}

// ExtractUsernameFromToken разбирает токен, проверяет его валидность и возвращает username
func (s *UserService) ExtractUsernameFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("username not found in token")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", errors.New("invalid expiration time")
	}

	if time.Now().Unix() > int64(exp) {
		return "", errors.New("token expired")
	}

	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil || user == nil {
		return "", errs.ErrUserNotFound
	}

	return username, nil
}

// BuyMerch - обработка покупки предмета
func (s *UserService) BuyMerch(username string, merch *models.Merch) error {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrUserNotFound
		}
		return errs.ErrInternalServer
	}
	if user == nil {
		return errs.ErrInternalServer
	}

	// Проверяем, хватает ли монет
	if user.Coins < merch.Price {
		return errs.ErrNotEnoughCoins
	}
	// Покупаем предмет
	return s.userRepo.BuyMerch(user, merch)
}

// SendCoin - обработка отправки монет другому пользователю
func (s *UserService) SendCoin(username string, req models.SendCoinRequest) error {
	fromUser, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrUserNotFound
		}
		return errs.ErrInternalServer
	}
	if fromUser == nil {
		return errs.ErrInternalServer
	}

	toUser, err := s.userRepo.GetUserByUsername(req.ToUser)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrUserNotFound
		}
		return errs.ErrInternalServer
	}
	if toUser == nil {
		return errs.ErrInternalServer
	}

	// проверяем что количество монет положительное
	if req.Amount <= 0 {
		return errs.ErrNegativeCoins
	}
	// Проверяем, хватает ли монет у отправителя
	if fromUser.Coins < req.Amount {
		return errs.ErrNotEnoughCoins
	}
	if fromUser.Username == toUser.Username {
		return errs.ErrSendCoinsToYourself
	}
	// оправляем монеты
	return s.userRepo.SendCoin(fromUser, toUser, req.Amount)
}

// GetUserInfo - получает информацию о пользователе (баланс, инвентарь, историю транзакций)
func (s *UserService) GetUserInfo(username string) (*models.InfoResponse, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, errs.ErrInternalServer
	}
	if user == nil {
		return nil, errs.ErrInternalServer
	}

	// Получаем инвентарь пользователя
	inventory, err := s.userRepo.GetUserInventory(user.ID)
	if err != nil {
		return nil, err
	}

	// Получаем историю транзакций
	coinHistory, err := s.userRepo.GetCoinHistory(user.ID)
	if err != nil {
		return nil, err
	}

	info := &models.InfoResponse{
		Coins:       user.Coins,
		Inventory:   inventory,
		CoinHistory: coinHistory,
	}

	return info, nil
}
