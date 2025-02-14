package models

// AuthRequest - структура для запроса аутентификации
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
