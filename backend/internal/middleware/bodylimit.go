package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodyBytes ограничивает размер тела запроса: чтение сверх limit возвращает ошибку,
// и ShouldBindJSON отвечает 400. Без лимита Gin читает всё тело в память — гигантский
// POST с нескольких IP обходит per-IP rate-limit и ведёт к OOM (техдолг #24).
func MaxBodyBytes(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}
