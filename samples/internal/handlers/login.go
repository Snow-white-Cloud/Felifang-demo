package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"../models"
	"../pkg/valider"

	apperrors "../pkg/errors"
)

// Login обрабатывает запрос на вход пользователя в систему.
// Эндпоинт: POST /login
//
// Ожидает JSON с полями: email, password.
//
// Ответы:
//   - 200 OK: успешный вход, возвращает JWT-токен
//   - 400 Bad Request: неверный формат запроса или ошибка валидации
//   - 401 Unauthorized: неверные email или пароль
//   - 403 Forbidden: email не подтверждён
//   - 500 Internal Server Error: внутренняя ошибка сервера
//
// Пример запроса:
//
//	{
//	  "email": "user@example.com",
//	  "password": "secure123"
//	}
//
// Пример успешного ответа:
//
//	{
//	  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
//	  "token_type": "Bearer"
//	}
//
// Примечание:
//   - Refresh-токен устанавливается в HttpOnly cookie для защиты от XSS
//   - Cookie доступен только на пути /auth/refresh
//   - Secure: true рекомендуется включить в production
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		http.Error(w, "Неверно введены данные", http.StatusBadRequest) // 400
		return
	}
	defer r.Body.Close()

	// Валидация
	if err := valider.Validate.Struct(&req); err != nil {
		log.Printf("Ошибка валидации: %v", err)
		validationErrorResponse(w, err)
		return
	}

	// Аутентификация пользователя
	accessToken, refreshToken, err := h.serviceAuth.LoginUser(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, apperrors.ErrInvalidPassword):
			http.Error(w, "Неверные почта или пароль", http.StatusUnauthorized)
		case errors.Is(err, apperrors.ErrNotVerified):
			http.Error(w, "Email не подтвержден", http.StatusForbidden)
		case errors.Is(err, apperrors.ErrNoRoles):
			log.Printf("ВАЖНО: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		default:
			log.Printf("Ошибка входа: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/auth/refresh",
		HttpOnly: true,
		// Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.serviceAuth.GetRefreshTokenTTL().Seconds()),
	}
	http.SetCookie(w, cookie)

	// Успешный ответ с JWT-токеном
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200
	resp := map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Ошибка отправки ответа: %v", err)
	}
}
