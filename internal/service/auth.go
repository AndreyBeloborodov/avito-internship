package service

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"merch-shop/internal/model"
	"merch-shop/internal/repo"
	"time"
)

var jwtSecret = []byte("key-1848237283829139213")

// AuthService - сервис для аутентификации
type AuthService struct{}

// AuthRequest - структура для запроса аутентификации
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse - структура для ответа с токеном
type AuthResponse struct {
	Token string `json:"token"`
}

// Authenticate - метод для аутентификации и создания пользователя
func (s *AuthService) Authenticate(req *AuthRequest) (*AuthResponse, error) {
	// Проверяем, есть ли пользователь в базе
	user, err := repo.GetUserByUsername(req.Username)
	if err != nil {
		// Если пользователя нет в базе, создаём нового
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user = &model.User{
			Username: req.Username,
			Password: string(hashedPassword),
			Coins:    1000, // Начальные монеты
		}
		if err := repo.CreateUser(user); err != nil {
			return nil, errors.New("could not create user")
		}
	} else {
		// Если пользователь найден, проверяем пароль
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return nil, errors.New("invalid password")
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

	return &AuthResponse{Token: tokenString}, nil
}
