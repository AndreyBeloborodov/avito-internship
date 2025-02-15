package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"merch-shop/internal/handlers"
	"merch-shop/internal/middleware"
	"merch-shop/internal/models"
	"merch-shop/internal/repositories"
	"merch-shop/internal/services"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
		log.Println("failed to auto migrate: ", err)
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

	protectedRoutes := r.PathPrefix("/api").Subrouter()
	protectedRoutes.Use(middleware.AuthMiddleware(userService))

	protectedRoutes.HandleFunc("/buy/{item}", shopHandler.BuyItem).Methods("GET")
	protectedRoutes.HandleFunc("/sendCoin", shopHandler.SendCoin).Methods("GET")
	protectedRoutes.HandleFunc("/info", shopHandler.GetUserInfo).Methods("GET")

	// Создаём сервер
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Канал для сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-stop
	log.Println("Shutting down server...")

	// Создаём контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
