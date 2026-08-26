package middleware

import (
	"context"
	"net/http"
	"time"
)

// WithTimeout создаёт middleware для установки таймаута на выполнение запроса.
// Если обработчик не успевает завершиться за указанное время, контекст отменяется.
//
// Параметры:
//   - timeout: максимальное время выполнения запроса
//
// Возвращает:
//   - func(http.Handler) http.Handler: middleware-функция
func WithTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
