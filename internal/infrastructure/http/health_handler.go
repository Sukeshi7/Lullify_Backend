package http

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	infraredis "Lullify_Backend/internal/infrastructure/redis"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *infraredis.Client
}

func NewHealthHandler(db *pgxpool.Pool, redis *infraredis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.Liveness)
	mux.HandleFunc("GET /readyz", h.Readiness)
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{fieldStatus: "ok"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{}
	ready := true

	if err := h.db.Ping(ctx); err != nil {
		checks["postgres"] = "down"
		ready = false
	} else {
		checks["postgres"] = "up"
	}

	if err := h.redis.Ping(ctx); err != nil {
		checks["redis"] = "down"
		ready = false
	} else {
		checks["redis"] = "up"
	}

	status := http.StatusOK
	overall := "ready"
	if !ready {
		status = http.StatusServiceUnavailable
		overall = "not ready"
	}

	writeJSON(w, status, map[string]any{
		fieldStatus: overall,
		"checks":    checks,
	})
}
