package models

// ErrorResponse - структура для ответа с ошибкой.
type ErrorResponse struct {
	Errors string `json:"errors"` // Сообщение об ошибке, описывающее проблему.
}

// AuthResponse - структура для ответа с токеном
type AuthResponse struct {
	Token string `json:"token"` // JWT-токен для доступа к защищенным ресурсам.
}

// InfoResponse - структура для ответа с информацией о монетах, инвентаре и истории транзакций.
type InfoResponse struct {
	Coins       int         `json:"coins"`       // Количество доступных монет
	Inventory   []Item      `json:"inventory"`   // Инвентарь пользователя
	CoinHistory CoinHistory `json:"coinHistory"` // История транзакций с монетами
}

// Item - структура для предмета в инвентаре.
type Item struct {
	Type     string `json:"type"`     // Тип предмета
	Quantity int    `json:"quantity"` // Количество предметов
}

// CoinHistory - структура для истории монетных операций.
type CoinHistory struct {
	Received []CoinTransaction `json:"received"` // Полученные монеты
	Sent     []CoinTransaction `json:"sent"`     // Отправленные монеты
}

// CoinTransaction - структура для транзакции с монетами.
type CoinTransaction struct {
	FromUser string `json:"fromUser,omitempty"` // Отправитель (если монеты получены)
	ToUser   string `json:"toUser,omitempty"`   // Получатель (если монеты отправлены)
	Amount   int    `json:"amount"`             // Количество монет
}
