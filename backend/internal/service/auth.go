package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"iracus-logistic/backend/internal/domain"
)

// ErrInvalidCredentials — неверная пара email/пароль. Один и тот же на оба случая (нет
// такого email ИЛИ неверный пароль), чтобы не подсказывать атакующему, какие email есть.
var ErrInvalidCredentials = errors.New("invalid credentials")

// dummyHash — валидный bcrypt-хеш для холостого сравнения в ветке «нет такого менеджера»:
// прогоняем его, чтобы время логина не выдавало, существует ли email (анти-enumeration).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

// ManagerStore — то, что AuthService требует от хранилища менеджеров. Интерфейс на
// стороне потребителя (как LeadStore): сервис не зависит от пакета repository.
type ManagerStore interface {
	GetByEmail(ctx context.Context, email string) (*domain.Manager, error)
	Create(ctx context.Context, manager *domain.Manager) error
}

type AuthService struct {
	managers ManagerStore
	secret   []byte
	ttl      time.Duration
}

func NewAuthService(managers ManagerStore, secret string, ttl time.Duration) *AuthService {
	return &AuthService{managers: managers, secret: []byte(secret), ttl: ttl}
}

// Login проверяет учётные данные и при успехе возвращает подписанный JWT.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)

	manager, err := s.managers.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return "", err
	}
	if manager == nil {
		// Нет менеджера (ErrNotFound или nil-результат). Всё равно прогоняем bcrypt против
		// фиктивного хеша, чтобы время ответа не выдавало существование email.
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return "", ErrInvalidCredentials
	}

	// CompareHashAndPassword сравнивает за постоянное время (защита от тайминг-атак).
	if err := bcrypt.CompareHashAndPassword([]byte(manager.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.issueToken(manager.ID)
}

// CreateManager хеширует пароль и сохраняет менеджера. Единственное место, где рождается
// bcrypt-хеш — поэтому CLI создания менеджера тоже идёт через этот метод.
func (s *AuthService) CreateManager(ctx context.Context, email, name, password string) (*domain.Manager, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)
	if email == "" || name == "" || password == "" {
		return nil, fmt.Errorf("%w: email, name and password are required", ErrValidation)
	}
	if len(password) > 72 {
		// bcrypt молча усекает пароль длиннее 72 байт — отклоняем явно, чтобы два разных
		// длинных пароля с общим префиксом не оказались взаимозаменяемыми.
		return nil, fmt.Errorf("%w: password must be at most 72 bytes", ErrValidation)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	manager := &domain.Manager{
		Email:    email,
		Name:     name,
		Password: string(hash),
	}
	if err := s.managers.Create(ctx, manager); err != nil {
		return nil, err
	}

	return manager, nil
}

// issueToken кладёт id менеджера в Subject стандартных claims и подписывает HS256.
func (s *AuthService) issueToken(managerID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   managerID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.secret)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
