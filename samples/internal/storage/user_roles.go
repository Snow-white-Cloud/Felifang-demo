package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	apperrors "../pkg/errors"

	"github.com/google/uuid"
)

// UserRoles представляет связь пользователя с ролью.
// Соответствует таблице user_roles в базе данных PostgreSQL.
//
// Поля:
//   - UserID: идентификатор пользователя (часть составного PK)
//   - RoleID: идентификатор роли (часть составного PK)
//   - GrantedAt: дата и время назначения роли
//   - GrantedBy: кто назначил роль (не обязательно)
type UserRoles struct {
	UserID    uuid.UUID           `db:"user_id" json:"user_id"`
	RoleID    int                 `db:"role_id" json:"role_id"`
	GrantedAt time.Time           `db:"granted_at" json:"granted_at"`
	GrantedBy sql.Null[uuid.UUID] `db:"granted_by" json:"granted_by,omitempty"`
}

// AssignUserRole назначает пользователю роль.
// Проверяет существование роли по имени, затем создаёт запись в user_roles.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//   - tx: транзакция PostgreSQL (должна быть открыта)
//   - userID: UUID пользователя, которому назначается роль
//   - grantedBy: UUID пользователя, который назначает роль (может быть uuid.Nil)
//   - role: имя роли
//
// Возвращает:
//   - error: ошибка выполнения SQL-запроса или ошибка PostgreSQL
//
// Ошибки:
//   - apperrors.ErrNotFound: роль с указанным именем не найдена
//   - Ошибка дубликата (нарушение уникальности: роль уже назначена)
//   - Ошибка внешнего ключа (user_id или role_id не существуют)
//   - Любая другая ошибка PostgreSQL (преобразуется через apperrors.PQError)
func (s *AuthStorage) AssignUserRole(ctx context.Context, tx *sql.Tx, userID, grantedBy uuid.UUID, role string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, granted_by)
		SELECT $1, id, $2 FROM roles WHERE name = $3
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := tx.ExecContext(ctx, query, userID, grantedBy, role)
	if err == nil {
		return nil
	}

	return apperrors.PQError(err, "AssignUserRole")
}

// ListUserRoles возвращает список ролей пользователя.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//   - userID: UUID пользователя
//
// Возвращает:
//   - []string: список имён ролей пользователя
//   - error: ошибка выполнения SQL-запроса
func (s *AuthStorage) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `SELECT r.name 
              FROM user_roles ur 
              JOIN roles r ON ur.role_id = r.id 
              WHERE ur.user_id = $1`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("Не удалось запросить роли пользователя: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("Не удалось прочесть роль: %w", err)
		}
		roles = append(roles, name)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации строк: %w", err)
	}

	return roles, nil
}
