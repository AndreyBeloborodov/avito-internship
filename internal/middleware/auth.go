package middleware

import (
	"context"
	"log"
	"merch-shop/internal/errs"
	"merch-shop/internal/handlers"
	"merch-shop/internal/services"
	"net/http"
	"strings"
)

type ContextKey string

// Определяем ключ для использования в контексте
const usernameKey ContextKey = "username"

// AuthMiddleware проверяет токен и передаёт username в контекст запроса.
func AuthMiddleware(userService *services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handlers.WriteErrorResponse(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Ожидаем формат "Bearer <token>"
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			username, err := userService.ExtractUsernameFromToken(tokenString)
			if err != nil {
				handlers.WriteErrorResponse(w, errs.ErrInvalidToken.Error(), http.StatusUnauthorized)
				log.Println("failed extract username from token ", err)
				return
			}

			// Добавляем username в контекст запроса
			ctx := r.Context()
			ctx = context.WithValue(ctx, usernameKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
