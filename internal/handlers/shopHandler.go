package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
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
	// Получаем имя предмета из URL
	vars := mux.Vars(r)
	merchName := vars["item"]

	// Достаём username из токена
	username, err := h.userService.ExtractUsernameFromToken(r)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusUnauthorized)
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
		return
	case errors.Is(err, errs.ErrNotEnoughCoins):
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("merch successfully purchased"))
}

// SendCoin - обработчик отправки монет другому пользователю
func (h *ShopHandler) SendCoin(w http.ResponseWriter, r *http.Request) {
	var sendCoinRequest models.SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&sendCoinRequest); err != nil {
		writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Достаём username из токена
	username, err := h.userService.ExtractUsernameFromToken(r)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = h.userService.SendCoin(username, sendCoinRequest)

	switch {
	case errors.Is(err, errs.ErrUserNotFound):
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, errs.ErrNegativeCoins):
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, errs.ErrNotEnoughCoins):
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The coins were sent successfully"))
}
