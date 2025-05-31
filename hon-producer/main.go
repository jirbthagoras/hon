package main

import (
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jirbthagoras/hon/shared"
)

func main() {
	// Creates some dependencies
	validate := validator.New()
	sql := shared.GetConnection()
	amqp := shared.NewAMQPConnection()
	producerService := NewProducerService(sql, amqp)

	// creates a server
	server := fiber.New(fiber.Config{
		ErrorHandler: shared.ErrorHandler,
	})

	producerHandlers := NewProducerHandler(validate, producerService)
	app := server.Group("/api")
	producerHandlers.RegisterRoutes(app)

	if err := server.Listen(":3000"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
