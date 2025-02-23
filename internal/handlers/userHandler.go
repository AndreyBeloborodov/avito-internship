package handlers

import (
	"encoding/json"
	"errors"
	"log"
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
		WriteErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.Authenticate(&authReq)
	if errors.Is(err, errs.ErrInvalidPassword) {
		WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		WriteErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Failed authenticate: ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		WriteErrorResponse(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
		log.Printf("Error encoding response to JSON: %v", err)
	}
}

// WriteErrorResponse - вспомогательная функция для отправки ошибки
func WriteErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(models.ErrorResponse{Errors: message}); err != nil {
		log.Printf("Error encoding response to JSON: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
