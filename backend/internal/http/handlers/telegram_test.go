package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeUpdateProcessor struct {
	body []byte
	err  error
}

func (f *fakeUpdateProcessor) ProcessUpdateJSON(_ context.Context, body []byte) error {
	f.body = body
	return f.err
}

func performWebhookRequest(handler TelegramWebhookHandler, secret, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(recorder)
	engine.POST("/webhook", handler.Handle)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
	if secret != "" {
		req.Header.Set("X-Telegram-Bot-Api-Secret-Token", secret)
	}
	engine.ServeHTTP(recorder, req)

	return recorder
}

func TestTelegramWebhookRejectsWrongSecret(t *testing.T) {
	handler := NewTelegramWebhookHandler(&fakeUpdateProcessor{}, "expected")

	recorder := performWebhookRequest(handler, "wrong", `{}`)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong secret, got %d", recorder.Code)
	}
}

func TestTelegramWebhookRejectsMissingSecret(t *testing.T) {
	handler := NewTelegramWebhookHandler(&fakeUpdateProcessor{}, "expected")

	recorder := performWebhookRequest(handler, "", `{}`)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing secret, got %d", recorder.Code)
	}
}

func TestTelegramWebhookPassesBodyToProcessor(t *testing.T) {
	processor := &fakeUpdateProcessor{}
	handler := NewTelegramWebhookHandler(processor, "s3cret")

	recorder := performWebhookRequest(handler, "s3cret", `{"update_id":1}`)

	if recorder.Code != http.StatusNoContent || string(processor.body) != `{"update_id":1}` {
		t.Errorf("expected 204 with body passed through, got %d, body %q", recorder.Code, processor.body)
	}
}

func TestTelegramWebhookRespondsNoContentOnProcessorError(t *testing.T) {
	processor := &fakeUpdateProcessor{err: errors.New("decode failed")}
	handler := NewTelegramWebhookHandler(processor, "s3cret")

	recorder := performWebhookRequest(handler, "s3cret", `not-json`)

	// 2xx даже на нечитаемый апдейт: Telegram ретраит любой не-2xx, а повтор того же
	// тела валиднее не станет.
	if recorder.Code != http.StatusNoContent {
		t.Errorf("expected 204 on processor error, got %d", recorder.Code)
	}
}
