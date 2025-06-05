package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http"
	"go.oease.dev/goe/v2/core/log"
)

// TestUser for validation
type TestUser struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=18"`
}

func main() {
	// Set test environment variables
	os.Setenv("APP_NAME", "Feature Test App")
	os.Setenv("APP_VERSION", "2.0.0")
	os.Setenv("FIBER_SERVER_HEADER", "Goe/Test")
	os.Setenv("FIBER_BODY_LIMIT", "1048576") // 1MB

	// Create application
	_ = goe.New(goe.Options{
		WithHTTP: true,
		Invokers: []any{
			func(http contract.HTTPKernel) {
				fiberApp := http.App()

				// Register custom validation
				customValidator := http.Validator().(*goehttp.CustomValidator)
				customValidator.RegisterValidation("customtag", func(fl validator.FieldLevel) bool {
					return fl.Field().String() == "valid"
				})

				// Test configuration endpoint
				fiberApp.Get("/config", func(c fiber.Ctx) error {
					config := goehttp.GetConfig(c)
					return c.JSON(fiber.Map{
						"app_name":      config.GetString("APP_NAME"),
						"app_version":   config.GetString("APP_VERSION"),
						"server_header": config.GetString("FIBER_SERVER_HEADER"),
						"body_limit":    config.GetInt("FIBER_BODY_LIMIT"),
					})
				})

				// Test validation endpoint
				fiberApp.Post("/validate", func(c fiber.Ctx) error {
					validator := goehttp.GetValidator(c)
					logger := goehttp.GetLogger(c)

					var user TestUser
					if err := c.Bind().JSON(&user); err != nil {
						return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
					}

					if err := validator.Validate(user); err != nil {
						logger.Info("Validation failed", log.NewField("error", err.Error()))
						return c.Status(400).JSON(fiber.Map{"error": err.Error()})
					}

					return c.JSON(fiber.Map{"message": "Valid user", "user": user})
				})

				// Test server header
				fiberApp.Get("/headers", func(c fiber.Ctx) error {
					return c.JSON(fiber.Map{
						"server": c.Response().Header.Peek("Server"),
					})
				})
			},
		},
	})

	// Start server in background
	go func() {
		goe.Log().Info("Starting test server",
			log.NewField("port", 8888),
		)
		if err := goe.HTTP().Listen(":8888"); err != nil {
			goe.Log().Error("Server error", log.NewField("error", err.Error()))
		}
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Run tests
	fmt.Println("Testing new features...")

	// Test 1: Configuration
	fmt.Println("\n1. Testing configuration:")
	resp, err := http.Get("http://localhost:8888/config")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Config response: %s\n", body)
	}

	// Test 2: Server header
	fmt.Println("\n2. Testing server header:")
	resp, err = http.Get("http://localhost:8888/headers")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Server header from response: %s\n", resp.Header.Get("Server"))
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Headers response: %s\n", body)
	}

	// Test 3: Validation - valid request
	fmt.Println("\n3. Testing validation (valid):")
	validUser := TestUser{Name: "John Doe", Email: "john@example.com", Age: 25}
	jsonData, _ := json.Marshal(validUser)
	resp, err = http.Post("http://localhost:8888/validate", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Valid request response: %s\n", body)
	}

	// Test 4: Validation - invalid request
	fmt.Println("\n4. Testing validation (invalid):")
	invalidUser := TestUser{Name: "Jo", Email: "invalid", Age: 17}
	jsonData, _ = json.Marshal(invalidUser)
	resp, err = http.Post("http://localhost:8888/validate", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Invalid request response (status %d): %s\n", resp.StatusCode, body)
	}

	// Test 5: Body limit
	fmt.Println("\n5. Testing body limit:")
	largeBody := bytes.Repeat([]byte("a"), 2*1024*1024) // 2MB
	resp, err = http.Post("http://localhost:8888/validate", "application/json", bytes.NewReader(largeBody))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Large body response status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	fmt.Println("\nAll tests completed!")

	// Keep running
	select {}
}
