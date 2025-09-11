package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/teammachinist/tutuplapak/internal"
)

func main() {
	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres@localhost:5432/tutuplapak?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	// Initialize database
	ctx := context.Background()
	db, err := internal.NewDatabase(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize cache
	cache := NewCacheService(redisURL)
	defer cache.Close()

	// Initialize layers
	userRepo := NewUserRepository(db.Queries)
	userService := NewUserService(userRepo)
	userHandler := NewUserHandler(userService)
	healthHandler := NewHealthHandler(db.Pool, cache)

	router := gin.Default()

	// Health check endpoints
	router.GET("/healthz", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/live", healthHandler.LivenessCheck)

	// Authentication endpoints
	router.POST("/v1/login/phone", userHandler.LoginByPhone)
	router.POST("/v1/register/phone", userHandler.RegisterByPhone)

	// Simple token generation endpoint for testing
	router.POST("/v1/token", func(c *gin.Context) {
		var req struct {
			UserID string `json:"user_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		jwtService := NewJWTService()
		token, err := jwtService.GenerateToken(req.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	log.Printf("Auth service starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
