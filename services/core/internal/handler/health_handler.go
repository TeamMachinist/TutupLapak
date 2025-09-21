package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/teammachinist/tutuplapak/services/core/internal/cache"
	"github.com/teammachinist/tutuplapak/services/core/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/internal/logger"
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

// /healthz - Always OK (liveness probe)
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "core",
	})
}

// /readyz - Database dependency check (readiness probe)
func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Database is critical for Core service
	if err := h.db.HealthCheck(ctx); err != nil {
		logger.WarnCtx(ctx, "Readiness check failed - database unavailable", "error", err.Error())
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not ready",
			"error":  "database unavailable",
		})
	}

	// Cache check (non-blocking - Redis failure doesn't fail readiness)
	if err := h.checkCache(ctx); err != nil {
		logger.WarnCtx(ctx, "Cache unavailable but service still ready", "error", err.Error())
	}

	return c.JSON(fiber.Map{"status": "ready"})
}

func (h *HealthHandler) checkCache(ctx context.Context) error {
	if h.cache == nil {
		return nil // No cache configured
	}
	return h.cache.Ping(ctx)
}
