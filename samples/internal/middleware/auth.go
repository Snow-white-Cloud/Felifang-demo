package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"../pkg/jwtpub"

	"github.com/google/uuid"
)

// contextKey определяет тип для ключей контекста.
// Используется для избежания коллизий с другими пакетами.
type contextKey string

// userIDKey — ключ для хранения ID пользователя в контексте запроса.
const userIDKey contextKey = "user_id"

// AuthMiddleware создаёт middleware для проверки авторизации через JWT.
// При успешной проверке добавляет user_id в контекст запроса.
//
// Параметры:
//   - keys: ключи для валидации JWT (публичные)
//   - cache: кэш для ускорения проверки токенов
//
// Возвращает:
//   - func(http.Handler) http.Handler: middleware-функция
func AuthMiddleware(keys *jwtpub.Keys, cache *jwtpub.TokenCache, cfgJWTCache *jwtpub.JWTCacheValidationConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("AuthMiddleware: нету Authorization header")
				http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				log.Printf("AuthMiddleware: неверный формат: %v", parts)
				http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			userIDString, err := jwtpub.ValidateTokenWithCache(cfgJWTCache, parts[1], keys, cache)
			if err != nil {
				log.Printf("AuthMiddleware: ошибка валидации: %v", err)
				http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(userIDString)
			if err != nil {
				log.Printf("AuthMiddleware: неверный UUID формат: %v", err)
				http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID возвращает user_id из контекста.
//
// Параметры:
//   - ctx: контекст запроса (содержит user_id после прохождения AuthMiddleware)
//
// Возвращает:
//   - uuid.UUID: ID пользователя
//   - bool: true — если значение найдено и имеет тип uuid.UUID, иначе false
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
