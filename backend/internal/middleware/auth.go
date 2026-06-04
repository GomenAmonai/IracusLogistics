package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/token"
)

// Ключи, под которыми в gin.Context лежит id субъекта из токена.
const (
	contextKeyManager = "managerID"
	contextKeyClient  = "clientID"
)

// RequireAuth пропускает только менеджерский токен (role=manager) и кладёт id менеджера в
// контекст. Без валидного токена нужной роли — 401 и обрыв цепочки обработчиков.
func RequireAuth(secret string) gin.HandlerFunc {
	return requireRole(secret, domain.RoleManager, contextKeyManager)
}

// RequireClientAuth пропускает только клиентский токен (role=client) и кладёт id клиента в
// контекст. Разделение ролей не даёт клиентскому токену пройти менеджерскую проверку и
// наоборот — иначе валидно подписанный токен одной роли открыл бы ручки другой.
func RequireClientAuth(secret string) gin.HandlerFunc {
	return requireRole(secret, domain.RoleClient, contextKeyClient)
}

func requireRole(secret string, role domain.Role, ctxKey string) gin.HandlerFunc {
	key := []byte(secret)

	return func(c *gin.Context) {
		raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || raw == "" {
			abortUnauthorized(c)
			return
		}

		claims, err := token.Parse(key, raw)
		if err != nil || claims.Role != role {
			abortUnauthorized(c)
			return
		}

		subject, err := uuid.Parse(claims.Subject)
		if err != nil {
			abortUnauthorized(c)
			return
		}

		c.Set(ctxKey, subject)
		c.Next()
	}
}

// ManagerID достаёт id менеджера, положенный RequireAuth. ok=false, если RequireAuth не
// отработал (ручка вне защищённой группы) — защита от тихого uuid.Nil.
func ManagerID(c *gin.Context) (uuid.UUID, bool) {
	return subjectFromContext(c, contextKeyManager)
}

// ClientID достаёт id клиента, положенный RequireClientAuth.
func ClientID(c *gin.Context) (uuid.UUID, bool) {
	return subjectFromContext(c, contextKeyClient)
}

func subjectFromContext(c *gin.Context, ctxKey string) (uuid.UUID, bool) {
	value, exists := c.Get(ctxKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	return id, ok
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{"code": "unauthorized", "message": "authorization required"},
	})
}
