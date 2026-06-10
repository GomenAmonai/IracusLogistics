package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/domain"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Enqueue(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// Due возвращает уведомления, готовые к отправке. Без блокировок: диспетчер один на
// процесс, а процесс один на деплой; упавшая между отправкой и MarkSent строка останется
// pending и уйдёт повторно (at-least-once — возможный дубль сообщения приемлем).
func (r *NotificationRepository) Due(ctx context.Context, limit int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.WithContext(ctx).
		Where("status = ? and next_attempt_at <= now()", domain.NotificationStatusPending).
		Order("created_at asc, id asc").
		Limit(limit).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) MarkSent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":  domain.NotificationStatusSent,
			"sent_at": gorm.Expr("now()"),
		}).Error
}

// MarkFailed фиксирует неудачную попытку: инкремент счётчика, текст ошибки и время
// следующего ретрая. final переводит в терминальный failed — диспетчер больше не возьмёт.
func (r *NotificationRepository) MarkFailed(ctx context.Context, id uuid.UUID, lastError string, nextAttemptAt time.Time, final bool) error {
	updates := map[string]any{
		"attempts":        gorm.Expr("attempts + 1"),
		"last_error":      lastError,
		"next_attempt_at": nextAttemptAt,
	}
	if final {
		updates["status"] = domain.NotificationStatusFailed
	}

	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}
