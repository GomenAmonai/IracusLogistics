package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/domain"
)

// pgUniqueViolation — код ошибки Postgres для нарушения unique-индекса.
const pgUniqueViolation = "23505"

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(ctx context.Context, client *domain.Client) error {
	err := r.db.WithContext(ctx).Create(client).Error
	if err == nil {
		return nil
	}

	// Транслируем нарушения уникальности в типизированные sentinel'ы, чтобы сервис мог
	// отличить легитимную гонку (telegram_id / lead_id) от настоящего сбоя БД и не маскировал
	// последний под гонку.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		switch pgErr.ConstraintName {
		case "clients_telegram_id_key":
			return domain.ErrClientExists
		case "uq_clients_lead_id":
			return domain.ErrLeadAlreadyClaimed
		}
	}

	return err
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

func (r *ClientRepository) List(ctx context.Context, limit, offset int) ([]domain.Client, error) {
	var clients []domain.Client
	err := r.db.WithContext(ctx).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&clients).Error
	if err != nil {
		return nil, err
	}

	return clients, nil
}
