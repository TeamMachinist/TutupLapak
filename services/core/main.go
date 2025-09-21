package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/teammachinist/tutuplapak/services/auth/pkg/authz"
	"github.com/teammachinist/tutuplapak/services/core/internal/cache"
	"github.com/teammachinist/tutuplapak/services/core/internal/clients"
	"github.com/teammachinist/tutuplapak/services/core/internal/config"
	"github.com/teammachinist/tutuplapak/services/core/internal/database"
	"github.com/teammachinist/tutuplapak/services/core/internal/handler"
	"github.com/teammachinist/tutuplapak/services/core/internal/logger"
	"github.com/teammachinist/tutuplapak/services/core/internal/repository"
	"github.com/teammachinist/tutuplapak/services/core/internal/service"

	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	ctx := context.Background()

	// Initialize logger
	logger.Init()
	logger.Info("Starting Core service")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger.Init()

	database, err := database.NewDatabase(ctx, cfg.Database.DatabaseURL)
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
		Network: "tcp",
	})

	app.Use(fiberlog.New())
	app.Use(requestid.New())

	// app.Get("/healthz", func(c *fiber.Ctx) error {
	// 	return c.JSON(fiber.Map{
	// 		"status":  "healthy",
	// 		"service": "core",
	// 	})
	// })

	// Initiate auth middleware and client
	authClient := authz.NewAuthClient(cfg.App.AuthServiceURL)
	authMiddleware := authz.NewAuthMiddleware(cfg.App.AuthServiceURL)

	fileClient := clients.NewFileClient(cfg.App.FileUrl)

	productRepo := repository.NewProductRepository(database.Queries)
	purchaseRepo := repository.NewPurchaseRepository(database.Pool, database.Queries)
	userRepo := repository.NewUserRepository(database.Queries)

	productService := service.NewProductService(productRepo, fileClient, redisClient)
	purchaseService := service.NewPurchaseService(purchaseRepo, productRepo, fileClient)
	userService := service.NewUserService(userRepo, fileClient, redisClient, authClient)

	productHandler := handler.NewProductHandler(productService)
	purchaseHandler := handler.NewPurchaseHandler(purchaseService)
	userHandler := handler.NewUserHandler(userService)

	healthHandler := handler.NewHealthHandler(database, redisClient)
	internalHandler := handler.NewInternalHandler(userService)

	app.Get("/healthz", healthHandler.HealthCheck)
	app.Get("/readyz", healthHandler.ReadinessCheck)

	api := app.Group("/api/v1")

	products := api.Group("/product")
	{
		products.Get("", productHandler.GetAllProducts)
		products.Post("", authMiddleware.FiberMiddleware(), productHandler.CreateProduct)
		products.Put("/:productId", authMiddleware.FiberMiddleware(), productHandler.UpdateProduct)
		products.Delete("/:productId", authMiddleware.FiberMiddleware(), productHandler.DeleteProduct)

	}

	// User management endpoints (auth-protected)
	user := api.Group("/user")
	{
		user.Post("/link/phone", authMiddleware.FiberMiddleware(), userHandler.LinkPhone)
		user.Post("/link/email", authMiddleware.FiberMiddleware(), userHandler.LinkEmail)
		user.Get("", authMiddleware.FiberMiddleware(), userHandler.GetUserWithFileId)
		user.Put("", authMiddleware.FiberMiddleware(), userHandler.UpdateUser)
	}

	purchase := api.Group("/purchase")
	{
		purchase.Post("", purchaseHandler.CreatePurchase)
		purchase.Post("/:purchaseId", purchaseHandler.UploadPaymentProof)
	}

	internal := app.Group("/internal")
	{
		internal.Get("/user/:userAuthID", internalHandler.GetUserFromAuth)
		internal.Post("/user", internalHandler.CreateUserFromAuth)
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
