package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Lead — заявка с публичного сайта (калькулятора) до конвертации в клиента.
// Создаётся анонимно, без регистрации; менеджер обрабатывает её вручную.
type Lead struct {
	ID        uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string              `gorm:"type:varchar(255);not null" json:"name"`
	Phone     string              `gorm:"type:varchar(255);not null" json:"phone"`
	FromCity  string              `gorm:"type:varchar(255);not null" json:"from_city"`
	ToCity    string              `gorm:"type:varchar(255);not null" json:"to_city"`
	Weight    decimal.NullDecimal `gorm:"type:numeric" json:"weight"`
	Volume    decimal.NullDecimal `gorm:"type:numeric" json:"volume"`
	CargoType string              `gorm:"type:varchar(255)" json:"cargo_type"`
	Comment   string              `gorm:"type:text" json:"comment"`
	Status    LeadStatus          `gorm:"type:varchar(20);not null;default:new" json:"status"`
	// PrivacyNoticeVersion фиксирует редакцию политики, с которой согласился заявитель.
	// PrivacyConsentAt задаётся сервером: времени из браузера доверять нельзя.
	PrivacyNoticeVersion string     `gorm:"type:varchar(64)" json:"privacy_notice_version"`
	PrivacyConsentAt     *time.Time `json:"privacy_consent_at"`
	CreatedAt            time.Time  `gorm:"not null;default:now()" json:"created_at"`
}

type LeadStatus string

const (
	LeadStatusNew       LeadStatus = "new"
	LeadStatusContacted LeadStatus = "contacted"
	LeadStatusConverted LeadStatus = "converted"
	LeadStatusRejected  LeadStatus = "rejected"
)

// IsValid сообщает, входит ли статус в допустимый набор. Сервис проверяет это при смене
// статуса, чтобы вернуть 400, а не словить 500 от CHECK-ограничения в БД.
func (s LeadStatus) IsValid() bool {
	switch s {
	case LeadStatusNew, LeadStatusContacted, LeadStatusConverted, LeadStatusRejected:
		return true
	default:
		return false
	}
}
