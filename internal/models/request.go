package models

// AuthRequest - структура для запроса аутентификации
type AuthRequest struct {
	Username string `json:"username"` // Имя пользователя для аутентификации
	Password string `json:"password"` // Пароль для аутентификации
}

// SendCoinRequest - структура для запроса отправки монет другому пользователю
type SendCoinRequest struct {
	ToUser string `json:"toUser"` // Имя пользователя, которому нужно отправить монеты
	Amount int    `json:"amount"` // Количество монет, которые необходимо отправить
}
