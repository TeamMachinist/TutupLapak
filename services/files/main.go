package main

import (
	"log"
	"net/http"
	"os"

	"services/files/handlers"
	"services/files/storage"

	"github.com/go-chi/chi/v5"
)

type Config struct{
	// MinIO Configuration
	MinIOEndpoint       string `env:"MINIO_ENDPOINT" envDefault:"minio:9000"`
	MinIOBucket         string `env:"MINIO_BUCKET" envDefault:"fitbyte-uploads"`
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

	// Initialize MinIO storage
	minioConfig := &storage.MinIOConfig{
		Endpoint:       cfg.MinIOEndpoint,
		BucketName:     cfg.MinIOBucket,
		PublicEndpoint: cfg.MinIOPublicEndpoint,
		UseSSL:         cfg.MinIOUseSSL,
	}

	minioStorage, err := storage.NewMinIOStorage(minioConfig)
	if err != nil {
		log.Fatal("Failed to initialize MinIO storage:", err)
	}

	fileHandler := handlers.NewFileHandler(minioStorage)
	
	r := chi.NewRouter()

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/", func (w http.ResponseWriter, r *http.Request)  {
		w.Write([]byte("hi"))
	})

	r.Post("/file", fileHandler.UploadFile)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}
	log.Printf("Files service starting on port %s", port)
	
	
	http.HandleFunc("/healthz", healthHandler)
	
	log.Fatal(http.ListenAndServe(":"+port, r))
}
