package main

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"time"

	"go.oease.dev/goe/v2"
)

func main() {
	// Create a new Goe application
	app := goe.New()

	// Configure the HTTP server
	app.Http().Get("/", func(ctx fiber.Ctx) error {
		return ctx.JSON(goe.Map{
			"message": "Hello, world!",
			"env":     app.Config().GetDefault("APP_ENV", "development"),
		})
	})

	// Configure the logger
	app.Log().Debug("This is a debug message")
	app.Log().Info("Starting application")
	app.Log().Warn("This is a warning message")
	app.Log().Error("This is an error message")

	// Log with fields
	app.Log().Info("User logged in", "user_id", 123, "username", "john")

	// Create a named logger
	userLogger := app.Log().Named("user")
	userLogger.Info("User action")

	// Set a value in the cache
	err := app.Cache().Set(context.Background(), "greeting", "Hello from cache")
	if err != nil {
		app.Log().Fatal("Failed to set cache value", "error", err)
	}

	// Get a value from the cache
	greeting, err := app.Cache().GetString(context.Background(), "greeting")
	if err != nil {
		app.Log().Fatal("Failed to get cache value", "error", err)
	}
	app.Log().Info("Greeting from cache", "greeting", greeting)

	// Publish an event
	app.Event().Publish(context.Background(), "app.started", map[string]interface{}{
		"time": "now",
	})

	// Subscribe to an event
	app.Event().Subscribe("app.event", func(ctx context.Context, payload interface{}) error {
		app.Log().Info("Event received", "payload", payload)
		return nil
	})

	// Run the application with a timeout
	if err := app.RunWithTimeout(10 * time.Second); err != nil {
		app.Log().Fatal("Application failed", "error", err)
	}
}
