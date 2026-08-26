package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"../pkg/valider"

	"../models"
	apperrors "../pkg/errors"
)

// VerifyEmail обрабатывает запрос на подтверждение email.
// Эндпоинт: POST /verify-email
//
// Ожидает JSON с полями: email, code.
//
// Ответы:
//   - 200 OK: email успешно подтверждён
//   - 400 Bad Request: неверный формат запроса, ошибка валидации,
//     неверный email/код, код истёк или email уже подтверждён
//   - 500 Internal Server Error: внутренняя ошибка сервера
//
// Пример запроса:
//
//	{
//	  "email": "user@example.com",
//	  "code": "123123"
//	}
//
// Пример успешного ответа:
//
//	{
//	  "message": "Email подтверждён"
//	}
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyEmailRequest
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

	// Подтверждение email
	err := h.serviceAuth.VerifyEmail(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			http.Error(w, "Неверный email или код", http.StatusBadRequest)
		case errors.Is(err, apperrors.ErrCodeExpired):
			http.Error(w, "Код подтверждения истёк", http.StatusBadRequest)
		case errors.Is(err, apperrors.ErrAlreadyVerified):
			http.Error(w, "Email уже подтверждён", http.StatusBadRequest)
		default:
			log.Printf("Ошибка верификации email: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Email подтверждён",
	}); err != nil {
		log.Printf("Ошибка отправки ответа: %v", err)
	}
}
