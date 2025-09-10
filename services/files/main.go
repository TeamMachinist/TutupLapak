package main

import (
	"log"
	"net/http"
	"os"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct{
	// MinIO Configuration
	MinIOEndpoint       string `env:"MINIO_ENDPOINT" envDefault:"minio:9000"`
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
	// Add miniO
	// Add Repo, Service, and Handler
	// Add upload image function
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	fileHandler := NewFileHandler(minioStorage)
	
	r := chi.NewRouter()

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/", func (w http.ResponseWriter, r *http.Request)  {
		w.Write([]byte("hi"))
	})

	r.Post("/file", fileHandler.UploadFile)
	r.Get("/healthz", healthHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}
	log.Printf("Files service starting on port %s", port)
	
	
	
	log.Fatal(http.ListenAndServe(":"+port, r))
}
