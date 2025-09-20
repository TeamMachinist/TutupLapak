package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teammachinist/tutuplapak/internal/cache"
)

type HealthCheck struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Services map[string]string `json:"services"`
}

type HealthHandler struct {
	db    *pgxpool.Pool
	cache *cache.RedisCache
}

func NewHealthHandler(db *pgxpool.Pool, cache *cache.RedisCache) *HealthHandler {
	return &HealthHandler{
		db:    db,
		cache: cache,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := HealthCheck{
		Status:   "healthy",
		Version:  "1.0.0",
		Services: make(map[string]string),
	}

	if err := h.checkDatabase(ctx); err != nil {
		health.Services["database"] = "down"
		health.Status = "degraded"
	} else {
		health.Services["database"] = "ok"
	}

	if err := h.checkCache(ctx); err != nil {
		health.Services["cache"] = "down"
		if health.Status != "degraded" {
			health.Status = "degraded"
		}
	} else {
		health.Services["cache"] = "ok"
	}

	statusCode := http.StatusOK
	if health.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.checkDatabase(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

func (h *HealthHandler) checkDatabase(ctx context.Context) error {
	if h.db == nil {
		return nil
	}
	return h.db.Ping(ctx)
}

func (h *HealthHandler) checkCache(ctx context.Context) error {
	if h.cache == nil {
		return nil
	}
	return h.cache.Ping(ctx)
}
