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
	CreatedAt time.Time           `gorm:"not null;default:now()" json:"created_at"`
}

type LeadStatus string

const (
	LeadStatusNew       LeadStatus = "new"
	LeadStatusContacted LeadStatus = "contacted"
	LeadStatusConverted LeadStatus = "converted"
	LeadStatusRejected  LeadStatus = "rejected"
)
