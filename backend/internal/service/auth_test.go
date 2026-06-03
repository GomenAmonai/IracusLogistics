package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"iracus-logistic/backend/internal/domain"
	"iracus-logistic/backend/internal/service"
)

const testSecret = "test-secret"

type stubManagerStore struct {
	manager *domain.Manager
	getErr  error
	created *domain.Manager
}

func (s *stubManagerStore) GetByEmail(ctx context.Context, email string) (*domain.Manager, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.manager, nil
}

func (s *stubManagerStore) Create(ctx context.Context, manager *domain.Manager) error {
	s.created = manager
	return nil
}

func storeWithManager(t *testing.T, password string) (*stubManagerStore, *domain.Manager) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	manager := &domain.Manager{ID: uuid.New(), Email: "manager@iracus.io", Name: "Менеджер", Password: string(hash)}
	return &stubManagerStore{manager: manager}, manager
}

func TestAuthService_LoginUnknownEmailReturnsInvalidCredentials(t *testing.T) {
	store := &stubManagerStore{getErr: domain.ErrNotFound}
	svc := service.NewAuthService(store, testSecret, time.Hour)

	_, err := svc.Login(context.Background(), "ghost@iracus.io", "whatever")

	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_LoginWrongPasswordReturnsInvalidCredentials(t *testing.T) {
	store, manager := storeWithManager(t, "correct-horse")
	svc := service.NewAuthService(store, testSecret, time.Hour)

	_, err := svc.Login(context.Background(), manager.Email, "wrong-password")

	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_LoginSuccessIssuesTokenForManager(t *testing.T) {
	store, manager := storeWithManager(t, "correct-horse")
	svc := service.NewAuthService(store, testSecret, time.Hour)

	tokenStr, err := svc.Login(context.Background(), manager.Email, "correct-horse")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	var claims jwt.RegisteredClaims
	if _, err := jwt.ParseWithClaims(tokenStr, &claims, func(*jwt.Token) (any, error) {
		return []byte(testSecret), nil
	}); err != nil {
		t.Fatalf("parse issued token: %v", err)
	}
	if claims.Subject != manager.ID.String() {
		t.Errorf("expected subject %q, got %q", manager.ID.String(), claims.Subject)
	}
}

func TestAuthService_CreateManagerStoresBcryptHash(t *testing.T) {
	store := &stubManagerStore{}
	svc := service.NewAuthService(store, testSecret, time.Hour)

	if _, err := svc.CreateManager(context.Background(), "new@iracus.io", "Новый", "s3cret"); err != nil {
		t.Fatalf("create manager: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(store.created.Password), []byte("s3cret")); err != nil {
		t.Errorf("stored password is not a bcrypt hash of the input: %v", err)
	}
}

func TestAuthService_CreateManagerNormalizesEmail(t *testing.T) {
	store := &stubManagerStore{}
	svc := service.NewAuthService(store, testSecret, time.Hour)

	if _, err := svc.CreateManager(context.Background(), "  Mixed@Case.IO  ", "Имя", "pw"); err != nil {
		t.Fatalf("create manager: %v", err)
	}

	if store.created.Email != "mixed@case.io" {
		t.Errorf("expected normalized email %q, got %q", "mixed@case.io", store.created.Email)
	}
}
