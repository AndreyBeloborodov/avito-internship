package handlers

import (
	"encoding/json"
	"merch-shop/internal/models"
	"merch-shop/internal/services"
	"net/http"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Authenticate - обработчик аутентификации
func (h *UserHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authReq := models.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	resp, err := h.userService.Authenticate(&authReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
