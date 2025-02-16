package integration_tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
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
	"testing"
	"time"
)

var db *gorm.DB
var srv *http.Server

func TestMain(m *testing.M) {
	// Загрузить переменные окружения из .env файла
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var cleanup func()
	db, cleanup = setupTestDatabase()
	defer cleanup()

	srv = setupServer(db)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Ждём, чтобы сервер успел запуститься
	time.Sleep(100 * time.Millisecond)

	code := m.Run()

	// Останавливаем сервер
	if err = srv.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}

	os.Exit(code)
}

func setupTestDatabase() (*gorm.DB, func()) {
	dsn := GetTestDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	// Автомиграция
	if err = db.AutoMigrate(&models.User{}, &models.Merch{}, &models.Purchase{}, &models.Transaction{}); err != nil {
		log.Printf("Error during DB migration: %v", err)
	}

	// Функция очистки данных после тестов
	cleanup := func() {
		db.Exec("TRUNCATE users, merches, purchases, transactions RESTART IDENTITY CASCADE")
	}

	return db, cleanup
}

func setupServer(db *gorm.DB) *http.Server {
	userRepo := repositories.NewUserRepo(db)
	merchRepo := repositories.NewMerchRepo(db)
	userService := services.NewUserService(userRepo)
	merchService := services.NewMerchService(merchRepo)
	shopHandler := handlers.NewShopHandler(userService, merchService)
	userHandler := handlers.NewUserHandler(userService)

	r := mux.NewRouter()
	r.HandleFunc("/api/auth", userHandler.Authenticate).Methods("POST")

	protectedRoutes := r.PathPrefix("/api").Subrouter()
	protectedRoutes.Use(middleware.AuthMiddleware(userService))

	protectedRoutes.HandleFunc("/buy/{item}", shopHandler.BuyItem).Methods("GET")
	protectedRoutes.HandleFunc("/sendCoin", shopHandler.SendCoin).Methods("POST")
	protectedRoutes.HandleFunc("/info", shopHandler.GetUserInfo).Methods("GET")

	return &http.Server{
		Addr:    GetTestServerPort(),
		Handler: r,
	}
}

func authenticateUser(username, password string) (string, error) {
	// Подготовка тела запроса для авторизации
	authBody, _ := json.Marshal(models.AuthRequest{
		Username: username,
		Password: password,
	})

	// Формируем запрос авторизации
	reqAuth, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("http://localhost%s/api/auth", srv.Addr),
		bytes.NewBuffer(authBody),
	)
	reqAuth.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	respAuth, err := client.Do(reqAuth)
	if err != nil {
		return "", fmt.Errorf("could not send auth request: %v", err)
	}
	defer func() {
		if err = respAuth.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	// Проверяем статус-код авторизации
	if respAuth.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth failed, status: %d", respAuth.StatusCode)
	}

	// Парсим JSON-ответ для получения токена
	var authResp models.AuthResponse
	if err := json.NewDecoder(respAuth.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("could not parse auth response: %v", err)
	}

	return authResp.Token, nil
}

func TestAuthenticationIntegration(t *testing.T) {
	// Очищаем данные перед тестом
	db.Exec("TRUNCATE users RESTART IDENTITY CASCADE")

	// Создаём тестового пользователя
	user := &models.User{
		Username: "test_user",
		Password: "test_pass",
	}

	dbUser := user
	dbUser.Password, _ = services.GetHashPassword(dbUser.Password)
	db.Create(dbUser)

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectToken    bool
	}{
		{
			name:           "Successful authentication",
			username:       "test_user",
			password:       "test_pass",
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name:           "Incorrect password",
			username:       "test_user",
			password:       "wrong_pass",
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка тела запроса
			authBody, _ := json.Marshal(models.AuthRequest{
				Username: tt.username,
				Password: tt.password,
			})

			// Формируем запрос
			req, _ := http.NewRequest(
				"POST",
				fmt.Sprintf("http://localhost%s/api/auth", srv.Addr),
				bytes.NewBuffer(authBody),
			)
			req.Header.Set("Content-Type", "application/json")

			// Отправляем запрос
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("could not send auth request: %v", err)
			}
			defer func() {
				if err = resp.Body.Close(); err != nil {
					log.Printf("Error closing response body: %v", err)
				}
			}()

			// Проверяем HTTP-код
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Если аутентификация успешна, проверяем наличие токена
			if tt.expectToken {
				var authResp models.AuthResponse
				err := json.NewDecoder(resp.Body).Decode(&authResp)
				if err != nil {
					t.Fatalf("could not decode auth response: %v", err)
				}
				assert.NotEmpty(t, authResp.Token)
			}
		})
	}
}

