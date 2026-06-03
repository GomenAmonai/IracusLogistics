package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) HealthHandler {
	return HealthHandler{db: db}
}

func (h HealthHandler) Handle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	databaseStatus := "ok"
	if err := pingDB(ctx, h.db); err != nil {
		databaseStatus = "unavailable"
	}

	status := http.StatusOK
	apiStatus := "ok"
	if databaseStatus != "ok" {
		status = http.StatusServiceUnavailable
		apiStatus = "degraded"
	}

	c.JSON(status, gin.H{
		"status":   apiStatus,
		"database": databaseStatus,
	})
}

func pingDB(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}
