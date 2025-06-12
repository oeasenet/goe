# 14. Practical Examples & Use Cases 💡

This section provides more comprehensive examples to illustrate how to build applications with Goe, combining various concepts like modules, dependency injection, HTTP handling, and database interaction.

## Example 1: Simple CRUD API - Task Management

Let's build a simple API to manage tasks. Tasks will have an ID, Title, and a Done status. We'll use an in-memory store for simplicity in this example, but you could easily swap it out for a GORM-based repository.

### 1. Project Structure (Simplified)

```
goe-tasks-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── task/
│   │   ├── model.go      # Task struct
│   │   ├── store.go      # In-memory store for tasks
│   │   ├── service.go    # Business logic for tasks
│   │   └── handler.go    # HTTP handlers for task routes
│   └── common/
│       └── httperrors/
│           └── errors.go # Common HTTP error helpers (optional)
├── go.mod
└── .env (optional, for APP_PORT etc.)
```

### 2. Task Model (`internal/task/model.go`)

```go
package task

import "sync"

// Task represents a single task item.
type Task struct {
	ID    string `json:"id"`
	Title string `json:"title" validate:"required,min=3"`
	Done  bool   `json:"done"`
}

// InMemoryTaskStore provides thread-safe in-memory storage for tasks.
// In a real application, this would be a database repository.
type InMemoryTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]Task
	nextID int
}

func NewInMemoryTaskStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		tasks:  make(map[string]Task),
		nextID: 1,
	}
}
```

### 3. Task Store (`internal/task/store.go`)
(Continuing `InMemoryTaskStore` from `model.go` for this example)
```go
package task

import (
	"fmt"
	"strconv"
	// "sync" // Already in model.go if combined
)

// AddTask adds a new task to the store.
func (s *InMemoryTaskStore) AddTask(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	idStr := strconv.Itoa(s.nextID)
	s.nextID++

	task := Task{
		ID:    idStr,
		Title: title,
		Done:  false,
	}
	s.tasks[idStr] = task
	return task
}

// GetTask retrieves a task by its ID.
func (s *InMemoryTaskStore) GetTask(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, exists := s.tasks[id]
	return task, exists
}

// GetAllTasks retrieves all tasks.
func (s *InMemoryTaskStore) GetAllTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	allTasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		allTasks = append(allTasks, task)
	}
	return allTasks
}

// UpdateTask updates an existing task.
func (s *InMemoryTaskStore) UpdateTask(id string, title string, done bool) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.tasks[id]
	if !exists {
		return Task{}, false
	}
	task.Title = title
	task.Done = done
	s.tasks[id] = task
	return task, true
}

// DeleteTask removes a task from the store.
func (s *InMemoryTaskStore) DeleteTask(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.tasks[id]
	if exists {
		delete(s.tasks, id)
		return true
	}
	return false
}
```

### 4. Task Service (`internal/task/service.go`)

```go
package task

import (
	"fmt"
	"go.oease.dev/goe/v2/contract"
)

// TaskService handles the business logic for tasks.
type TaskService struct {
	logger contract.Logger
	store  *InMemoryTaskStore // In a real app, this would be an interface type
}

func NewTaskService(logger contract.Logger, store *InMemoryTaskStore) *TaskService {
	return &TaskService{logger: logger, store: store}
}

func (s *TaskService) CreateTask(title string) Task {
	s.logger.Info("Creating new task", contract.NewField("title", title))
	return s.store.AddTask(title)
}

func (s *TaskService) FindTaskByID(id string) (Task, error) {
	s.logger.Info("Fetching task by ID", contract.NewField("task_id", id))
	task, exists := s.store.GetTask(id)
	if !exists {
		return Task{}, fmt.Errorf("task with ID '%s' not found", id) // Simple error
	}
	return task, nil
}

func (s *TaskService) ListAllTasks() []Task {
	s.logger.Info("Listing all tasks")
	return s.store.GetAllTasks()
}

func (s *TaskService) ModifyTask(id string, title string, done bool) (Task, error) {
	s.logger.Info("Updating task", contract.NewField("task_id", id))
	task, updated := s.store.UpdateTask(id, title, done)
	if !updated {
		return Task{}, fmt.Errorf("task with ID '%s' not found for update", id)
	}
	return task, nil
}

func (s *TaskService) RemoveTask(id string) error {
	s.logger.Info("Deleting task", contract.NewField("task_id", id))
	if !s.store.DeleteTask(id) {
		return fmt.Errorf("task with ID '%s' not found for deletion", id)
	}
	return nil
}
```

