package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/domain"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(ctx context.Context, client *domain.Client) error {
	return r.db.WithContext(ctx).Create(client).Error
}

func (r *ClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	var client domain.Client
	err := r.db.WithContext(ctx).First(&client, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &client, nil
}

// GetByTelegramID ищет клиента по telegram_id — естественному ключу Telegram-авторизации.
func (r *ClientRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Client, error) {
	var client domain.Client
	err := r.db.WithContext(ctx).First(&client, "telegram_id = ?", telegramID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (r *ClientRepository) List(ctx context.Context) ([]domain.Client, error) {
	var clients []domain.Client
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&clients).Error; err != nil {
		return nil, err
	}

	return clients, nil
}
