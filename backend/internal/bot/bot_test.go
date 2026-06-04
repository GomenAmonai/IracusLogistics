package bot

import (
	"context"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"icaris-logistic/backend/internal/domain"
)

// Сообщение без отправителя (анонимный админ группы, автопересылка) не должно ронять бот:
// обе команды требуют msg.From.ID, и без guard это была паника, валящая весь процесс.
func TestHandleCommandIgnoresMessageWithoutSender(t *testing.T) {
	b := &Bot{disabled: true}
	msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}, Text: "/status", From: nil}

	// Паники быть не должно; deps пустые — при From==nil они не должны вызываться.
	b.handleCommand(context.Background(), RunDeps{}, msg)
}

// callback без отправителя/сообщения (как и команда) не должен ронять бот.
func TestHandleCallbackIgnoresQueryWithoutSender(t *testing.T) {
	b := &Bot{disabled: true}
	cb := &tgbotapi.CallbackQuery{ID: "cb", Data: "status", From: nil, Message: nil}

	// Паники быть не должно; в disabled-режиме answerCallback — no-op, deps не трогаем.
	b.handleCallback(context.Background(), RunDeps{}, cb)
}

func TestMainMenuExposesCoreActions(t *testing.T) {
	var actions []string
	for _, row := range mainMenu().InlineKeyboard {
		for _, button := range row {
			if button.CallbackData != nil {
				actions = append(actions, *button.CallbackData)
			}
		}
	}

	for _, want := range []string{"status", "help", "manager"} {
		if !strings.Contains(strings.Join(actions, ","), want) {
			t.Errorf("menu must expose %q action, got %v", want, actions)
		}
	}
}

func TestHelpTextMentionsStatusCommand(t *testing.T) {
	if !strings.Contains(helpText(), "/status") {
		t.Errorf("help text must mention /status command, got: %q", helpText())
	}
}

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
