package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) HealthHandler {
	return HealthHandler{db: db}
}

func (h HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
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

	writeJSON(w, status, map[string]string{
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
