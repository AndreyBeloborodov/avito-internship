package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"merch-shop/internal/handlers"
	"merch-shop/internal/models"
	"merch-shop/internal/repositories"
	"merch-shop/internal/services"
	"net/http"
	"os"
)

func main() {
	// Читаем переменные окружения
	host := os.Getenv("DATABASE_HOST")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	port := os.Getenv("DATABASE_PORT")

	// Формируем DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
	fmt.Println(dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Автоматическая миграция
	if err = db.AutoMigrate(&models.User{}, &models.Merch{}, &models.Purchase{}, models.Transaction{}); err != nil {
		log.Println("failed to auto migrate: %w", err)
	}

	userRepo := repositories.NewUserRepo(db)
	merchRepo := repositories.NewMerchRepo(db)
	userService := services.NewUserService(userRepo)
	merchService := services.NewMerchService(merchRepo)
	userHandler := handlers.NewUserHandler(userService)
	shopHandler := handlers.NewShopHandler(userService, merchService)

	// Инициализация роутеров
	r := mux.NewRouter()
	r.HandleFunc("/api/auth", userHandler.Authenticate).Methods("POST")
	r.HandleFunc("/api/buy/{item}", shopHandler.BuyItem).Methods("GET")
	r.HandleFunc("/api/sendCoin", shopHandler.SendCoin).Methods("GET")

	http.ListenAndServe(":8080", r)
}
