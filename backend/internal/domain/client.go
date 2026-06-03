package domain

import (
	"time"

	"github.com/google/uuid"
)

// Client — клиент, авторизованный через Telegram. Создаётся после подтверждения сделки.
type Client struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TelegramID int64     `gorm:"uniqueIndex;not null" json:"telegram_id"`
	Username   string    `gorm:"type:varchar(255)" json:"username"`
	Name       string    `gorm:"type:varchar(255);not null" json:"name"`
	Phone      string    `gorm:"type:varchar(255)" json:"phone"`
	// LeadID — заявка, из которой вырос клиент. Nullable: клиент мог прийти не через форму.
	LeadID    *uuid.UUID `gorm:"type:uuid" json:"lead_id"`
	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
}
