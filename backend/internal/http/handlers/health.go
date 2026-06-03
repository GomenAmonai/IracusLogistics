package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) HealthHandler {
	return HealthHandler{db: db}
}

func (h HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	databaseStatus := "ok"
	if err := h.db.Ping(ctx); err != nil {
		databaseStatus = "unavailable"
	}

	status := http.StatusOK
	if databaseStatus != "ok" {
		status = http.StatusServiceUnavailable
	}

	apiStatus := "ok"
	if databaseStatus != "ok" {
		apiStatus = "degraded"
	}

	writeJSON(w, status, map[string]string{
		"status":   apiStatus,
		"database": databaseStatus,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
