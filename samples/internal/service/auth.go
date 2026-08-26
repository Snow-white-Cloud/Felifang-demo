package services

...

// PingDB проверяет доступность базы данных.
// Используется для health-чеков.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//
// Возвращает:
//   - error: ошибка при недоступности БД
func (serv *AuthService) PingDB(ctx context.Context) error {
	return serv.storage.system.Ping(ctx)
}

// RegisterUser регистрирует нового пользователя.
// Создаёт пользователя, хеширует пароль, назначает роль "player", уведомляет микросервисы о создании.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//   - req: запрос с данными для регистрации
//
// Возвращает:
//   - error: ошибка при регистрации
//
// Ошибки:
//   - apperrors.ErrNotUnique: email или логин уже заняты
//   - Ошибка создания транзакции
//   - Ошибка хеширования пароля
//   - Ошибка сохранения пользователя
//   - Ошибка назначения роли
//   - Ошибка подтверждения транзакции
func (serv *AuthService) RegisterUser(ctx context.Context, req *models.RegisterRequest) error {
	tx, err := serv.storage.system.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	userID := uuid.New()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), serv.bcryptCostCfg)
	if err != nil {
		return err
	}

	err = serv.storage.users.CreateUser(ctx, tx, userID, req.Email, req.Login, req.Phone, string(passwordHash))
	if err != nil {
		return err
	}

	err = serv.storage.roles.AssignUserRole(ctx, tx, userID, uuid.Nil, "player")
	if err != nil {
		return err
	}

	// TODO: Отправить событие в другие микросервисы

	if err := tx.Commit(); err != nil {
		return apperrors.ErrTransaction
	}

	return nil
}

// VerifyEmail подтверждает email пользователя по коду верификации.
// Проверяет, не подтверждён ли уже email, существование и срок действия кода.
//
// Параметры:
//   - ctx: контекст для управления временем выполнения
//   - req: запрос с email и кодом подтверждения
//
// Возвращает:
//   - error: ошибка верификации
//
// Ошибки:
//   - apperrors.ErrAlreadyVerified: email уже подтверждён
//   - apperrors.ErrNotFound: запись верификации не найдена
//   - apperrors.ErrCodeExpired: код истёк
//   - apperrors.ErrTransaction: ошибка подтверждения транзакции
func (serv *AuthService) VerifyEmail(ctx context.Context, req *models.VerifyEmailRequest) error {
	// Проверяем, не подтверждён ли уже email
	verify, err := serv.storage.email.IsEmailVerified(ctx, req.Email)
	if err == nil && verify {
		return apperrors.ErrAlreadyVerified
	}

	// Получаем запись верификации
	verification, err := serv.storage.email.GetEmailVerification(ctx, req.Email, req.Code)
	if err != nil {
		return err
	}

	// Проверяем, не истёк ли код
	if time.Now().After(verification.ExpiresAt) {
		return apperrors.ErrCodeExpired
	}

	// Отмечаем email как подтверждённый через транзакцию
	tx, err := serv.storage.system.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := serv.storage.email.DoneEmailVerified(ctx, tx, verification.UserID, verification.Code); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperrors.ErrTransaction
	}

	return nil
}