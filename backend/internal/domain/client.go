package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrClientExists — клиент с таким telegram_id уже зарегистрирован (нарушен unique по
// telegram_id). Для Register это легитимная гонка: перечитать и вернуть существующего.
var ErrClientExists = errors.New("client already exists")

// ErrLeadAlreadyClaimed — этот lead уже привязан к другому клиенту (нарушен partial-unique
// uq_clients_lead_id). Register по нему заводит клиента без привязки к заявке.
var ErrLeadAlreadyClaimed = errors.New("lead already claimed by another client")

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
