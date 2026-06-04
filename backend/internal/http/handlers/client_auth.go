package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"icaris-logistic/backend/internal/service"
)

type ClientAuthHandler struct {
	service *service.ClientService
}

func NewClientAuthHandler(service *service.ClientService) ClientAuthHandler {
	return ClientAuthHandler{service: service}
}

type telegramAuthRequest struct {
	InitData string `json:"init_data"`
}

// Telegram принимает initData из Telegram WebApp, проверяет подпись и выдаёт client-JWT.
func (h ClientAuthHandler) Telegram(c *gin.Context) {
	var req telegramAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	token, client, err := h.service.AuthenticateWebApp(c.Request.Context(), req.InitData)
	if err != nil {
		if errors.Is(err, service.ErrTelegramAuth) {
			respondError(c, http.StatusUnauthorized, "telegram_auth", "telegram authentication failed")
			return
		}
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "client": client})
}
