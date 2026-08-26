package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"../config"

	_ "github.com/lib/pq"
)

// AuthStorage представляет хранилище данных аутентификации.
// Содержит пул соединений с базой данных PostgreSQL.
// Используется для всех операций, связанных с пользователями и сессиями.
type AuthStorage struct {
	db  *sql.DB                // Пул соединений с PostgreSQL
	cfg *config.DatabaseConfig // Настройки подключения к PostgreSQL.
}

// NewAuthStorage создаёт новое подключение к базе данных PostgreSQL.
// Инициализирует пул соединений с предварительно настроенными параметрами.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//   - cfg: конфигурация подключения к БД
//
// Возвращает:
//   - *AuthStorage: инициализированное хранилище с активным подключением
//   - error: ошибка при открытии соединения или проверке связи с БД
//
// Ошибки:
//   - Ошибка при формировании DSN-строки
//   - Ошибка при открытии соединения с БД
//   - Ошибка при проверке связи (db.Ping())
func NewAuthStorage(ctx context.Context, cfg *config.DatabaseConfig) (*AuthStorage, error) {
	// Формат: postgres://user:password@host:port/dbname?sslmode=mode
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password,
		cfg.Host, cfg.Port,
		cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Не удалось открыть базу данных: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("Не удалось проверить связь с базой данных: %w", err)
	}

	return &AuthStorage{db: db, cfg: cfg}, nil
}

// Close закрывает соединение с базой данных.
// Должен вызываться при завершении работы приложения.
//
// Возвращает:
//   - error: ошибка при закрытии соединения (обычно nil)
func (s *AuthStorage) Close() error {
	return s.db.Close()
}

// Ping проверяет доступность базы данных.
// Отправляет ping-запрос к PostgreSQL.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//
// Возвращает:
//   - error: ошибка при недоступности БД или превышении таймаута
func (s *AuthStorage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("Не удалось проверить связь с базой данных: %w", err)
	}
	return nil
}

// BeginTx начинает новую транзакцию базы данных.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//   - opts: опции транзакции (уровень изоляции, режим только для чтения и т.д.)
//     может быть nil для использования стандартных настроек
//
// Возвращает:
//   - *sql.Tx: объект транзакции для выполнения запросов
//   - error: ошибка при начале транзакции
func (s *AuthStorage) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}