### 5. Task HTTP Handler (`internal/task/handler.go`)

```go
package task

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
	goehttp "go.oease.dev/goe/v2/core/http" // For GetValidator, GetLogger
)

// TaskHandler handles HTTP requests for tasks.
type TaskHandler struct {
	logger  contract.Logger
	service *TaskService
}

func NewTaskHandler(logger contract.Logger, service *TaskService) *TaskHandler {
	return &TaskHandler{logger: logger, service: service}
}

// RegisterRoutes registers task routes with the Fiber app.
func (h *TaskHandler) RegisterRoutes(router fiber.Router) {
	taskGroup := router.Group("/tasks")
	taskGroup.Post("/", h.createTask)
	taskGroup.Get("/", h.getAllTasks)
	taskGroup.Get("/:id", h.getTask)
	taskGroup.Put("/:id", h.updateTask)
	taskGroup.Delete("/:id", h.deleteTask)
	h.logger.Info("Task routes registered under /tasks")
}

type CreateTaskRequest struct {
	Title string `json:"title" validate:"required,min=3"`
}

type UpdateTaskRequest struct {
	Title string `json:"title" validate:"required,min=3"`
	Done  bool   `json:"done"`
}

func (h *TaskHandler) createTask(c fiber.Ctx) error {
	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	validator := goehttp.GetValidator(c)
	if err := validator.Validate(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	task := h.service.CreateTask(req.Title)
	return c.Status(fiber.StatusCreated).JSON(task)
}

func (h *TaskHandler) getTask(c fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.service.FindTaskByID(id)
	if err != nil {
		// Simple error check; could use errors.Is for specific error types
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(task)
}

func (h *TaskHandler) getAllTasks(c fiber.Ctx) error {
	tasks := h.service.ListAllTasks()
	return c.JSON(tasks)
}

func (h *TaskHandler) updateTask(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	validator := goehttp.GetValidator(c)
	if err := validator.Validate(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	task, err := h.service.ModifyTask(id, req.Title, req.Done)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(task)
}

func (h *TaskHandler) deleteTask(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.RemoveTask(id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
```

### 6. Main Application (`cmd/server/main.go`)

```go
package main

import (
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.uber.org/fx"

	// Adjust import path to your project structure
	"example.com/goe-tasks-api/internal/task"
)

func main() {
	_ = goe.New(goe.Options{
		WithHTTP: true, // Enable HTTP module
		// WithLog: true by default

		// Provide all necessary components to Fx
		Providers: []any{
			task.NewInMemoryTaskStore, // Provides *task.InMemoryTaskStore
			task.NewTaskService,       // Provides *task.TaskService, depends on Logger and InMemoryTaskStore
			task.NewTaskHandler,       // Provides *task.TaskHandler, depends on Logger and TaskService
		},

		// Invoke a function to register routes
		Invokers: []any{
			func(kernel contract.HTTPKernel, taskHandler *task.TaskHandler, logger contract.Logger) {
				logger.Info("Setting up application routes...")
				taskHandler.RegisterRoutes(kernel.App()) // Register routes on the Fiber app
			},
		},
	})

	goe.Run()
}
```

**To run this example:**
1. Create the directory structure and files as shown.
2. Replace `example.com/goe-tasks-api` with your actual Go module path in `main.go`.
3. Run `go mod tidy` in the `goe-tasks-api` root.
4. Run `go run cmd/server/main.go`.
5. You can then use `curl` or Postman to interact with the API:
    * `POST /tasks` with JSON `{"title": "My First Task"}`
    * `GET /tasks`
    * `GET /tasks/1`
    * `PUT /tasks/1` with JSON `{"title": "Updated Task", "done": true}`
    * `DELETE /tasks/1`

