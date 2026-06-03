package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// contextKey — ключ, под которым в gin.Context лежит id менеджера из токена.
const contextKey = "managerID"

// RequireAuth проверяет Bearer-JWT и кладёт id менеджера в контекст. Без валидного
// токена — 401 и обрыв цепочки обработчиков.
func RequireAuth(secret string) gin.HandlerFunc {
	key := []byte(secret)

	return func(c *gin.Context) {
		raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || raw == "" {
			abortUnauthorized(c)
			return
		}

		var claims jwt.RegisteredClaims
		token, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
			// Принимаем только HMAC: иначе возможна alg-подмена (подсунуть "none" или RS256).
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return key, nil
		},
			// WithValidMethods дублирует проверку алгоритма на уровне парсера; WithExpirationRequired
			// заставляет отвергать корректно подписанный токен без exp (иначе он жил бы вечно).
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !token.Valid {
			abortUnauthorized(c)
			return
		}

		managerID, err := uuid.Parse(claims.Subject)
		if err != nil {
			abortUnauthorized(c)
			return
		}

		c.Set(contextKey, managerID)
		c.Next()
	}
}

// ManagerID достаёт id менеджера, положенный RequireAuth. ok=false, если RequireAuth не
// отработал (ручка вне защищённой группы) — защита от тихого uuid.Nil.
func ManagerID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(contextKey)
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
