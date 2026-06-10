package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartWebhook переводит бота в webhook-режим: регистрирует URL у Telegram (setWebhook с
// secret_token — Telegram будет слать его в заголовке каждого запроса) и запоминает deps
// для обработки входящих апдейтов через ProcessUpdateJSON. Run при этом не вызывается —
// источник апдейтов теперь HTTP-ручка, а не long polling.
//
// NOTE: WebhookConfig v5.5.1 не знает про secret_token (Bot API 6.0), поэтому setWebhook
// уходит через MakeRequest с явными параметрами.
func (b *Bot) StartWebhook(deps RunDeps, url, secret string) error {
	if b.disabled {
		return nil
	}

	params := tgbotapi.Params{}
	params["url"] = url
	params["secret_token"] = secret
	if _, err := b.api.MakeRequest("setWebhook", params); err != nil {
		// Текст ошибки библиотеки может содержать URL с токеном бота — наружу не отдаём.
		slog.Error("bot: set webhook", "kind", fmt.Sprintf("%T", err))
		return fmt.Errorf("bot: set webhook failed")
	}

	b.webhookDeps = &deps
	slog.Info("bot: webhook mode", "url", url)

	return nil
}

// ProcessUpdateJSON обрабатывает один апдейт, пришедший на webhook-ручку. Принимает сырой
// JSON, чтобы HTTP-слой не зависел от типов tgbotapi. Ошибка — только нечитаемое тело;
// ошибки обработки команд логируются внутри и не доезжают до Telegram (ретраи не помогут).
func (b *Bot) ProcessUpdateJSON(ctx context.Context, body []byte) error {
	if b.disabled {
		return nil
	}
	if b.webhookDeps == nil {
		// Ручка зарегистрирована, а webhook-режим не включён — конфигурационный перекос,
		// апдейт честно игнорируем (поллер его всё равно получит).
		slog.Warn("bot: webhook update ignored: webhook mode is not started")
		return nil
	}

	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		return fmt.Errorf("bot: decode update: %w", err)
	}

	b.processUpdate(ctx, *b.webhookDeps, update)

	return nil
}
