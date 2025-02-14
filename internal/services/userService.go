package services

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"merch-shop/internal/models"
	"merch-shop/internal/repositories"
	"time"
)

var jwtSecret = []byte("key-1848237283829139213")

// UserService - сервис для работы с пользователями
type UserService struct {
	repo *repositories.UserRepo
}

func NewUserService(repo *repositories.UserRepo) *UserService {
	return &UserService{repo: repo}
}

// Authenticate - метод для аутентификации и создания пользователя
func (s *UserService) Authenticate(req *models.AuthRequest) (*models.AuthResponse, error) {
	// Проверяем, есть ли пользователь в базе
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		// Если пользователя нет в базе, создаём нового
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user = &models.User{
			Username: req.Username,
			Password: string(hashedPassword),
			Coins:    1000, // Начальные монеты
		}
		if err := s.repo.CreateUser(user); err != nil {
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

	return &models.AuthResponse{Token: tokenString}, nil
}
