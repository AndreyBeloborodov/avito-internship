package repositories

import (
	"errors"
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

// BuyMerch - списывает монеты и добавляет предмет в инвентарь
func (r *UserRepo) BuyMerch(user *models.User, merch *models.Merch) error {
	// Транзакция на случай ошибки
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Проверяем, хватает ли монет
		if user.Coins < merch.Price {
			return errors.New("not enough coins")
		}

		// Списываем монеты
		user.Coins -= merch.Price
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		// Добавляем предмет в инвентарь (запись в purchases)
		purchase := models.Purchase{
			UserID:  user.ID,
			MerchID: merch.ID,
		}

		if err := tx.Create(&purchase).Error; err != nil {
			return err
		}

		return nil
	})
}
