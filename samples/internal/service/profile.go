package services

import (
	"context"
	"felifang/auth/internal/models"
	apperrors "felifang/pkg/errors"
)

// ProfileService реализует бизнес-логику управления профилем пользователя.
// Включает добавление телефона и другие операции изменения настроек аккаунта.
type ProfileService struct {
	system DBSystem                    // Репозиторий системных операций ОБД
	users  UserRepository              // Репозиторий пользователей
	phone  PhoneVerificationRepository // Репозиторий верификации телефона
}

// NewProfileService создаёт новый экземпляр ProfileService.
// Внедряет все необходимые зависимости.
//
// Параметры:
//   - system: репозиторий системных операций ОБД
//   - userRepo: репозиторий пользователей
//   - phone: репозиторий верификации телефона
//
// Возвращает:
//   - *ProfileService: готовый сервис
func NewProfileService(system DBSystem, userRepo UserRepository, phone PhoneVerificationRepository) *ProfileService {
	return &ProfileService{
		system: system,
		users:  userRepo,
		phone:  phone,
	}
}

// AddPhone добавляет номер телефона пользователю.
// Обновляет телефон и устанавливает статус необходимости верификации.
// Операция выполняется в транзакции.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//   - req: запрос с ID пользователя и номером телефона
//
// Возвращает:
//   - error: ошибка выполнения операции
//
// Ошибки:
//   - apperrors.ErrNotUnique: телефон уже используется другим пользователем
//   - apperrors.ErrNotFound: пользователь не найден
//   - apperrors.ErrTransaction: ошибка подтверждения транзакции
func (serv *ProfileService) AddPhone(ctx context.Context, req *models.AddPhoneRequest) error {
	tx, err := serv.system.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновляем телефон пользователя
	if err := serv.users.UpdateUserPhone(ctx, tx, req.UserID, req.Phone); err != nil {
		return err
	}

	// Указываем необходимость верификации
	if err := serv.phone.NeedPhoneVerified(ctx, tx, req.UserID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperrors.ErrTransaction
	}

	return nil
}
