// Package main demonstrates the GOE event system using Redis Streams.
//
// This example shows:
// - Publishing events to topics
// - Subscribing to events with consumer groups
// - Typed event handlers for automatic payload unmarshaling
// - Dead letter queue handling for failed messages
// - Event system health checks and statistics
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
//	# Publish an order event
//	curl -X POST http://localhost:3000/orders -H "Content-Type: application/json" \
//	  -d '{"customer_id":"cust_123","items":[{"product":"Widget","quantity":2,"price":29.99}]}'
//
//	# Check event system stats
//	curl http://localhost:3000/events/stats
//
//	# Check dead letter queue
//	curl http://localhost:3000/events/dlq/order.created
package main

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/event"
	"go.oease.dev/goe/v2/webresult"
)

func main() {
	_ = goe.New(goe.Options{
		WithHTTP:  true, // Enable HTTP server
		WithEvent: true, // Enable event system (Redis Streams)

		// Providers register services
		Providers: []any{
			NewOrderService,
			NewNotificationWorker,
			NewInventoryWorker,
		},

		// Invokers run after all services are ready
		Invokers: []any{
			RegisterRoutes,
			StartEventConsumers,
		},
	})

	goe.Run()
}

// RegisterRoutes sets up HTTP endpoints for publishing events.
func RegisterRoutes(kernel contract.HTTPKernel, orderService *OrderService, logger contract.Logger) {
	app := kernel.App()

	// Order endpoints (publishers)
	app.Post("/orders", orderService.CreateOrder)

	// Event system monitoring endpoints
	events := app.Group("/events")
	events.Get("/stats", func(c fiber.Ctx) error {
		stats, err := goe.EventManager().Stats(c.Context())
		if err != nil {
			return webresult.SystemBusy(err)
		}
		return webresult.SendSucceed(c, stats)
	})
	events.Get("/health", func(c fiber.Ctx) error {
		if err := goe.EventManager().Health(c.Context()); err != nil {
			return webresult.SendFailed(c, "Event system unhealthy: "+err.Error())
		}
		return webresult.SendSucceed(c, fiber.Map{"status": "healthy"})
	})
	events.Get("/dlq/:topic", func(c fiber.Ctx) error {
		topic := c.Params("topic")
		dlq := goe.EventManager().GetDeadLetterQueue()
		messages, err := dlq.GetMessages(c.Context(), topic, 10)
		if err != nil {
			return webresult.SystemBusy(err)
		}
		return webresult.SendSucceed(c, messages)
	})

	logger.Info("Event-driven routes registered")
}

// StartEventConsumers starts all event consumers.
func StartEventConsumers(
	ctx context.Context,
	notificationWorker *NotificationWorker,
	inventoryWorker *InventoryWorker,
	logger contract.Logger,
) {
	eventManager := goe.EventManager()

	// Subscribe notification worker to order events
	// Each consumer group processes messages independently (fan-out pattern)
	if err := eventManager.Subscribe(
		ctx,
		TopicOrders,
		"notification-service",
		notificationWorker.Handler(),
	); err != nil {
		logger.Fatalw("Failed to subscribe notification worker", "error", err)
	}

	// Subscribe inventory worker to order events
	if err := eventManager.Subscribe(
		ctx,
		TopicOrders,
		"inventory-service",
		inventoryWorker.Handler(),
	); err != nil {
		logger.Fatalw("Failed to subscribe inventory worker", "error", err)
	}

	logger.Info("Event consumers started")
}

// Topics for the event system
const (
	TopicOrders = "orders"
)

// Event names
const (
	EventOrderCreated = "order.created"
)

// OrderItem represents an item in an order.
type OrderItem struct {
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// Order represents a customer order.
type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id" validate:"required"`
	Items      []OrderItem `json:"items" validate:"required,min=1"`
	Total      float64     `json:"total"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
}

// OrderCreatedEvent is the event payload when an order is created.
type OrderCreatedEvent struct {
	OrderID    string      `json:"order_id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	Total      float64     `json:"total"`
	CreatedAt  time.Time   `json:"created_at"`
}

// OrderService handles order business logic and event publishing.
type OrderService struct {
	eventManager contract.EventManager
	logger       contract.Logger
}

// NewOrderService creates a new OrderService.
func NewOrderService(eventManager contract.EventManager, logger contract.Logger) *OrderService {
	return &OrderService{
		eventManager: eventManager,
		logger:       logger,
	}
}

// CreateOrder handles order creation and publishes an event.
func (s *OrderService) CreateOrder(c fiber.Ctx) error {
	var order Order
	if err := c.Bind().JSON(&order); err != nil {
		return webresult.SendFailed(c, "Invalid request body")
	}

	// Generate order ID and calculate total
	order.ID = "ord_" + time.Now().Format("20060102150405")
	order.Status = "pending"
	order.CreatedAt = time.Now()

	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.Total = total

	// Create the event payload
	eventPayload := OrderCreatedEvent{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		Items:      order.Items,
		Total:      order.Total,
		CreatedAt:  order.CreatedAt,
	}

	// Publish the event using helper function
	if err := event.PublishJSON(
		c.Context(),
		s.eventManager,
		TopicOrders,
		EventOrderCreated,
		eventPayload,
	); err != nil {
		s.logger.Errorw("Failed to publish order event", "error", err, "order_id", order.ID)
		return webresult.SystemBusy(err)
	}

	s.logger.Infow("Order created and event published",
		"order_id", order.ID,
		"customer_id", order.CustomerID,
		"total", order.Total,
	)

	return webresult.SendSucceed(c, order)
}

// NotificationWorker handles order notification events.
type NotificationWorker struct {
	logger contract.Logger
}

// NewNotificationWorker creates a new NotificationWorker.
func NewNotificationWorker(logger contract.Logger) *NotificationWorker {
	return &NotificationWorker{logger: logger}
}

// Handler returns the event handler for this worker.
func (w *NotificationWorker) Handler() contract.EventHandler {
	// Use typed event handler for automatic payload unmarshaling
	return event.CreateTypedEventHandler(func(ctx context.Context, e contract.Event, payload OrderCreatedEvent) error {
		w.logger.Infow("📧 Sending order confirmation email",
			"event_id", e.ID(),
			"event_name", e.Name(),
			"order_id", payload.OrderID,
			"customer_id", payload.CustomerID,
			"total", payload.Total,
		)

		// Simulate email sending
		// In production, this would call an email service
		time.Sleep(100 * time.Millisecond)

		w.logger.Infow("✅ Order confirmation email sent",
			"order_id", payload.OrderID,
		)

		return nil
	})
}

// InventoryWorker handles inventory updates for orders.
type InventoryWorker struct {
	logger contract.Logger
}

// NewInventoryWorker creates a new InventoryWorker.
func NewInventoryWorker(logger contract.Logger) *InventoryWorker {
	return &InventoryWorker{logger: logger}
}

// Handler returns the event handler for this worker.
func (w *InventoryWorker) Handler() contract.EventHandler {
	return event.CreateTypedEventHandler(func(ctx context.Context, e contract.Event, payload OrderCreatedEvent) error {
		w.logger.Infow("📦 Reserving inventory for order",
			"event_id", e.ID(),
			"order_id", payload.OrderID,
			"items_count", len(payload.Items),
		)

		// Simulate inventory reservation
		for _, item := range payload.Items {
			w.logger.Infow("Reserving item",
				"product", item.Product,
				"quantity", item.Quantity,
			)
		}

		time.Sleep(50 * time.Millisecond)

		w.logger.Infow("✅ Inventory reserved",
			"order_id", payload.OrderID,
		)

		return nil
	})
}
