package repo

import (
	"gorm.io/gorm"
	"merch-shop/internal/model"
)

// DB - глобальная переменная для работы с БД
var DB *gorm.DB

// GetUserByUsername - ищет пользователя по имени
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser - создаёт нового пользователя
func CreateUser(user *model.User) error {
	return DB.Create(user).Error
}