func TestBuyMerchIntegration(t *testing.T) {
	// Очищаем данные перед тестом
	db.Exec("TRUNCATE users, merches, purchases RESTART IDENTITY CASCADE")

	// Создаём тестовые данные
	user := &models.User{
		Username: "test_user",
		Password: "test_pass",
		Coins:    1000,
	}

	merch := &models.Merch{
		Name:  "T-Shirt",
		Price: 500,
	}
	db.Create(merch)

	tests := []struct {
		name           string
		username       string
		userpass       string
		merchName      string
		coinsBefore    int
		coinsAfter     int
		expectedStatus int
	}{
		{
			name:           "Successful purchase",
			username:       user.Username,
			userpass:       user.Password,
			merchName:      merch.Name,
			coinsBefore:    1000,
			coinsAfter:     500,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Not enough coins",
			username:       user.Username,
			userpass:       user.Password,
			merchName:      merch.Name,
			coinsBefore:    50,
			coinsAfter:     50,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Merch not found",
			username:       user.Username,
			userpass:       user.Password,
			merchName:      "NonExistentItem",
			coinsBefore:    1000,
			coinsAfter:     1000,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем нужное количество монет перед тестом
			db.Model(&models.User{}).Where("username = ?", tt.username).Update("coins", tt.coinsBefore)

			// Авторизуемся и получаем токен
			token, err := authenticateUser(tt.username, tt.userpass)
			if err != nil {
				t.Fatalf("authentication failed: %v", err)
			}

			// Формируем запрос покупки с токеном
			reqBuy, _ := http.NewRequest(
				"GET",
				fmt.Sprintf("http://localhost%s/api/buy/%s", srv.Addr, tt.merchName),
				nil,
			)
			reqBuy.Header.Set("Authorization", "Bearer "+token)

			// Отправляем запрос на покупку
			client := &http.Client{}
			respBuy, err := client.Do(reqBuy)
			if err != nil {
				t.Fatalf("could not send buy request: %v", err)
			}
			defer func() {
				if err = respBuy.Body.Close(); err != nil {
					log.Printf("Error closing response body: %v", err)
				}
			}()

			// Проверяем баланс пользователя после покупки
			var updatedUser models.User
			db.First(&updatedUser, "username = ?", tt.username)
			assert.Equal(t, tt.coinsAfter, updatedUser.Coins)

			// Проверяем HTTP-код ответа
			assert.Equal(t, tt.expectedStatus, respBuy.StatusCode)
		})
	}
}

