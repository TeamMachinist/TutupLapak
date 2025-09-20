package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	DatabaseURL string `env:"DATABASE_URL" envDefault:""`

	// JWT Configuration
	JWTSecret   string        `env:"JWT_SECRET" envDefault:"tutupsecret"`
	JWTDuration time.Duration `env:"JWT_DURATION" envDefault:"24h"`
	JWTIssuer   string        `env:"JWT_ISSUER" envDefault:"tutuplapak-auth"`

	// Redis Configuration
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
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

type Dependencies struct {
	DB         *internal.DB
	RedisCache *cache.RedisCache
	JWTService *internal.JWTService
	MinIO      *MinIOStorage
}

type Services struct {
	FileService FileService
	FileHandler *FileHandler
}

func main() {
	// 1. Initialize logger
	initializeLogger()

	// 2. Load configuration
	cfg := loadConfiguration()

	// 3. Setup dependencies
	deps := setupDependencies(cfg)
	defer cleanupDependencies(deps)

	// 4. Setup services & handlers
	services := setupServices(deps)

	// 5. Setup routes
	router := setupRoutes(services, deps)

	// 6. Start server with graceful shutdown
	startServerWithShutdown(router, cfg)
}

func initializeLogger() {
	logger.Init()
	logger.Info("Starting Files service")
}

func loadConfiguration() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("Configuration loaded",
		"port", cfg.HTTPPort,
		"jwt_issuer", cfg.JWTIssuer,
		"minio_endpoint", cfg.MinIOEndpoint,
		"minio_bucket", cfg.MinIOBucket,
		"redis_addr", cfg.RedisAddr,
	)

	return cfg
}

func setupDependencies(cfg Config) Dependencies {
	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := internal.NewDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("Database connection failed", "error", err, "url", cfg.DatabaseURL)
		os.Exit(1)
	}
	logger.Info("Database connected successfully")

	// Initialize Redis cache
	redisConfig := cache.CacheConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	redisCache := cache.NewRedisCache(redisConfig)

	// Test Redis connection (non-blocking)
	if err := redisCache.Ping(ctx); err != nil {
		logger.Warn("Redis connection failed - running without cache", "error", err)
	}

	// Initialize JWT service
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
		os.Exit(1)
	}
	logger.Info("MinIO storage initialized", "endpoint", cfg.MinIOEndpoint, "bucket", cfg.MinIOBucket)

	return Dependencies{
		DB:         db,
		RedisCache: redisCache,
		JWTService: jwtService,
		MinIO:      minioStorage,
	}
}

func setupServices(deps Dependencies) Services {
	fileService := NewFileService(deps.DB.Queries, deps.RedisCache)
	fileHandler := NewFileHandler(deps.MinIO, fileService)

	return Services{
		FileService: fileService,
		FileHandler: fileHandler,
	}
}

func setupRoutes(services Services, deps Dependencies) *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(requestIDMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(requestLoggerMiddleware)

	// Static files -> What does it serve?
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// Health check (no auth required)
	r.Get("/healthz", healthHandler(deps.DB, deps.RedisCache))

	// Pragmatically no auth for now, to ease fetch by core service (purchase)
	r.Get("/api/v1/file/{fileId}", services.FileHandler.GetFile)
	r.Get("/api/v1/file", services.FileHandler.GetFiles)

	// API routes with authentication
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(deps.JWTService.ChiMiddleware)
		r.Post("/file", services.FileHandler.UploadFile)
		// r.Get("/file/{fileId}", services.FileHandler.GetFile)
		r.Delete("/file/{fileId}", services.FileHandler.DeleteFile)
	})

	return r
}

func startServerWithShutdown(router *chi.Mux, cfg Config) {
	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	serverErrors := make(chan error, 1)

	// Start HTTP server in goroutine
	go func() {
		logger.Info("Files service starting", "port", cfg.HTTPPort, "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either server error or shutdown signal
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}

	case sig := <-shutdown:
		logger.Info("Shutdown signal received", "signal", sig.String())

		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown server gracefully", "error", err)

			// Force close if graceful shutdown fails
			if closeErr := server.Close(); closeErr != nil {
				logger.Error("Failed to force close server", "error", closeErr)
				os.Exit(1)
			}
			os.Exit(1)
		}

		logger.Info("Server shutdown completed gracefully")
	}
}

func cleanupDependencies(deps Dependencies) {
	logger.Info("Cleaning up dependencies")

	if deps.RedisCache != nil {
		if err := deps.RedisCache.Close(); err != nil {
			logger.Error("Failed to close Redis connection", "error", err)
		}
	}

	if deps.DB != nil {
		deps.DB.Close()
		logger.Info("Database connection closed")
	}
}

// Request ID middleware
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.WithRequestID(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Request logging middleware
func requestLoggerMiddleware(next http.Handler) http.Handler {
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
}

// Health handler
func healthHandler(db *internal.DB, redisCache *cache.RedisCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
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

		// Check Redis (non-critical)
		redisOk := true
		if err := redisCache.Ping(ctx); err != nil {
			logger.WarnCtx(requestCtx, "Health check warning - Redis ping failed", "error", err)
			redisOk = false
		}

		logger.InfoCtx(requestCtx, "Health check passed", "redis_ok", redisOk)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		status := "ok"
		if !redisOk {
			status = "degraded"
		}

		w.Write([]byte(`{"status":"healthy","service":"files","database":"ok","cache":"` + status + `"}`))
	}
}
