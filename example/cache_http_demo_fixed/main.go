package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/cache"
	"go.oease.dev/goe/v2/core/http"
	"go.uber.org/fx"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	// Create environment file for demo
	createEnvFile()

	app := goe.New(goe.Options{
		WithHTTP:  true,
		WithCache: true,
		Providers: []any{
			// Provide route registration as a service
			fx.Annotate(
				NewRoutes,
				fx.As(new(Routes)),
			),
		},
		Invokers: []any{
			// Register routes during application setup, before HTTP server starts
			func(routes Routes, httpKernel contract.HTTPKernel) {
				routes.Register(httpKernel.App())
			},
		},
	})

	app.Container().Run()
}

// Routes interface for route registration
type Routes interface {
	Register(app *fiber.App)
}

// routes struct holds dependencies for route handlers
type routes struct {
	cacheManager contract.CacheManager
	logger       contract.Logger
}

// NewRoutes creates a new routes instance with dependencies
func NewRoutes(cacheManager contract.CacheManager, logger contract.Logger) Routes {
	return &routes{
		cacheManager: cacheManager,
		logger:       logger,
	}
}

// Register registers all routes
func (r *routes) Register(app *fiber.App) {
	r.logger.Info("Registering cache HTTP demo routes")

	// Product routes
	api := app.Group("/api")
	api.Get("/products/:id", r.getProductHandler())
	api.Post("/products", r.createProductHandler())
	api.Delete("/cache", r.clearCacheHandler())
	api.Get("/stats", r.getCacheStatsHandler())

	// Home route
	app.Get("/", func(c fiber.Ctx) error {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Cache HTTP Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { background: #f4f4f4; padding: 10px; margin: 10px 0; border-radius: 5px; }
        code { background: #e0e0e0; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>Cache HTTP Demo</h1>
    <h2>Available Endpoints:</h2>
    
    <div class="endpoint">
        <h3>GET /api/products/:id</h3>
        <p>Get a product by ID (cached for 5 minutes)</p>
        <p>Example: <code>curl http://localhost:8080/api/products/1</code></p>
    </div>
    
    <div class="endpoint">
        <h3>POST /api/products</h3>
        <p>Create a new product (clears product cache)</p>
        <p>Example: <code>curl -X POST http://localhost:8080/api/products -H "Content-Type: application/json" -d '{"id":1,"name":"Laptop","price":999.99}'</code></p>
    </div>
    
    <div class="endpoint">
        <h3>DELETE /api/cache</h3>
        <p>Clear all cache entries</p>
        <p>Example: <code>curl -X DELETE http://localhost:8080/api/cache</code></p>
    </div>
    
    <div class="endpoint">
        <h3>GET /api/stats</h3>
        <p>Get cache statistics</p>
        <p>Example: <code>curl http://localhost:8080/api/stats</code></p>
    </div>
</body>
</html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	r.logger.Info("Routes registered successfully")
}

func (r *routes) getProductHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		logger := http.GetLogger(c)
		store := r.cacheManager.Store()

		productID := c.Params("id")
		cacheKey := fmt.Sprintf("product:%s", productID)

		// Try to get from cache using Remember pattern
		product, err := cache.RememberT(store, cacheKey, 5*time.Minute, func() (Product, error) {
			logger.Info("Cache miss - fetching product from database",
				http.NewField("product_id", productID))

			// Simulate database fetch
			time.Sleep(100 * time.Millisecond)

			// Mock product data
			return Product{
				ID:          1,
				Name:        "Gaming Laptop",
				Description: "High-performance laptop for gaming and development",
				Price:       1299.99,
				Stock:       15,
			}, nil
		})

		if err != nil {
			logger.Error("Failed to get product", http.NewField("error", err))
			return c.Status(500).JSON(APIResponse{
				Success: false,
				Error:   "Failed to retrieve product",
			})
		}

		// Track view count
		viewKey := fmt.Sprintf("product:%s:views", productID)
		views, _ := store.Increment(viewKey)

		return c.JSON(APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"product": product,
				"views":   views,
				"cached":  true,
			},
		})
	}
}

func (r *routes) createProductHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		logger := http.GetLogger(c)
		store := r.cacheManager.Store()

		var product Product
		if err := c.Bind().Body(&product); err != nil {
			return c.Status(400).JSON(APIResponse{
				Success: false,
				Error:   "Invalid product data",
			})
		}

		// Simulate saving to database
		logger.Info("Creating new product", http.NewField("product", product))

		// Invalidate product cache
		cacheKey := fmt.Sprintf("product:%d", product.ID)
		err := store.Forget(cacheKey)
		if err != nil {
			logger.Warn("Failed to invalidate cache",
				http.NewField("key", cacheKey),
				http.NewField("error", err))
		}

		// Cache the new product
		err = store.Set(cacheKey, product, 5*time.Minute)
		if err != nil {
			logger.Warn("Failed to cache new product", http.NewField("error", err))
		}

		return c.Status(201).JSON(APIResponse{
			Success: true,
			Data:    product,
		})
	}
}

func (r *routes) clearCacheHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		logger := http.GetLogger(c)
		store := r.cacheManager.Store()

		err := store.Flush()
		if err != nil {
			logger.Error("Failed to clear cache", http.NewField("error", err))
			return c.Status(500).JSON(APIResponse{
				Success: false,
				Error:   "Failed to clear cache",
			})
		}

		logger.Info("Cache cleared successfully")

		return c.JSON(APIResponse{
			Success: true,
			Data:    "Cache cleared successfully",
		})
	}
}

func (r *routes) getCacheStatsHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		store := r.cacheManager.Store()

		// Get some sample keys to show what's cached
		stats := map[string]interface{}{
			"driver":      r.cacheManager.Driver(),
			"prefix":      store.GetPrefix(),
			"sample_keys": []string{},
		}

		// Check some common keys
		sampleKeys := []string{"product:1", "product:1:views", "product:2", "product:2:views"}
		cachedKeys := []string{}

		for _, key := range sampleKeys {
			if store.Has(key) {
				cachedKeys = append(cachedKeys, key)
			}
		}

		stats["cached_keys"] = cachedKeys
		stats["cached_count"] = len(cachedKeys)

		return c.JSON(APIResponse{
			Success: true,
			Data:    stats,
		})
	}
}

func createEnvFile() {
	envContent := `# Application Configuration
APP_NAME=CacheHTTPDemo
APP_VERSION=1.0.0

# HTTP Configuration
FIBER_HOST=localhost
FIBER_PORT=8080
FIBER_PREFORK=false

# Cache Configuration using Fiber drivers
CACHE_STORE=default
CACHE_DRIVER=memory
CACHE_PREFIX=http_demo
CACHE_MEMORY_GC_INTERVAL=10s

# Alternative: Use Redis for better performance
# CACHE_DRIVER=redis
# CACHE_REDIS_URL=redis://localhost:6379/0
# CACHE_REDIS_CACHE_SIZE=134217728
# CACHE_REDIS_ALWAYS_PIPELINING=true

# Logging
LOG_LEVEL=info
LOG_FORMAT=text
`
	_ = os.WriteFile(".env", []byte(envContent), 0644)
	defer os.Remove(".env")
}