This example demonstrates how to structure a feature (tasks) with its own model, store (in-memory), service, and HTTP handler, all wired together using Goe's Fx-based dependency injection.

## Example 2: Background Worker Module

Let's create a module that performs a simulated background task periodically.

### 1. Background Worker Module (`internal/worker/background_worker.go`)

```go
package worker

import (
	"context"
	"time"
	"go.oease.dev/goe/v2/contract"
)

// BackgroundWorkerModule simulates a module performing periodic tasks.
type BackgroundWorkerModule struct {
	logger   contract.Logger
	ticker   *time.Ticker
	doneChan chan bool
}

func NewBackgroundWorkerModule(logger contract.Logger) *BackgroundWorkerModule {
	return &BackgroundWorkerModule{
		logger:   logger,
		doneChan: make(chan bool),
	}
}

func (m *BackgroundWorkerModule) Name() string {
	return "background_worker"
}

func (m *BackgroundWorkerModule) OnStart(ctx context.Context) error {
	m.logger.Info("BackgroundWorkerModule starting...")
	m.ticker = time.NewTicker(10 * time.Second) // Perform task every 10 seconds

	go func() {
		for {
			select {
			case <-ctx.Done(): // Listen for application shutdown context
				m.logger.Info("Context done, worker stopping periodic task.")
				return
			case <-m.doneChan: // Listen for explicit stop signal from OnStop
				m.logger.Info("DoneChan signaled, worker stopping periodic task.")
				return
			case t := <-m.ticker.C:
				m.performTask(t)
			}
		}
	}()
	return nil
}

func (m *BackgroundWorkerModule) OnStop(ctx context.Context) error {
	m.logger.Info("BackgroundWorkerModule stopping...")
	if m.ticker != nil {
		m.ticker.Stop()
	}
	// Signal the goroutine to stop
	// Use a non-blocking send in case the goroutine already exited via ctx.Done()
	select {
	case m.doneChan <- true:
		m.logger.Info("Sent stop signal to worker goroutine.")
	default:
		m.logger.Info("Worker goroutine likely already stopped.")
	}
	// Add any other cleanup, like waiting for the goroutine with a timeout using ctx
	return nil
}

func (m *BackgroundWorkerModule) performTask(tickTime time.Time) {
	m.logger.Info("Performing background task",
		contract.NewField("tick_time", tickTime.Format(time.RFC3339)),
		contract.NewField("module", m.Name()),
	)
	// Simulate work
	time.Sleep(1 * time.Second)
	m.logger.Info("Background task completed.")
}
```

### 2. Registering the Worker Module (`cmd/server/main.go` or a new `cmd/worker/main.go`)

```go
package main

import (
	"go.oease.dev/goe/v2"
	"go.uber.org/fx"

	// Adjust import path
	"example.com/yourproject/internal/worker"
)

func main() {
	_ = goe.New(goe.Options{
		// No HTTP needed if it's just a worker, but can be combined
		// WithHTTP: true,

		Providers: []any{
			worker.NewBackgroundWorkerModule, // Fx will inject contract.Logger
		},
		// Goe automatically handles contract.Module lifecycle if NewBackgroundWorkerModule returns it
	})

	goe.Run() // This will block and keep the worker running
}
```

**To run this example:**
1. Create the `internal/worker/background_worker.go` file.
2. Create a `main.go` (e.g., `cmd/worker/main.go`) to register and run it.
3. Run `go run cmd/worker/main.go`.
4. You'll see log messages every 10 seconds as the worker performs its task. Press `Ctrl+C` to trigger `OnStop` and graceful shutdown.

These examples should provide a solid starting point for building various types of applications and modules with Goe. Remember to adapt the patterns to your specific needs, especially regarding error handling, data persistence, and configuration for production environments.
```
