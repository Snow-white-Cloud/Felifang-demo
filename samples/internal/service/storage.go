package services

import (
	"context"
	"database/sql"

	storage "../storage/postgresql"

	"github.com/google/uuid"
)

// ===== Интерфейсы =====

// DBSystem — управление подключением и транзакциями ОБД
type DBSystem interface {
	// Ping проверяет доступность базы данных.
	Ping(ctx context.Context) error
	// BeginTx начинает новую транзакцию базы данных с контекстом.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// UserRepository — работа с пользователями
type UserRepository interface {
	// CreateUser создаёт нового пользователя.
	CreateUser(ctx context.Context, tx *sql.Tx, id uuid.UUID, email, login, phone, passwordHash string) error
	// GetUserByEmail возвращает пользователя по email.
	GetUserByEmail(ctx context.Context, email string) (*storage.User, error)
	// GetUserIDByEmail возвращает ID пользователя по email.
	GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error)
	// CheckLoginExists проверяет существование логина.
	CheckLoginExists(ctx context.Context, login string) (bool, error)
	// GetUserByPhone возвращает пользователя по телефону.
	GetUserByPhone(ctx context.Context, phone string) (*storage.User, error)
	// GetUserIDByPhone возвращает ID пользователя по телефону.
	GetUserIDByPhone(ctx context.Context, phone string) (uuid.UUID, error)
	// UpdateUserPhone обновляет телефон пользователя.
	UpdateUserPhone(ctx context.Context, tx *sql.Tx, userID uuid.UUID, phone string) error
}
