package domain

import (
	"time"

	"github.com/google/uuid"
)

// Notification — исходящее Telegram-уведомление в outbox. Уведомление сначала
// сохраняется в БД (текст рендерится при постановке), отдельный диспетчер шлёт его с
// ретраями — сбой Telegram или рестарт процесса больше не теряет сообщение
// (техдолг #11/#15). Kind — человекочитаемая метка источника для диагностики.
type Notification struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Kind          string             `gorm:"type:varchar(30);not null" json:"kind"`
	ChatID        int64              `gorm:"not null" json:"chat_id"`
	Text          string             `gorm:"type:text;not null" json:"text"`
	Status        NotificationStatus `gorm:"type:varchar(10);not null;default:pending" json:"status"`
	Attempts      int                `gorm:"not null;default:0" json:"attempts"`
	NextAttemptAt time.Time          `gorm:"not null;default:now()" json:"next_attempt_at"`
	LastError     string             `gorm:"type:text" json:"last_error"`
	SentAt        *time.Time         `json:"sent_at"`
	CreatedAt     time.Time          `gorm:"not null;default:now()" json:"created_at"`
}

// NotificationStatus: pending — ждёт отправки (или ретрая), sent — доставлено,
// failed — исчерпаны попытки, требует ручного внимания.
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)
