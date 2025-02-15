package middleware

import (
	"context"
	"merch-shop/internal/services"
	"net/http"
	"strings"
)

// AuthMiddleware проверяет токен и передаёт username в контекст запроса.
func AuthMiddleware(userService *services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Ожидаем формат "Bearer <token>"
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			username, err := userService.ExtractUsernameFromToken(tokenString)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Добавляем username в контекст запроса
			ctx := r.Context()
			ctx = context.WithValue(ctx, "username", username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
