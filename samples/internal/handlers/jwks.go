package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// JWKS возвращает публичные ключи в формате JWKS (JSON Web Key Set).
// Эндпоинт: GET /.well-known/jwks.json
//
// Используется клиентами для верификации JWT-токенов.
// Возвращает список публичных ключей, которыми подписаны токены.
//
// Ответы:
//   - 200 OK: успешный возврат JWKS
//   - 500 Internal Server Error: внутренняя ошибка сервера
//
// Пример ответа:
//
//	{
//	  "keys": [
//	    {
//	      "kty": "RSA",
//	      "kid": "key-id",
//	      "n": "modulus",
//	      "e": "AQAB",
//	      "alg": "RS256",
//	      "use": "sig"
//	    }
//	  ]
//	}
func (h *AuthHandler) JWKS(w http.ResponseWriter, r *http.Request) {
	jwksResponse, err := h.serviceAuth.ListJWKS(r.Context())
	if err != nil {
		log.Printf("Ошибка получения JWKS: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(jwksResponse); err != nil {
		log.Printf("Ошибка отправки JWKS: %v", err)
	}
}
