package services

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"merch-shop/internal/errs"
	"merch-shop/internal/models"
	"merch-shop/internal/repositories"
	"net/http"
	"strings"
	"time"
)

var jwtSecret = []byte("key-1848237283829139213")

// UserService - сервис для работы с пользователями
type UserService struct {
	userRepo *repositories.UserRepo
}

func NewUserService(repo *repositories.UserRepo) *UserService {
	return &UserService{userRepo: repo}
}

// Authenticate - метод для аутентификации и создания пользователя
func (s *UserService) Authenticate(req *models.AuthRequest) (*models.AuthResponse, error) {
	// Проверяем, есть ли пользователь в базе
	user, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		// Если пользователя нет в базе, создаём нового
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user = &models.User{
			Username: req.Username,
			Password: string(hashedPassword),
			Coins:    1000, // Начальные монеты
		}
		if err := s.userRepo.CreateUser(user); err != nil {
			return nil, errors.New("could not create user")
		}
	} else {
		// Если пользователь найден, проверяем пароль
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return nil, errs.ErrInvalidPassword
		}
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

// ExtractUsernameFromToken разбирает токен, проверяет его валидность и возвращает username
func (s *UserService) ExtractUsernameFromToken(r *http.Request) (string, error) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	// Ожидаем формат "Bearer <token>"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Разбираем токен и проверяем подпись
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	// Извлекаем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	// Получаем username
	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("username not found in token")
	}

	// Проверяем срок действия токена (exp в формате Unix timestamp)
	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", errors.New("invalid expiration time")
	}

	if time.Now().Unix() > int64(exp) {
		return "", errors.New("token expired")
	}

	// Проверяем, существует ли пользователь в БД
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil || user == nil {
		return "", errs.ErrUserNotFound
	}

	// Токен валиден, пользователь существует
	return username, nil
}

// BuyMerch - обработка покупки предмета
func (s *UserService) BuyMerch(username string, merch *models.Merch) error {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil || user == nil {
		return errs.ErrUserNotFound
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
	if err != nil || fromUser == nil {
		return errs.ErrUserNotFound
	}
	toUser, err := s.userRepo.GetUserByUsername(req.ToUser)
	if err != nil || toUser == nil {
		return errs.ErrUserNotFound
	}
	// проверяем что количество монет положительное
	if req.Amount <= 0 {
		return errs.ErrNegativeCoins
	}
	// Проверяем, хватает ли монет у отправителя
	if fromUser.Coins < req.Amount {
		return errs.ErrNotEnoughCoins
	}
	// оправляем монеты
	return s.userRepo.SendCoin(fromUser, toUser, req.Amount)
}
