package handlers

import "merch-shop/internal/services"

type ShopHandler struct {
	userService  *services.UserService
	merchService *services.MerchService
}

func NewShopHandler(userService *services.UserService, merchService *services.MerchService) *ShopHandler {
	return &ShopHandler{userService: userService, merchService: merchService}
}
