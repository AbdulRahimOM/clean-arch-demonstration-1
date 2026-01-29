package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"myapp/internal/application/usecases"
	"myapp/internal/infrastructure/persistence"
	"myapp/internal/infrastructure/services"
	"myapp/internal/interfaces/http"
)

func main() {
	// 1. Setup MongoDB
	mongoClient := connectMongoDB()
	defer mongoClient.Disconnect(context.Background())

	// 2. Setup Infrastructure Layer
	uow := persistence.NewMongoUnitOfWork(mongoClient, "inventory_db")
	notificationSvc := services.NewNotificationService(
		"https://hooks.slack.com/...",
		"https://email-service.com/api",
	)

	// 3. Setup Application Layer
	addStockUseCase := usecases.NewAddStockUseCase(uow, notificationSvc, nil)

	// 4. Setup HTTP Layer
	stockHandler := http.NewStockHandler(addStockUseCase)

	// 5. Setup Fiber App
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(logger.New())
	app.Use(authMiddleware) // Assume this exists

	// 6. Routes
	app.Post("/api/v1/stock/add", stockHandler.AddStock)

	// 7. Start server
	log.Fatal(app.Listen(":3000"))
}

func connectMongoDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")
	return client
}

func authMiddleware(c *fiber.Ctx) error {
	// Simple auth middleware
	// In real app, validate JWT, etc.
	c.Locals("user_id", "user_123")
	return c.Next()
}