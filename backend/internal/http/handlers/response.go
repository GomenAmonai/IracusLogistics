package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"icaris-logistic/backend/internal/domain"
)

// respondError отдаёт единый формат ошибки: {"error": {"code", "message"}}.
func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

// respondNotFoundOr500 — частый хвост read-ручек: domain.ErrNotFound → 404 с заданным
// сообщением, любая другая ошибка → обобщённый 500 (без утечки первопричины наружу).
func respondNotFoundOr500(c *gin.Context, err error, notFoundMessage string) {
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, "not_found", notFoundMessage)
		return
	}
	respondError(c, http.StatusInternalServerError, "internal", "internal server error")
}
