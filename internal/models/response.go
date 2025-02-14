package models

// DefaultResponse представляет тандартный ответ сервера
type DefaultResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// AuthResponse - структура для ответа с токеном
type AuthResponse struct {
	Token string `json:"token"`
}