func TestTransferCoinsIntegration(t *testing.T) {
	// Очищаем данные перед тестом
	db.Exec("TRUNCATE users, purchases RESTART IDENTITY CASCADE")

	// Создаём тестовые данные
	sender := &models.User{
		Username: "sender_user",
		Password: "sender_pass",
		Coins:    1000,
	}
	receiver := &models.User{
		Username: "receiver_user",
		Password: "receiver_pass",
		Coins:    500,
	}

	tests := []struct {
		name                string
		senderUsername      string
		senderPassword      string
		receiverUsername    string
		amount              int
		coinsBeforeSender   int
		coinsAfterSender    int
		coinsBeforeReceiver int
		coinsAfterReceiver  int
		expectedStatus      int
		tokenOverride       string
	}{
		{
			name:                "Successful transfer",
			senderUsername:      sender.Username,
			senderPassword:      sender.Password,
			receiverUsername:    receiver.Username,
			amount:              200,
			coinsBeforeSender:   1000,
			coinsAfterSender:    800,
			coinsBeforeReceiver: 500,
			coinsAfterReceiver:  700,
			expectedStatus:      http.StatusOK,
		},
		{
			name:                "Transfer to self",
			senderUsername:      sender.Username,
			senderPassword:      sender.Password,
			receiverUsername:    sender.Username,
			amount:              200,
			coinsBeforeSender:   1000,
			coinsAfterSender:    1000,
			coinsBeforeReceiver: 1000,
			coinsAfterReceiver:  1000,
			expectedStatus:      http.StatusBadRequest,
		},
		{
			name:                "Not enough coins",
			senderUsername:      sender.Username,
			senderPassword:      sender.Password,
			receiverUsername:    receiver.Username,
			amount:              2000,
			coinsBeforeSender:   1000,
			coinsAfterSender:    1000,
			coinsBeforeReceiver: 500,
			coinsAfterReceiver:  500,
			expectedStatus:      http.StatusBadRequest,
		},
		{
			name:                "Receiver not found",
			senderUsername:      sender.Username,
			senderPassword:      sender.Password,
			receiverUsername:    "unknown_user",
			amount:              200,
			coinsBeforeSender:   1000,
			coinsAfterSender:    1000,
			coinsBeforeReceiver: 500,
			coinsAfterReceiver:  500,
			expectedStatus:      http.StatusBadRequest,
		},
		{
			name:                "Invalid token",
			senderUsername:      sender.Username,
			senderPassword:      sender.Password,
			receiverUsername:    receiver.Username,
			amount:              200,
			coinsBeforeSender:   1000,
			coinsAfterSender:    1000,
			coinsBeforeReceiver: 500,
			coinsAfterReceiver:  500,
			expectedStatus:      http.StatusUnauthorized,
			tokenOverride:       "invalid.token.here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Авторизуем отправителя, если не указан фиктивный токен
			token := tt.tokenOverride
			if token == "" {
				var err error
				token, err = authenticateUser(tt.senderUsername, tt.senderPassword)
				if err != nil {
					t.Fatalf("authentication failed: %v", err)
				}
				if tt.name != "Receiver not found" {
					_, err = authenticateUser(tt.receiverUsername, tt.senderPassword)
					if err != nil {
						t.Fatalf("authentication failed: %v", err)
					}
				}
			}

			// Устанавливаем начальные балансы
			db.Model(&models.User{}).Where("username = ?", tt.senderUsername).Update("coins", tt.coinsBeforeSender)
			db.Model(&models.User{}).Where("username = ?", tt.receiverUsername).Update("coins", tt.coinsBeforeReceiver)

			// Подготавливаем тело запроса
			transferBody, _ := json.Marshal(models.SendCoinRequest{
				ToUser: tt.receiverUsername,
				Amount: tt.amount,
			})

			// Формируем запрос на перевод монет
			reqTransfer, _ := http.NewRequest(
				"POST",
				fmt.Sprintf("http://localhost%s/api/sendCoin", srv.Addr),
				bytes.NewBuffer(transferBody),
			)
			reqTransfer.Header.Set("Content-Type", "application/json")
			reqTransfer.Header.Set("Authorization", "Bearer "+token)

			// Отправляем запрос
			client := &http.Client{}
			respTransfer, err := client.Do(reqTransfer)
			if err != nil {
				t.Fatalf("could not send transfer request: %v", err)
			}
			defer func() {
				if err = respTransfer.Body.Close(); err != nil {
					log.Printf("Error closing response body: %v", err)
				}
			}()

			// Проверяем баланс отправителя после операции (кроме Invalid Token)
			if tt.name != "Invalid token" {
				var updatedSender models.User
				db.First(&updatedSender, "username = ?", tt.senderUsername)
				assert.Equal(t, tt.coinsAfterSender, updatedSender.Coins)

				// Проверяем баланс получателя (если он существует)
				var updatedReceiver models.User
				if err := db.First(&updatedReceiver, "username = ?", tt.receiverUsername).Error; err == nil {
					assert.Equal(t, tt.coinsAfterReceiver, updatedReceiver.Coins)
				}
			}

			// Проверяем HTTP-код ответа
			assert.Equal(t, tt.expectedStatus, respTransfer.StatusCode)
		})
	}
}

