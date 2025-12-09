// Package main demonstrates a complete REST API with MongoDB and Redis caching.
//
// This example shows:
// - MongoDB integration for data persistence
// - Redis caching for performance optimization
// - Service layer pattern with dependency injection
// - Handler pattern for HTTP endpoints
// - Input validation with go-playground/validator
// - Error handling patterns
//
// Prerequisites:
//
//	cd examples && docker compose up -d
//
// Run:
//
//	go run .
//
// Test:
//
//	curl http://localhost:3000/todos
//	curl -X POST http://localhost:3000/todos -H "Content-Type: application/json" -d '{"title":"Learn GOE"}'
//	curl http://localhost:3000/todos/{id}
//	curl -X PUT http://localhost:3000/todos/{id} -H "Content-Type: application/json" -d '{"completed":true}'
//	curl -X DELETE http://localhost:3000/todos/{id}
package main

import (
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
)

func main() {
	_ = goe.New(goe.Options{
		WithHTTP:    true, // Enable HTTP server
		WithMongoDB: true, // Enable MongoDB
		WithCache:   true, // Enable caching (Redis)

		// Providers register services with the DI container
		Providers: []any{
			NewTodoService, // Register our Todo service
			NewTodoHandler, // Register our HTTP handler
		},

		// Invokers are called after all services are ready
		Invokers: []any{
			RegisterRoutes,
		},
	})

	goe.Run()
}

// RegisterRoutes sets up all HTTP routes for the Todo API.
func RegisterRoutes(kernel contract.HTTPKernel, handler *TodoHandler, logger contract.Logger) {
	app := kernel.App()

	// API group for versioning
	api := app.Group("/todos")

	// CRUD endpoints
	api.Get("/", handler.List)
	api.Post("/", handler.Create)
	api.Get("/:id", handler.Get)
	api.Put("/:id", handler.Update)
	api.Delete("/:id", handler.Delete)

	// Stats endpoint (demonstrates caching)
	api.Get("/stats/summary", handler.Stats)

	logger.Info("Todo API routes registered")
}
