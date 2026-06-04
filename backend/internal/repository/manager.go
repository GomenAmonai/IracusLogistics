package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"icaris-logistic/backend/internal/domain"
)

type ManagerRepository struct {
	db *gorm.DB
}

func NewManagerRepository(db *gorm.DB) *ManagerRepository {
	return &ManagerRepository{db: db}
}

func (r *ManagerRepository) Create(ctx context.Context, manager *domain.Manager) error {
	return r.db.WithContext(ctx).Create(manager).Error
}

// GetByEmail ищет менеджера по email — естественному ключу входа. Нет совпадения →
// domain.ErrNotFound, чтобы вызывающий не зависел от gorm.ErrRecordNotFound.
func (r *ManagerRepository) GetByEmail(ctx context.Context, email string) (*domain.Manager, error) {
	var manager domain.Manager
	err := r.db.WithContext(ctx).First(&manager, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &manager, nil
}