func TestGetUserInfoIntegration(t *testing.T) {
	// Очищаем данные перед тестом
	db.Exec("TRUNCATE users, transactions, purchases, merches RESTART IDENTITY CASCADE")

	// Создаём тестового пользователя
	user1 := &models.User{
		Username: "test_user",
		Password: "test_pass",
		Coins:    1000,
	}
	user1.ID = 1

	token, err := authenticateUser(user1.Username, user1.Password)
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	// Создаём тестового пользователя
	user2 := &models.User{
		Username: "test_user_2",
		Password: "test_pass_2",
		Coins:    1000,
	}
	user2.ID = 2

	_, err = authenticateUser(user2.Username, user2.Password)
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	// Создаём тестовый мерч и покупку
	merch := &models.Merch{
		Name:  "T-Shirt",
		Price: 500,
	}
	db.Create(merch)

	purchase := &models.Purchase{
		UserID:  user1.ID,
		MerchID: merch.ID,
	}
	db.Create(purchase)

	// Создаём тестовую транзакцию монет
	transaction := &models.Transaction{
		SenderId:   user1.ID,
		ReceiverId: user2.ID,
		Amount:     200,
	}
	db.Create(transaction)

	tests := []struct {
		name           string
		username       string
		userpass       string
		expectedCoins  int
		expectedItems  []models.Item
		expectedStatus int
	}{
		{
			name:           "Successful info retrieval",
			username:       user1.Username,
			userpass:       user1.Password,
			expectedCoins:  1000,
			expectedItems:  []models.Item{{Type: "T-Shirt", Quantity: 1}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized request",
			username:       "unknown_user",
			userpass:       "wrong_pass",
			expectedCoins:  0,
			expectedItems:  nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Формируем запрос на получение информации
			reqInfo, _ := http.NewRequest(
				"GET",
				fmt.Sprintf("http://localhost%s/api/info", srv.Addr),
				nil,
			)

			// Добавляем заголовок с токеном, если есть
			if tt.name != "Unauthorized request" {
				reqInfo.Header.Set("Authorization", "Bearer "+token)
			}

			// Отправляем запрос
			client := &http.Client{}
			respInfo, err := client.Do(reqInfo)
			if err != nil {
				t.Fatalf("could not send request: %v", err)
			}
			defer func() {
				if err = respInfo.Body.Close(); err != nil {
					log.Printf("Error closing response body: %v", err)
				}
			}()

			// Проверяем HTTP-код ответа
			assert.Equal(t, tt.expectedStatus, respInfo.StatusCode)

			// Если запрос успешен, проверяем содержимое ответа
			if respInfo.StatusCode == http.StatusOK {
				var info models.InfoResponse
				err := json.NewDecoder(respInfo.Body).Decode(&info)
				if err != nil {
					t.Fatalf("could not decode response: %v", err)
				}

				// Проверяем баланс
				assert.Equal(t, tt.expectedCoins, info.Coins)

				// Проверяем инвентарь
				assert.Equal(t, tt.expectedItems, info.Inventory)
			}
		})
	}
}
