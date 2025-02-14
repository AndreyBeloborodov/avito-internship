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

// UserExists - проверяет, существует ли пользователь с данным именем
func (r *UserRepo) UserExists(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// BuyMerch - списывает монеты и добавляет предмет в инвентарь
func (r *UserRepo) BuyMerch(user *models.User, merch *models.Merch) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *UserRepo) SendCoin(fromUser, toUser *models.User, amount int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Списываем монеты у отправителя
		fromUser.Coins -= amount
		if err := tx.Save(fromUser).Error; err != nil {
			return err
		}

		// Начисляем монеты получателю
		toUser.Coins += amount
		if err := tx.Save(toUser).Error; err != nil {
			return err
		}

		// Записываем транзакцию в историю
		transaction := models.Transaction{
			SenderId:   fromUser.ID,
			ReceiverId: toUser.ID,
			Amount:     amount,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		return nil
	})
}
