package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"merch-shop/internal/handler"
	"merch-shop/internal/model"
	"merch-shop/internal/repo"
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
		panic("failed to connect to database")
	}
	repo.DB = db
	db.AutoMigrate(&model.User{})

	// Инициализация роутеров
	r := mux.NewRouter()
	r.HandleFunc("/api/auth", handler.Authenticate).Methods("POST")

	http.ListenAndServe(":8080", r)
}
