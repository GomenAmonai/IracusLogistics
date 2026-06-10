package handlers

import (
	"context"
	"crypto/subtle"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateProcessor обрабатывает сырой JSON-апдейт Telegram (реализует bot.Bot). Сырые
// байты, а не типы tgbotapi — чтобы HTTP-слой не зависел от телеграм-библиотеки.
type UpdateProcessor interface {
	ProcessUpdateJSON(ctx context.Context, body []byte) error
}

// maxUpdateBody ограничивает тело апдейта: реальные апдейты — килобайты, всё крупнее —
// мусор, который незачем читать в память.
const maxUpdateBody = 1 << 20

// TelegramWebhookHandler — приём апдейтов от Telegram в webhook-режиме. Аутентификация —
// секрет из setWebhook: Telegram возвращает его в каждом запросе заголовком
// X-Telegram-Bot-Api-Secret-Token, посторонний трафик отсекаем сравнением.
type TelegramWebhookHandler struct {
	processor UpdateProcessor
	secret    string
}

func NewTelegramWebhookHandler(processor UpdateProcessor, secret string) TelegramWebhookHandler {
	return TelegramWebhookHandler{processor: processor, secret: secret}
}

func (h TelegramWebhookHandler) Handle(c *gin.Context) {
	header := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
	// Constant-time: не даём подбирать секрет по времени ответа.
	if subtle.ConstantTimeCompare([]byte(header), []byte(h.secret)) != 1 {
		respondError(c, http.StatusUnauthorized, "unauthorized", "invalid webhook secret")
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxUpdateBody))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "cannot read request body")
		return
	}

	if err := h.processor.ProcessUpdateJSON(c.Request.Context(), body); err != nil {
		// Нечитаемый апдейт: 400 без ретраев со стороны Telegram нам и нужен — повтор
		// того же тела не станет валиднее.
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid update payload")
		return
	}

	c.Status(http.StatusNoContent)
}
