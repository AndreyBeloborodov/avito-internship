package handler

import (
	"encoding/json"
	"merch-shop/internal/service"
	"net/http"
)

// AuthRequest - структура для запроса аутентификации
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse - структура для ответа с токеном
type AuthResponse struct {
	Token string `json:"token"`
}

// Authenticate - обработчик аутентификации
func Authenticate(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authService := &service.AuthService{}
	authReq := service.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	resp, err := authService.Authenticate(&authReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
