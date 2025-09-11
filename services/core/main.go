package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/teammachinist/tutuplapak/services/core/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	healthHandler := handlers.NewHealthHandler()

	// Setup Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/healthz", healthHandler.HealthCheck)

	log.Printf("Core service starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
