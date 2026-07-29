package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/webresult"
)

// TodoHandler handles HTTP requests for todo operations.
type TodoHandler struct {
	service *TodoService
	logger  contract.Logger
}

// NewTodoHandler creates a new TodoHandler with injected dependencies.
func NewTodoHandler(service *TodoService, logger contract.Logger) *TodoHandler {
	return &TodoHandler{
		service: service,
		logger:  logger,
	}
}

// List returns all todos.
// GET /todos
func (h *TodoHandler) List(c fiber.Ctx) error {
	todos, err := h.service.List(c.Context())
	if err != nil {
		h.logger.Errorw("Failed to list todos", "error", err)
		return webresult.SystemBusy(err)
	}
	return webresult.SendSucceed(c, todos)
}

// Create creates a new todo.
// POST /todos
func (h *TodoHandler) Create(c fiber.Ctx) error {
	var req CreateTodoRequest
	// Bind parses AND validates: it calls the fiber.StructValidator that the GOE
	// HTTP kernel installs, so the `validate` tags on CreateTodoRequest are
	// enforced here. A separate validation call would run them a second time.
	// See https://docs.gofiber.io/guide/validation.
	if err := c.Bind().JSON(&req); err != nil {
		return webresult.SendFailed(c, "Invalid request body")
	}

	todo, err := h.service.Create(c.Context(), &req)
	if err != nil {
		return webresult.SystemBusy(err)
	}

	return webresult.SendSucceed(c, todo)
}

// Get returns a single todo by ID.
// GET /todos/:id
func (h *TodoHandler) Get(c fiber.Ctx) error {
	id := c.Params("id")

	todo, err := h.service.Get(c.Context(), id)
	if err != nil {
		if err.Error() == "todo not found" || err.Error() == "invalid todo ID" {
			return webresult.NotFound(err.Error())
		}
		return webresult.SystemBusy(err)
	}

	return webresult.SendSucceed(c, todo)
}

// Update updates a todo by ID.
// PUT /todos/:id
func (h *TodoHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateTodoRequest
	if err := c.Bind().JSON(&req); err != nil {
		return webresult.SendFailed(c, "Invalid request body")
	}

	todo, err := h.service.Update(c.Context(), id, &req)
	if err != nil {
		if err.Error() == "todo not found" || err.Error() == "invalid todo ID" {
			return webresult.NotFound(err.Error())
		}
		return webresult.SystemBusy(err)
	}

	return webresult.SendSucceed(c, todo)
}

// Delete removes a todo by ID.
// DELETE /todos/:id
func (h *TodoHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")

	err := h.service.Delete(c.Context(), id)
	if err != nil {
		if err.Error() == "todo not found" || err.Error() == "invalid todo ID" {
			return webresult.NotFound(err.Error())
		}
		return webresult.SystemBusy(err)
	}

	return webresult.SendSucceed(c, fiber.Map{"message": "Todo deleted successfully"})
}

// Stats returns aggregate statistics about todos.
// GET /todos/stats/summary
func (h *TodoHandler) Stats(c fiber.Ctx) error {
	stats, err := h.service.Stats(c.Context())
	if err != nil {
		return webresult.SystemBusy(err)
	}
	return webresult.SendSucceed(c, stats)
}
