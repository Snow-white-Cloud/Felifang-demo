package errors

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// PQError преобразует ошибку PostgreSQL в бизнес-ошибку приложения.
// Обрабатывает различные коды ошибок PostgreSQL и возвращает соответствующие доменные ошибки.
//
// Параметры:
//   - err: исходная ошибка (может быть nil, context error или pq.Error)
//   - operation: название операции, в которой произошла ошибка
//
// Возвращает:
//   - error: бизнес-ошибка с контекстом операции
func PQError(err error, operation string) error {
	// Обработка отмены контекста (таймаут)
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: %w", operation, ErrTimeout)
	}

	// Если ошибка не PostgreSQL, возвращаем как есть
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		return fmt.Errorf("%s: %w", operation, err)
	}

	switch pgErr.Code {
	// Ошибки подключения (класс 08)
	case "08000": // connection_exception
		return fmt.Errorf("%s: ошибка подключения: %w", operation, ErrConnection)
	case "08003": // connection_does_not_exist
		return fmt.Errorf("%s: соединение не существует: %w", operation, ErrConnection)
	case "08006": // connection_failure
		return fmt.Errorf("%s: потеря соединения: %w", operation, ErrConnection)

	// Ошибки при обработке данных (класс 22)
	case "22001": // string_data_right_truncation
		return fmt.Errorf("%s: значение слишком длинное: %w", operation, ErrValidation)
	case "22003": // numeric_value_out_of_range
		return fmt.Errorf("%s: значение вне допустимого диапазона: %w", operation, ErrValidation)
	case "22007": // invalid_datetime_format
		return fmt.Errorf("%s: неверный формат даты/времени: %w", operation, ErrValidation)

	// Нарушения ограничений целостности (класс 23)
	case "23502": // not_null_violation
		return fmt.Errorf("%s: обязательное поле не может быть NULL: %w", operation, ErrValidation)
	case "23503": // foreign_key_violation
		return fmt.Errorf("%s: ссылка на несуществующую запись: %w", operation, ErrNotFound)
	case "23505": // unique_violation
		return fmt.Errorf("%s: %w", operation, ErrNotUnique)
	case "23514": // check_violation
		return fmt.Errorf("%s: нарушение ограничения CHECK: %w", operation, ErrValidation)

	// Недопустимое состояние транзакции (класс 25)
	case "25001": // active_sql_transaction
		return fmt.Errorf("%s: активная транзакция: %w", operation, ErrTransaction)
	case "25006": // read_only_sql_transaction
		return fmt.Errorf("%s: попытка записи в read-only транзакции: %w", operation, ErrTransaction)

	// Deadlock (класс 40)
	case "40P01": // deadlock_detected
		return fmt.Errorf("%s: обнаружен дедлок: %w", operation, ErrTransaction)

	// Неопределено (класс 42)
	case "42P01": // undefined_table
		return fmt.Errorf("%s: таблица не существует: %w", operation, ErrNotFound)
	case "42703": // undefined_column
		return fmt.Errorf("%s: колонка не существует: %w", operation, ErrNotFound)

	// Ошибка синтаксиса (класс 42)
	case "42601": // syntax_error
		return fmt.Errorf("%s: синтаксическая ошибка запроса: %w", operation, ErrInternal)

	// Блокировка недоступна (класс 55)
	case "55P03": // lock_not_available
		return fmt.Errorf("%s: блокировка недоступна: %w", operation, ErrTimeout)

	// Внутренние ошибки (класс 57)
	case "57014": // query_canceled
		return fmt.Errorf("%s: запрос отменён: %w", operation, ErrTimeout)

	default:
		return fmt.Errorf("%s: необработанная ошибка БД (код %s): %w", operation, pgErr.Code, err)
	}
}
