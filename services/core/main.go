package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/services/core/config"
	"github.com/teammachinist/tutuplapak/services/core/handlers"
	"github.com/teammachinist/tutuplapak/services/core/repositories"
	"github.com/teammachinist/tutuplapak/services/core/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// func healthHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"status":"healthy","service":"core"}`))
// }

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	database, err := internal.NewDatabase(ctx, cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	app := fiber.New(fiber.Config{
		Prefork: true,
		AppName: "Core Service v1.0",
	})

	app.Use(logger.New())

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "core",
		})
	})

	productRepo := repositories.NewProductRepository(database.Queries)

	productService := services.NewProductService(productRepo)

	productHandler := handlers.NewProductHandler(productService)

	api := app.Group("/api/v1")

	products := api.Group("/products")
	{
		products.Get("", productHandler.GetAllProducts)
		products.Post("", productHandler.CreateProduct)
		products.Put("/:productId", productHandler.UpdateProduct)
		products.Delete("/:productId", productHandler.DeleteProduct)

	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		app.Shutdown()
	}()

	log.Printf("Core service starting on port %s", cfg.App.Port)
	if err := app.Listen(":" + cfg.App.Port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
