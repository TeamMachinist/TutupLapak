package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/services/auth/internal/cache"
	"github.com/teammachinist/tutuplapak/services/auth/internal/database"
	"github.com/teammachinist/tutuplapak/services/auth/internal/handler"
	"github.com/teammachinist/tutuplapak/services/auth/internal/logger"
	"github.com/teammachinist/tutuplapak/services/auth/internal/repository"
	"github.com/teammachinist/tutuplapak/services/auth/internal/service"

	"github.com/caarlos0/env/v8"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	HTTPPort    string `env:"PORT" envDefault:"8001"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:""`

	// JWT Configuration
	JWTSecret   string        `env:"JWT_SECRET" envDefault:"your-secret-key"`
	JWTDuration time.Duration `env:"JWT_DURATION" envDefault:"24h"`
	JWTIssuer   string        `env:"JWT_ISSUER" envDefault:"fitbyte-app"`

	// Redis Configuration
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"redis:6378"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`
}

func main() {
	// Initialize logger
	logger.Init()
	logger.Info("Starting Auth service")

	// Load config
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	ctx := context.Background()
	db, err := database.NewDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("Database connection failed: %v", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// TODO: Use authz package
	// Initialize JWT service
	jwtConfig := &internal.JWTConfig{
		Key:      cfg.JWTSecret,
		Duration: cfg.JWTDuration,
		Issuer:   cfg.JWTIssuer,
	}
	jwtService := internal.NewJWTService(jwtConfig)

	// Initialize Redis cache
	redisConfig := cache.CacheConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	redisCache := cache.NewRedisCache(redisConfig)
	defer func() {
		if err := redisCache.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	// Test Redis connection (non-blocking)
	if err := redisCache.Ping(ctx); err != nil {
		log.Printf("Redis connection failed - running without cache: %v", err)
	}

	// Initialize layers
	userRepo := repository.NewUserRepository(db.Queries)
	userService := service.NewUserService(userRepo, *jwtService, db.Queries)
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler(db.Pool, redisCache)

	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Health check endpoints
	router.GET("/healthz", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/live", healthHandler.LivenessCheck)

	// Authentication endpoints
	v1 := router.Group("/api/v1")
	v1.POST("/login/phone", userHandler.LoginByPhone)
	v1.POST("/register/phone", userHandler.RegisterByPhone)
	v1.POST("/login/email", userHandler.LoginWithEmail)
	v1.POST("/register/email", userHandler.RegisterWithEmail)

	// Simple token generation endpoint for testing
	v1.POST("/token", func(c *gin.Context) {
		var req struct {
			UserID string `json:"user_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		token, err := jwtService.GenerateToken(req.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	log.Printf("Auth service starting on port %s", cfg.HTTPPort)
	log.Fatal(router.Run(":" + cfg.HTTPPort))
}
