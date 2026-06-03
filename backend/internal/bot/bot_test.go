package bot

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"iracus-logistic/backend/internal/domain"
)

func sampleLead() *domain.Lead {
	return &domain.Lead{
		ID:        uuid.New(),
		Name:      "Иван Петров",
		Phone:     "+79990001122",
		FromCity:  "Guangzhou",
		ToCity:    "Moscow",
		Weight:    decimal.NewNullDecimal(decimal.RequireFromString("150.5")),
		Volume:    decimal.NewNullDecimal(decimal.RequireFromString("2.3")),
		CargoType: "Электроника",
		Comment:   "Хрупкий груз",
	}
}

func TestFormatLeadMessageIncludesName(t *testing.T) {
	lead := sampleLead()

	msg := formatLeadMessage(lead)

	if !strings.Contains(msg, "Иван Петров") {
		t.Errorf("message must contain lead name, got: %q", msg)
	}
}

func TestFormatLeadMessageIncludesPhone(t *testing.T) {
	lead := sampleLead()

	msg := formatLeadMessage(lead)

	if !strings.Contains(msg, "+79990001122") {
		t.Errorf("message must contain lead phone, got: %q", msg)
	}
}

func TestFormatLeadMessageIncludesRoute(t *testing.T) {
	lead := sampleLead()

	msg := formatLeadMessage(lead)

	if !strings.Contains(msg, "Guangzhou → Moscow") {
		t.Errorf("message must contain route from→to, got: %q", msg)
	}
}

func TestFormatLeadMessageIncludesWeightWhenValid(t *testing.T) {
	lead := sampleLead()

	msg := formatLeadMessage(lead)

	if !strings.Contains(msg, "150.5") {
		t.Errorf("message must contain weight when valid, got: %q", msg)
	}
}

func TestFormatLeadMessageOmitsWeightWhenNull(t *testing.T) {
	lead := sampleLead()
	lead.Weight = decimal.NullDecimal{}

	msg := formatLeadMessage(lead)

	if strings.Contains(msg, "Вес:") {
		t.Errorf("message must omit weight label when null, got: %q", msg)
	}
}
