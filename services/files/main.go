package main

import (
	"context"
	"net/http"
	"time"

	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/internal/cache"
	"github.com/teammachinist/tutuplapak/internal/logger"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	HTTPPort    string `env:"PORT" envDefault:"8003"`
	DatabaseURL string `env:"DATABASE_URL"`

	// JWT Configuration
	JWTSecret   string        `env:"JWT_SECRET" envDefault:"tutupsecret"`
	JWTDuration time.Duration `env:"JWT_DURATION" envDefault:"24h"`
	JWTIssuer   string        `env:"JWT_ISSUER" envDefault:"tutuplapak-auth"`

	// Redis Configuration
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// MinIO Configuration
	MinIOEndpoint       string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	MinIOAccessKey      string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
	MinIOSecretKey      string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
	MinIOBucket         string `env:"MINIO_BUCKET" envDefault:"tutuplapak-uploads"`
	MinIOPublicEndpoint string `env:"MINIO_PUBLIC_ENDPOINT" envDefault:"http://localhost:9000"`
	MinIOUseSSL         bool   `env:"MINIO_USE_SSL" envDefault:"false"`
}

// Request ID middleware
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.WithRequestID(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func healthHandler(ctx context.Context, db *internal.DB, redisCache *cache.RedisCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := r.Context()
		logger.InfoCtx(requestCtx, "Health check request received", "remote_addr", r.RemoteAddr)

		// Check database
		if err := db.HealthCheck(ctx); err != nil {
			logger.ErrorCtx(requestCtx, "Health check failed - database ping error", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","service":"files","error":"database unavailable"}`))
			return
		}

		// Check Redis
		if err := redisCache.Ping(ctx); err != nil {
			logger.WarnCtx(requestCtx, "Health check warning - Redis ping failed", "error", err)
			// Don't fail health check for Redis - it's not critical for basic functionality
		}

		logger.InfoCtx(requestCtx, "Health check passed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"files","database":"ok","cache":"ok"}`))
	}
}

func main() {
	// Initialize logger first
	logger.Init()
	logger.Info("Starting Files service")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse configuration
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Error("Failed to load config", "error", err)
		panic(err)
	}

	logger.Info("Configuration loaded",
		"port", cfg.HTTPPort,
		"jwt_issuer", cfg.JWTIssuer,
		"minio_endpoint", cfg.MinIOEndpoint,
		"minio_bucket", cfg.MinIOBucket,
	)

	// Connect to database
	db, err := internal.NewDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("Database connection failed", "error", err, "url", cfg.DatabaseURL)
		panic(err)
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	// Initialize Redis cache
	redisCache := cache.NewRedisCache(cache.CacheConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer func() {
		if err := redisCache.Close(); err != nil {
			logger.Error("Failed to close Redis connection", "error", err)
		}
	}()

	// Test Redis connection (non-blocking)
	if err := redisCache.Ping(ctx); err != nil {
		logger.Warn("Redis connection failed - running without cache", "error", err)
	}

	// Initialize JWT service for token validation
	jwtConfig := &internal.JWTConfig{
		Key:      cfg.JWTSecret,
		Duration: cfg.JWTDuration,
		Issuer:   cfg.JWTIssuer,
	}
	jwtService := internal.NewJWTService(jwtConfig)
	logger.Info("JWT service initialized", "issuer", cfg.JWTIssuer, "duration", cfg.JWTDuration)

	// Initialize MinIO storage
	minioConfig := &MinIOConfig{
		Endpoint:       cfg.MinIOEndpoint,
		BucketName:     cfg.MinIOBucket,
		PublicEndpoint: cfg.MinIOPublicEndpoint,
		UseSSL:         cfg.MinIOUseSSL,
		AccessKey:      cfg.MinIOAccessKey,
		SecretKey:      cfg.MinIOSecretKey,
	}

	minioStorage, err := NewMinIOStorage(minioConfig)
	if err != nil {
		logger.Error("Failed to initialize MinIO storage", "error", err, "endpoint", cfg.MinIOEndpoint)
		panic(err)
	}
	logger.Info("MinIO storage initialized", "endpoint", cfg.MinIOEndpoint, "bucket", cfg.MinIOBucket)

	// Initialize services
	fileService := NewFileService(db.Queries, redisCache)
	fileHandler := NewFileHandler(minioStorage, fileService)

	// Setup router with middleware
	r := chi.NewRouter()

	// Add middleware
	r.Use(requestIDMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Add request logging middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.InfoCtx(r.Context(), "Request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
				"user_agent", r.UserAgent(),
			)
		})
	})

	// Static files
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// Health check (no auth required)
	r.Get("/healthz", healthHandler(ctx, db, redisCache))

	// API routes with authentication
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(jwtService.ChiMiddleware)
		r.Post("/file", fileHandler.UploadFile)
		r.Get("/file/{fileId}", fileHandler.GetFile)
		r.Delete("/file/{fileId}", fileHandler.DeleteFile)
	})

	logger.Info("Files service starting", "port", cfg.HTTPPort)

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Error("Server stopped", "error", server.ListenAndServe())
}
