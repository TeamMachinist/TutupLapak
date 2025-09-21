package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/teammachinist/tutuplapak/services/auth/internal/cache"
	"github.com/teammachinist/tutuplapak/services/auth/internal/database"
	"github.com/teammachinist/tutuplapak/services/auth/internal/logger"
)

type HealthCheck struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

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
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "auth"})
}

// /readyz - Only fail if critical deps down (readiness)
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Database is critical
	if err := h.db.HealthCheck(ctx); err != nil {
		logger.WarnCtx(ctx, "Readiness check failed - database unavailable", "error", err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database unavailable",
		})
		return
	}

	// Cache check (non-blocking)
	if err := h.checkCache(ctx); err != nil {
		logger.WarnCtx(ctx, "Cache unavailable but service still ready", "error", err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *HealthHandler) checkCache(ctx context.Context) error {
	if h.cache == nil {
		return nil // No cache configured
	}
	return h.cache.Ping(ctx)
}
