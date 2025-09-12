package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/teammachinist/tutuplapak/internal"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	HTTPPort    string `env:"HTTP_PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL"`

	// MinIO Configuration
	MinIOEndpoint       string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	MinIOAccessKey      string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
	MinIOSecretKey      string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
	MinIOBucket         string `env:"MINIO_BUCKET" envDefault:"tutuplapak-uploads"`
	MinIOPublicEndpoint string `env:"MINIO_PUBLIC_ENDPOINT" envDefault:"http://localhost:9000"`
	MinIOUseSSL         bool   `env:"MINIO_USE_SSL" envDefault:"false"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"files"}`))
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := internal.NewDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect the database:", err)
	}

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
		log.Fatal("Failed to initialize MinIO storage:", err)
	}

	// fileRepo := NewFileRepository(db.Queries)
	fileService := NewFileService(db.Queries)
	fileHandler := NewFileHandler(minioStorage, fileService)

	r := chi.NewRouter()

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Route("/api/v1/", func(r chi.Router) {
		r.Use(dummyUserMiddleware)
		r.Post("/file", fileHandler.UploadFile)
		r.Get("/file/{fileId}", fileHandler.GetFile)
		r.Delete("/file/{fileId}", fileHandler.DeleteFile)

	})

	r.Get("/healthz", healthHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}
	log.Printf("Files service starting on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
	db.Close()
}

// Temporary middleware for testing (add to main.go if needed)
func dummyUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add dummy user ID to context for testing
		ctx := internal.WithUserID(r.Context(), "22222222-2222-2222-2222-222222222222") // Dummy UUID
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
