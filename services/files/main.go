package main

import (
	"log"
	"net/http"
	"os"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"files"}`))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	http.HandleFunc("/healthz", healthHandler)

	log.Printf("Files service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
