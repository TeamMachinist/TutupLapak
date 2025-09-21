package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	HTTPPort       string `env:"PORT" envDefault:"8001"`
	DatabaseURL    string `env:"DATABASE_URL" envDefault:""`
	CoreServiceURL string `env:"CORE_SERVICE_URL" envDefault:""`

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

	jwtConfig := &service.JWTConfig{
		Key:      cfg.JWTSecret,
		Duration: cfg.JWTDuration,
		Issuer:   cfg.JWTIssuer,
	}

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
	userService := service.NewUserService(userRepo, db.Queries, jwtConfig, cfg.CoreServiceURL, redisCache)
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler(db, redisCache)
	internalHandler := handler.NewInternalHandler(userService)

	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Health check endpoints
	router.GET("/healthz", healthHandler.HealthCheck)
	router.GET("/readyz", healthHandler.ReadinessCheck)

	// Authentication endpoints
	// v1 := router.Group("/api/v1")
	v1 := router.Group("/v1")
	{
		v1.POST("/login/phone", userHandler.LoginByPhone)
		v1.POST("/register/phone", userHandler.RegisterByPhone)
		v1.POST("/login/email", userHandler.LoginWithEmail)
		v1.POST("/register/email", userHandler.RegisterWithEmail)
	}

	internalHandler.RegisterInternalRoutes(router)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Auth service starting", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start server", "error", err.Error())
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("Shutdown signal received", "signal", sig.String())

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Shutting down server...")

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err.Error())
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Auth service stopped gracefully")
}
