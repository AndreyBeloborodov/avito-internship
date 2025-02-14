package handlers

import (
	"encoding/json"
	"errors"
	"merch-shop/internal/errs"
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
	var authReq models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&authReq); err != nil {
		writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.Authenticate(&authReq)
	if errors.Is(err, errs.ErrInvalidPassword) {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// writeErrorResponse - вспомогательная функция для отправки ошибки
func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{Errors: message})
}
