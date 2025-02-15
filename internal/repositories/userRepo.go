package repositories

import (
	"gorm.io/gorm"
	"merch-shop/internal/models"
)

type UserRepository interface {
	GetUserByUsername(username string) (*models.User, error)
	CreateUser(user *models.User) error
	SendCoin(fromUser, toUser *models.User, amount int) error
	BuyMerch(user *models.User, merch *models.Merch) error
	GetUserInventory(userID uint) ([]models.Item, error)
	GetCoinHistory(userID uint) (models.CoinHistory, error)
}

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

// GetUserInventory - получает список предметов в инвентаре пользователя
func (r *UserRepo) GetUserInventory(userID uint) ([]models.Item, error) {
	var items []models.Item
	err := r.db.Raw(`
		SELECT m.name AS type, COUNT(p.merch_id) AS quantity
		FROM purchases p
		JOIN merches m ON p.merch_id = m.id
		WHERE p.user_id = ?
		GROUP BY m.name
	`, userID).Scan(&items).Error

	return items, err
}

// GetCoinHistory - получает историю отправленных и полученных монет
func (r *UserRepo) GetCoinHistory(userID uint) (models.CoinHistory, error) {
	var history models.CoinHistory

	// Получаем полученные монеты
	err := r.db.Raw(`
		SELECT u.username AS from_user, t.amount
		FROM transactions t
		JOIN users u ON t.sender_id = u.id
		WHERE t.receiver_id = ?
	`, userID).Scan(&history.Received).Error
	if err != nil {
		return history, err
	}

	// Получаем отправленные монеты
	err = r.db.Raw(`
		SELECT u.username AS to_user, t.amount
		FROM transactions t
		JOIN users u ON t.receiver_id = u.id
		WHERE t.sender_id = ?
	`, userID).Scan(&history.Sent).Error

	return history, err
}
