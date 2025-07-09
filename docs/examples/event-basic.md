# Basic Event System Example

This example demonstrates the basic usage of the GOE event system with Redis Streams.

## Prerequisites

- Redis server running on localhost:6379
- Go 1.21 or later

## Example Application

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/event"
    "github.com/gofiber/fiber/v3"
)

// User represents a user in our system
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Plan  string `json:"plan"`
}

// UserSignupEvent represents a user signup event
type UserSignupEvent struct {
    UserID    int       `json:"userId"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Plan      string    `json:"plan"`
    Timestamp time.Time `json:"timestamp"`
}

// EmailWelcomeEvent represents a welcome email event
type EmailWelcomeEvent struct {
    UserID int    `json:"userId"`
    Email  string `json:"email"`
    Name   string `json:"name"`
}

// UserController handles user-related HTTP endpoints
type UserController struct {
    publisher contract.EventPublisher
}

func NewUserController(publisher contract.EventPublisher) *UserController {
    return &UserController{publisher: publisher}
}

// CreateUser handles user creation requests
func (uc *UserController) CreateUser(c fiber.Ctx) error {
    var user User
    if err := c.Bind().Body(&user); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // Simulate user creation logic
    user.ID = int(time.Now().UnixNano() % 10000)
    
    // Create and publish signup event
    signupEvent := event.NewEvent("user.signup", UserSignupEvent{
        UserID:    user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Plan:      user.Plan,
        Timestamp: time.Now(),
    })
    
    // Add some headers for filtering/routing
    signupEvent.WithHeaders(map[string]string{
        "source":      "web-api",
        "user_plan":   user.Plan,
        "environment": "production",
    })

    if err := uc.publisher.Publish(c.Context(), "user-events", signupEvent); err != nil {
        log.Printf("Failed to publish signup event: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
    }

    return c.JSON(fiber.Map{
        "message": "User created successfully",
        "user":    user,
    })
}

// EmailService handles email-related functionality
type EmailService struct {
    publisher contract.EventPublisher
}

func NewEmailService(publisher contract.EventPublisher) *EmailService {
    return &EmailService{publisher: publisher}
}

// HandleUserSignup processes user signup events
func (es *EmailService) HandleUserSignup(ctx context.Context, event contract.Event) error {
    var payload UserSignupEvent
    if err := event.UnmarshalEventPayload(event, &payload); err != nil {
        return err
    }

    log.Printf("Processing user signup: User ID=%d, Email=%s, Plan=%s", 
        payload.UserID, payload.Email, payload.Plan)

    // Send welcome email based on plan
    welcomeEvent := event.NewEvent("email.welcome", EmailWelcomeEvent{
        UserID: payload.UserID,
        Email:  payload.Email,
        Name:   payload.Name,
    })

    // Add plan-specific headers
    welcomeEvent.WithHeaders(map[string]string{
        "email_type": "welcome",
        "user_plan":  payload.Plan,
        "priority":   getPriorityForPlan(payload.Plan),
    })

    // Publish welcome email event
    if err := es.publisher.Publish(ctx, "email-events", welcomeEvent); err != nil {
        log.Printf("Failed to publish welcome email event: %v", err)
        return err
    }

    log.Printf("Welcome email event published for user %d", payload.UserID)
    return nil
}

// AnalyticsService handles analytics events
type AnalyticsService struct{}

func NewAnalyticsService() *AnalyticsService {
    return &AnalyticsService{}
}

// HandleUserSignup processes user signup events for analytics
func (as *AnalyticsService) HandleUserSignup(ctx context.Context, event contract.Event) error {
    var payload UserSignupEvent
    if err := event.UnmarshalEventPayload(event, &payload); err != nil {
        return err
    }

    log.Printf("Recording analytics: User %d signed up with plan %s", 
        payload.UserID, payload.Plan)

    // Simulate analytics recording
    // In a real app, this would send to analytics service
    return nil
}

// EmailProcessor handles email sending
type EmailProcessor struct{}

func NewEmailProcessor() *EmailProcessor {
    return &EmailProcessor{}
}

// HandleWelcomeEmail processes welcome email events
func (ep *EmailProcessor) HandleWelcomeEmail(ctx context.Context, event contract.Event) error {
    var payload EmailWelcomeEvent
    if err := event.UnmarshalEventPayload(event, &payload); err != nil {
        return err
    }

    headers := event.Headers()
    priority := headers["priority"]
    
    log.Printf("Sending welcome email to %s (User ID: %d) with priority: %s", 
        payload.Email, payload.UserID, priority)

    // Simulate email sending
    time.Sleep(100 * time.Millisecond)
    
    log.Printf("Welcome email sent successfully to %s", payload.Email)
    return nil
}

// Helper function to determine email priority based on plan
func getPriorityForPlan(plan string) string {
    switch plan {
    case "enterprise":
        return "high"
    case "premium":
        return "medium"
    default:
        return "low"
    }
}

// SetupEventHandlers configures event subscriptions
func SetupEventHandlers(consumer contract.EventConsumer, 
    emailService *EmailService, 
    analyticsService *AnalyticsService,
    emailProcessor *EmailProcessor) {
    
    // Subscribe email service to user events
    emailHandler := event.CreateEventHandler(emailService.HandleUserSignup)
    if err := consumer.Subscribe(context.Background(), "user-events", "email-service", emailHandler); err != nil {
        log.Fatalf("Failed to subscribe email service: %v", err)
    }

    // Subscribe analytics service to user events
    analyticsHandler := event.CreateEventHandler(analyticsService.HandleUserSignup)
    if err := consumer.Subscribe(context.Background(), "user-events", "analytics-service", analyticsHandler); err != nil {
        log.Fatalf("Failed to subscribe analytics service: %v", err)
    }

    // Subscribe email processor to email events
    emailProcessorHandler := event.CreateEventHandler(emailProcessor.HandleWelcomeEmail)
    if err := consumer.Subscribe(context.Background(), "email-events", "email-processor", emailProcessorHandler); err != nil {
        log.Fatalf("Failed to subscribe email processor: %v", err)
    }

    log.Println("Event handlers configured successfully")
}

// Routes sets up HTTP routes
func Routes(app *fiber.App, userController *UserController) {
    api := app.Group("/api/v1")
    
    api.Post("/users", userController.CreateUser)
    
    // Health check endpoint
    api.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "healthy"})
    })
}

func main() {
    app := goe.New(goe.Options{
        WithHTTP:  true,
        WithEvent: true,
    })

    // Register services
    app.Provide(NewUserController)
    app.Provide(NewEmailService)
    app.Provide(NewAnalyticsService)
    app.Provide(NewEmailProcessor)

    // Register route setup
    app.Invoke(Routes)
    app.Invoke(SetupEventHandlers)

    log.Println("Starting application with event system...")
    app.Run()
}
```

## Configuration

Create a `.env` file in your project root:

```bash
# HTTP Configuration
HTTP_PORT=8080

# Event System Configuration
EVENT_REDIS_ADDR=localhost:6379
EVENT_REDIS_PASSWORD=
EVENT_REDIS_DB=0
EVENT_CONSUMER_TIMEOUT=30s
EVENT_MAX_RETRIES=3
EVENT_RETRY_BACKOFF=1s
EVENT_DELAYED_QUEUE_ENABLED=true
```

## Running the Example

1. **Start Redis:**
   ```bash
   docker run -d -p 6379:6379 redis:7-alpine
   ```

2. **Run the application:**
   ```bash
   go run main.go
   ```

3. **Test the endpoints:**
   ```bash
   # Create a user
   curl -X POST http://localhost:8080/api/v1/users \
     -H "Content-Type: application/json" \
     -d '{
       "name": "John Doe",
       "email": "john@example.com",
       "plan": "premium"
     }'
   ```

4. **Check the logs:**
   You should see output like:
   ```
   Processing user signup: User ID=1234, Email=john@example.com, Plan=premium
   Welcome email event published for user 1234
   Recording analytics: User 1234 signed up with plan premium
   Sending welcome email to john@example.com (User ID: 1234) with priority: medium
   Welcome email sent successfully to john@example.com
   ```

## Testing with Multiple Consumers

To test the consumer group functionality, you can run multiple instances of the application:

```bash
# Terminal 1
go run main.go

# Terminal 2 (different port)
HTTP_PORT=8081 go run main.go

# Terminal 3 (different port)
HTTP_PORT=8082 go run main.go
```

Send multiple requests and observe how the events are distributed among the consumers.

## Monitoring

### Redis CLI Commands

```bash
# Check streams
redis-cli XINFO STREAM event:stream:user-events
redis-cli XINFO STREAM event:stream:email-events

# Check consumer groups
redis-cli XINFO GROUPS event:stream:user-events
redis-cli XINFO GROUPS event:stream:email-events

# Check pending messages
redis-cli XPENDING event:stream:user-events email-service
redis-cli XPENDING event:stream:user-events analytics-service
```

### Application Metrics

Add this endpoint to monitor application health:

```go
// Add to Routes function
api.Get("/metrics", func(c fiber.Ctx) error {
    eventManager := c.Locals("eventManager").(contract.EventManager)
    
    stats, err := eventManager.Stats(c.Context())
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.JSON(stats)
})
```

## Error Handling Example

```go
// ErrorHandler demonstrates error handling patterns
type ErrorHandler struct{}

func (eh *ErrorHandler) HandleEvent(ctx context.Context, event contract.Event) error {
    var payload UserSignupEvent
    if err := event.UnmarshalEventPayload(event, &payload); err != nil {
        // Log parsing error but don't retry
        log.Printf("Failed to parse event %s: %v", event.ID(), err)
        return nil // Return nil to acknowledge and skip
    }

    // Simulate transient error
    if payload.UserID%2 == 0 {
        return fmt.Errorf("transient error for user %d", payload.UserID)
    }

    // Process successfully
    log.Printf("Processed user %d successfully", payload.UserID)
    return nil
}
```

This example demonstrates:

- **Event Publishing**: Creating and publishing events with payloads and headers
- **Event Consumption**: Setting up consumers with different consumer groups
- **Fan-out Pattern**: Multiple services processing the same events
- **Event Chaining**: One event triggering another event
- **Error Handling**: Proper error handling in event handlers
- **Monitoring**: Basic monitoring and health checks

The event system provides reliable, scalable event processing with automatic retries, dead letter queues, and consumer group management.