package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"iracus-logistic/backend/internal/domain"
)

// Bot отправляет уведомления менеджеру в Telegram. Структурно удовлетворяет
// service.Notifier, но пакет service не импортируется — связь только по контракту метода.
//
// NOTE: MVP — уведомление идёт в один чат через chatID (не пер-менеджер); см. docs/tech-debt.md
type Bot struct {
	api    *tgbotapi.BotAPI
	chatID int64
	// disabled — режим no-op: токен не задан, ничего не шлём, приложение поднимается в dev.
	disabled bool
}

// New создаёт бот. Пустой token => рабочий no-op Bot (для dev без токена): он логирует
// "bot disabled" один раз и молча проглатывает уведомления. Пустой chatID при заданном
// token — ошибка конфигурации.
//
// NOTE: при заданном token New синхронно вызывает getMe (валидация токена), то есть
// требует доступности Telegram на старте — осознанно для send-only бота.
func New(token, chatID string) (*Bot, error) {
	if token == "" {
		log.Print("bot disabled: TELEGRAM_BOT_TOKEN not set, notifications are no-op")
		return &Bot{disabled: true}, nil
	}

	if chatID == "" {
		return nil, fmt.Errorf("bot: chat id is required when token is set")
	}

	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bot: parse chat id: %w", err)
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		// Ошибка библиотеки может содержать URL с токеном, поэтому не пробрасываем её —
		// возвращаем собственное сообщение без исходной строки.
		return nil, fmt.Errorf("bot: init telegram api failed")
	}

	return &Bot{api: api, chatID: id}, nil
}

// NotifyNewLead шлёт уведомление о лиде в чат менеджера.
//
// NOTE: MVP — библиотека v5 не принимает context, поэтому ctx не ограничивает таймаут и
// не отменяет отправку; зависший запрос блокирует вызывающую горутину; см. docs/tech-debt.md
func (b *Bot) NotifyNewLead(ctx context.Context, lead *domain.Lead) error {
	if b.disabled {
		return nil
	}

	msg := tgbotapi.NewMessage(b.chatID, formatLeadMessage(lead))
	if _, err := b.api.Send(msg); err != nil {
		// Ошибка из Send — *url.Error, чья строка содержит URL с токеном в пути
		// (.../bot<TOKEN>/sendMessage). Намеренно не пробрасываем первопричину, чтобы
		// токен не утёк в лог.
		return fmt.Errorf("bot: send lead notification failed")
	}

	return nil
}

// formatLeadMessage собирает русское сообщение о лиде. Чистая функция без сети —
// тестируется напрямую. Вес/объём показываем только когда NullDecimal валиден.
func formatLeadMessage(lead *domain.Lead) string {
	var b strings.Builder

	b.WriteString("Новый лид с сайта\n\n")
	b.WriteString("Имя: " + lead.Name + "\n")
	b.WriteString("Телефон: " + lead.Phone + "\n")
	b.WriteString("Маршрут: " + lead.FromCity + " → " + lead.ToCity + "\n")

	if lead.Weight.Valid {
		b.WriteString("Вес: " + lead.Weight.Decimal.String() + " кг\n")
	}
	if lead.Volume.Valid {
		b.WriteString("Объём: " + lead.Volume.Decimal.String() + " м³\n")
	}
	if lead.CargoType != "" {
		b.WriteString("Тип груза: " + lead.CargoType + "\n")
	}
	if lead.Comment != "" {
		b.WriteString("Комментарий: " + lead.Comment + "\n")
	}

	b.WriteString("ID: " + lead.ID.String())

	return b.String()
}
