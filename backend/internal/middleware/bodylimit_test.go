package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"icaris-logistic/backend/internal/middleware"
)

// bindRouter — роутер с лимитом 64 байта и ручкой, которая биндит JSON (как реальные хендлеры).
func bindRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/x", middleware.MaxBodyBytes(64), func(c *gin.Context) {
		var body map[string]any
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		c.Status(http.StatusOK)
	})

	return r
}

func TestMaxBodyBytesRejectsOversizedBody(t *testing.T) {
	rec := httptest.NewRecorder()
	big := `{"comment":"` + strings.Repeat("a", 200) + `"}`

	bindRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(big)))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversized body, got %d", rec.Code)
	}
}

func TestMaxBodyBytesPassesSmallBody(t *testing.T) {
	rec := httptest.NewRecorder()

	bindRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"a":1}`)))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for small body, got %d", rec.Code)
	}
}
