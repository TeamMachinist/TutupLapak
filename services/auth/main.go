package auth

import (
	"log"
	"net/http"
	"os"

	"https://github.com/TeamMachinist/TutupLapak/internal"
	"https://github.com/TeamMachinist/TutupLapak/services/auth"

	"github.com/gin-gonic/gin"
)

type Config struct {
	HTTPPort    string `env:"HTTP_PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL"`

	// JWT Configuration
	JWTSecret   string        `env:"JWT_SECRET" envDefault:"your-secret-key"`
	JWTDuration time.Duration `env:"JWT_DURATION" envDefault:"24h"`
	JWTIssuer   string        `env:"JWT_ISSUER" envDefault:"fitbyte-app"`

	// Redis Configuration
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// MinIO Configuration
	MinIOEndpoint       string `env:"MINIO_ENDPOINT" envDefault:"minio:9000"`
	MinIOAccessKey      string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
	MinIOSecretKey      string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
	MinIOBucket         string `env:"MINIO_BUCKET" envDefault:"fitbyte-uploads"`
	MinIOPublicEndpoint string `env:"MINIO_PUBLIC_ENDPOINT" envDefault:"http://localhost:9000"`
	MinIOUseSSL         bool   `env:"MINIO_USE_SSL" envDefault:"false"`
}

func main() {
	// Load config
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// TODO: Connect to database
	// db, err := database.Connect(cfg.DatabaseURL)
	// if err != nil {
	// 	log.Fatal("Failed to connect to database:", err)
	// }
	// defer db.Close()

	// TODO: Initialize Redis cache
	// redisClient := redis.NewClient(&redis.Options{
	// 	Addr:     cfg.RedisAddr,
	// 	Password: cfg.RedisPassword,
	// 	DB:       cfg.RedisDB,
	// })

	// defer func() {
	// 	if err := redisClient.Close(); err != nil {
	// 		log.Printf("Error closing Redis connection: %v", err)
	// 	}
	// }()

	// cache, err := cache.NewRedis(cache.RedisConfig{DB: redisClient})
	// if err != nil {
	// 	log.Fatal("Failed to initialize Redis cache:", err)
	// }

	// Initialize JWT service
	jwtConfig := &service.SecurityConfig{
		Key:    cfg.JWTSecret,
		Durasi: cfg.JWTDuration,
		Issues: cfg.JWTIssuer,
	}
	jwtService := service.NewJwtService(jwtConfig)

	// Initialize health handler
	healthHandler := handler.NewHealthHandler(db, cache)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup Gin router
	r := gin.Default()

	// Routes
	http.HandleFunc("/healthz", healthHandler)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", userHandler.CreateNewUser)
		v1.POST("/login", userHandler.Login)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Auth service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give 30 seconds for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"auth"}`))
}
