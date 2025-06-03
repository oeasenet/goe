package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	main2 "go.oease.dev/goe/v2/example"
)

// ExampleModule is a module that depends on other modules
type ExampleModule struct {
	name string
}

// NewExampleModule creates a new ExampleModule
func NewExampleModule() *ExampleModule {
	return &ExampleModule{
		name: "example",
	}
}

// Name returns the name of the module
func (m *ExampleModule) Name() string {
	return m.name
}

// Initialize initializes the module
func (m *ExampleModule) Initialize(ctx context.Context) error {
	return nil
}

// Start starts the module
func (m *ExampleModule) Start(ctx context.Context) error {
	return nil
}

// Stop stops the module
func (m *ExampleModule) Stop(ctx context.Context) error {
	return nil
}

// ExampleService is a service that depends on other modules
type ExampleService struct {
	deps *main2.Module // Inject all dependencies through the DI module
}

// NewExampleService creates a new ExampleService with injected dependencies
func NewExampleService(deps *main2.Module) *ExampleService {
	return &ExampleService{
		deps: deps,
	}
}

// Run runs the service
func (s *ExampleService) Run() {
	// Access the config module
	appEnv := s.deps.Config.GetDefault("APP_ENV", "development")
	fmt.Println("App environment:", appEnv)

	// Access the log module
	s.deps.Log.Info("ExampleService is running", "env", appEnv)

	// Access other modules as needed
	s.deps.Http.Get("/example", func(c fiber.Ctx) error {
		return c.JSON(map[string]interface{}{
			"message": "Example service is running",
			"env":     appEnv,
		})
	})
}

func main() {
	// Create a new Goe application
	app := goe.New()

	// Register the DI module provider
	main2.RegisterDI(app.App())

	// Register our example module
	app.App().RegisterModule(NewExampleModule())

	// Register our example service provider
	app.App().RegisterProvider(NewExampleService)

	// Invoke our service after the application has started
	app.App().Invoke(func(service *ExampleService) {
		service.Run()
	})

	// Run the application
	if err := app.Run(); err != nil {
		app.Log().Fatal("Application failed", "error", err)
	}
}
