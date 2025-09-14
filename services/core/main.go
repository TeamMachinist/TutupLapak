package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/teammachinist/tutuplapak/internal"
	"github.com/teammachinist/tutuplapak/internal/cache"
	"github.com/teammachinist/tutuplapak/internal/logger"
	"github.com/teammachinist/tutuplapak/services/core/clients"
	"github.com/teammachinist/tutuplapak/services/core/config"
	"github.com/teammachinist/tutuplapak/services/core/handlers"
	"github.com/teammachinist/tutuplapak/services/core/repositories"
	"github.com/teammachinist/tutuplapak/services/core/services"

	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger.Init()

	database, err := internal.NewDatabase(ctx, cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	redisClient := cache.NewRedisCache(cache.CacheConfig(cfg.Redis))
	defer redisClient.Close()

	var enablePrefork bool
	if cfg.App.Env == "production" {
		enablePrefork = true
	}

	app := fiber.New(fiber.Config{
		Prefork: enablePrefork,
		AppName: "Core Service v1.0",
	})

	app.Use(fiberlog.New())

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "core",
		})
	})

	jwtConfig := &internal.JWTConfig{
		Key:      cfg.JWT.Secret,
		Duration: cfg.JWT.Duration,
		Issuer:   cfg.JWT.Issuer,
	}
	jwtService := internal.NewJWTService(jwtConfig)

	fileClient := clients.NewFileClient(cfg.App.FileUrl)

	productRepo := repositories.NewProductRepository(database.Queries)
	purchaseRepo := repositories.NewPurchaseRepository(database.Pool)

	productService := services.NewProductService(productRepo, fileClient, redisClient)
	purchaseService := services.NewPurchaseService(purchaseRepo, fileClient)

	productHandler := handlers.NewProductHandler(productService)
	purchaseHandler := handlers.NewPurchaseHandler(purchaseService)

	api := app.Group("/api/v1")

	products := api.Group("/products")
	{
		products.Get("", productHandler.GetAllProducts)
		products.Post("", jwtService.FiberMiddleware(), productHandler.CreateProduct)
		products.Put("/:productId", jwtService.FiberMiddleware(), productHandler.UpdateProduct)
		products.Delete("/:productId", jwtService.FiberMiddleware(), productHandler.DeleteProduct)
	}
	purchase := api.Group("/purchase")
	{
		purchase.Post("", purchaseHandler.CreatePurchase)
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
