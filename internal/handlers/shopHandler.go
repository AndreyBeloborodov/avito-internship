package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"merch-shop/internal/errs"
	"merch-shop/internal/models"
	"merch-shop/internal/services"
	"net/http"
)

type ShopHandler struct {
	userService  *services.UserService
	merchService *services.MerchService
}

func NewShopHandler(userService *services.UserService, merchService *services.MerchService) *ShopHandler {
	return &ShopHandler{userService: userService, merchService: merchService}
}

// BuyItem - обработчик покупки предмета
func (h *ShopHandler) BuyItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	merchName := vars["item"]

	// Достаём username из контекста
	username, ok := r.Context().Value("username").(string)
	if !ok {
		writeErrorResponse(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	merch, err := h.merchService.GetMerchByName(merchName)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.userService.BuyMerch(username, merch)
	switch {
	case errors.Is(err, errs.ErrUserNotFound):
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, errs.ErrNotEnoughCoins):
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
	default:
		if err != nil {
			writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			log.Println("failed to buy merch: ", err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("merch successfully purchased"))
}

// SendCoin - обработчик отправки монет другому пользователю
func (h *ShopHandler) SendCoin(w http.ResponseWriter, r *http.Request) {
	// Декодируем JSON-запрос
	var sendCoinRequest models.SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&sendCoinRequest); err != nil {
		writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем username из контекста
	username, ok := r.Context().Value("username").(string)
	if !ok {
		writeErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Нельзя отправлять монеты самому себе
	if username == sendCoinRequest.ToUser {
		writeErrorResponse(w, "You can't send coins to yourself.", http.StatusBadRequest)
		return
	}

	// Вызываем сервис для отправки монет
	err := h.userService.SendCoin(username, sendCoinRequest)

	switch {
	case errors.Is(err, errs.ErrUserNotFound),
		errors.Is(err, errs.ErrNegativeCoins),
		errors.Is(err, errs.ErrNotEnoughCoins),
		errors.Is(err, errs.ErrSendCoinsToYourself):
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err != nil {
		writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		log.Println("failed to send coins: ", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The coins were sent successfully"))
}

// GetUserInfo - обработчик получения информации о монетах, инвентаре и истории транзакций
func (h *ShopHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	// Получаем username из контекста
	username, ok := r.Context().Value("username").(string)
	if !ok {
		writeErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем информацию о пользователе
	info, err := h.userService.GetUserInfo(username)
	if err != nil {
		writeErrorResponse(w, "Failed to fetch user info", http.StatusInternalServerError)
		log.Println("failed to get user info:", err)
		return
	}

	// Отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}
