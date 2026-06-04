package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/config"
	"icaris-logistic/backend/internal/db"
	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/repository"
)

func setupManagerRepo(t *testing.T) (*repository.ManagerRepository, *gorm.DB) {
	t.Helper()

	gdb, err := db.Connect(context.Background(), config.Load().DatabaseURL)
	if err != nil {
		t.Skipf("database unavailable, skipping integration test: %v", err)
	}

	return repository.NewManagerRepository(gdb), gdb
}

func createTestManager(t *testing.T, repo *repository.ManagerRepository, gdb *gorm.DB) *domain.Manager {
	t.Helper()

	manager := &domain.Manager{
		Email:    "test-" + uuid.NewString() + "@icaris.io",
		Name:     "Тест",
		Password: "not-a-real-hash",
	}
	if err := repo.Create(context.Background(), manager); err != nil {
		t.Fatalf("create manager: %v", err)
	}
	t.Cleanup(func() {
		gdb.Delete(&domain.Manager{}, "id = ?", manager.ID)
	})

	return manager
}

func TestManagerRepository_CreateAssignsID(t *testing.T) {
	repo, gdb := setupManagerRepo(t)

	manager := createTestManager(t, repo, gdb)

	if manager.ID == uuid.Nil {
		t.Errorf("expected database-generated ID, got nil UUID")
	}
}

func TestManagerRepository_GetByEmailReturnsCreated(t *testing.T) {
	repo, gdb := setupManagerRepo(t)
	manager := createTestManager(t, repo, gdb)

	got, err := repo.GetByEmail(context.Background(), manager.Email)
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}

	if got.ID != manager.ID {
		t.Errorf("expected manager %s, got %s", manager.ID, got.ID)
	}
}

func TestManagerRepository_GetByEmailUnknownReturnsNotFound(t *testing.T) {
	repo, _ := setupManagerRepo(t)

	_, err := repo.GetByEmail(context.Background(), "missing-"+uuid.NewString()+"@icaris.io")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
