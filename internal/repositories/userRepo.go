package repositories

import (
	"gorm.io/gorm"
	"merch-shop/internal/models"
)

// UserRepo - структура для работы с базой данных
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// GetUserByUsername - ищет пользователя по имени
func (r *UserRepo) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser - создаёт нового пользователя
func (r *UserRepo) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}
