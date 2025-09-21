package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/teammachinist/tutuplapak/services/files/internal/api"
	"github.com/teammachinist/tutuplapak/services/files/internal/cache"
	"github.com/teammachinist/tutuplapak/services/files/internal/database"
	"github.com/teammachinist/tutuplapak/services/files/internal/logger"
)

type HealthHandler struct {
	db    *database.DB
	cache *cache.RedisCache
}

func NewHealthHandler(db *database.DB, cache *cache.RedisCache) *HealthHandler {
	return &HealthHandler{
		db:    db,
		cache: cache,
	}
}

// /healthz - Always OK (liveness)
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "files",
	}
	api.WriteSuccess(w, r, response)
}

// /readyz - Only fail if critical deps down (readiness)
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Database is critical - use the comprehensive health check
	if h.db != nil {
		if err := h.db.HealthCheck(ctx); err != nil {
			logger.WarnCtx(ctx, "Readiness check failed - database unavailable", "error", err.Error())
			api.WriteError(w, r, http.StatusServiceUnavailable, "database unavailable")
			return
		}
	}

	// Cache check (non-blocking)
	if h.cache != nil {
		if err := h.checkCache(ctx); err != nil {
			logger.WarnCtx(ctx, "Cache unavailable but service still ready", "error", err.Error())
		}
	}

	api.WriteSuccess(w, r, map[string]string{"status": "ready"})
}

func (h *HealthHandler) checkCache(ctx context.Context) error {
	if h.cache == nil {
		return nil // No cache configured
	}
	return h.cache.Ping(ctx)
}
