package health

import (
	"context"
	"time"

	"go.oease.dev/goe/v2/contract"
	"gorm.io/gorm"
)

// DatabaseChecker checks the health of a GORM database connection
type DatabaseChecker struct {
	name string
	db   *gorm.DB
}

// NewDatabaseChecker creates a new database health checker
func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{
		name: "database",
		db:   db,
	}
}

// Name returns the checker name
func (c *DatabaseChecker) Name() string {
	return c.name
}

// Check performs the database health check
func (c *DatabaseChecker) Check(ctx context.Context) contract.HealthCheckResult {
	start := time.Now()

	sqlDB, err := c.db.DB()
	if err != nil {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusDown,
			Latency:   time.Since(start),
			Message:   "failed to get underlying DB: " + err.Error(),
			Timestamp: time.Now(),
		}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusDown,
			Latency:   time.Since(start),
			Message:   "ping failed: " + err.Error(),
			Timestamp: time.Now(),
		}
	}

	return contract.HealthCheckResult{
		Status:    contract.HealthStatusUp,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

// CacheChecker checks the health of a cache store
type CacheChecker struct {
	name  string
	cache contract.Cache
}

// NewCacheChecker creates a new cache health checker
func NewCacheChecker(cache contract.Cache) *CacheChecker {
	return &CacheChecker{
		name:  "cache",
		cache: cache,
	}
}

// Name returns the checker name
func (c *CacheChecker) Name() string {
	return c.name
}

// Check performs the cache health check
func (c *CacheChecker) Check(ctx context.Context) contract.HealthCheckResult {
	start := time.Now()

	// Try to set and get a test value
	testKey := "__health_check__"
	testValue := time.Now().UnixNano()

	if err := c.cache.Set(testKey, testValue, time.Minute); err != nil {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusDown,
			Latency:   time.Since(start),
			Message:   "set failed: " + err.Error(),
			Timestamp: time.Now(),
		}
	}

	var retrieved int64
	if err := c.cache.Get(testKey, &retrieved); err != nil {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusDown,
			Latency:   time.Since(start),
			Message:   "get failed: " + err.Error(),
			Timestamp: time.Now(),
		}
	}

	// Clean up
	_ = c.cache.Forget(testKey)

	return contract.HealthCheckResult{
		Status:    contract.HealthStatusUp,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

// MongoDBChecker checks the health of a MongoDB connection
type MongoDBChecker struct {
	name    string
	mongodb contract.MongoDB
}

// NewMongoDBChecker creates a new MongoDB health checker
func NewMongoDBChecker(mongodb contract.MongoDB) *MongoDBChecker {
	return &MongoDBChecker{
		name:    "mongodb",
		mongodb: mongodb,
	}
}

// Name returns the checker name
func (c *MongoDBChecker) Name() string {
	return c.name
}

// Check performs the MongoDB health check
func (c *MongoDBChecker) Check(ctx context.Context) contract.HealthCheckResult {
	start := time.Now()

	client := c.mongodb.Client()
	if client == nil {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusDown,
			Latency:   time.Since(start),
			Message:   "mongodb client is nil",
			Timestamp: time.Now(),
		}
	}

	if err := client.Ping(ctx, nil); err != nil {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusDown,
			Latency:   time.Since(start),
			Message:   "ping failed: " + err.Error(),
			Timestamp: time.Now(),
		}
	}

	return contract.HealthCheckResult{
		Status:    contract.HealthStatusUp,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

// EventChecker checks the health of the event system
type EventChecker struct {
	name  string
	event contract.EventManager
}

// NewEventChecker creates a new event system health checker
func NewEventChecker(event contract.EventManager) *EventChecker {
	return &EventChecker{
		name:  "event",
		event: event,
	}
}

// Name returns the checker name
func (c *EventChecker) Name() string {
	return c.name
}

// Check performs the event system health check
func (c *EventChecker) Check(ctx context.Context) contract.HealthCheckResult {
	start := time.Now()

	if err := c.event.Health(ctx); err != nil {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusDown,
			Latency:   time.Since(start),
			Message:   "health check failed: " + err.Error(),
			Timestamp: time.Now(),
		}
	}

	return contract.HealthCheckResult{
		Status:    contract.HealthStatusUp,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

// JobChecker checks the health of the job system
type JobChecker struct {
	name string
	job  contract.JobManager
}

// NewJobChecker creates a new job system health checker
func NewJobChecker(job contract.JobManager) *JobChecker {
	return &JobChecker{
		name: "job",
		job:  job,
	}
}

// Name returns the checker name
func (c *JobChecker) Name() string {
	return c.name
}

// Check performs the job system health check
func (c *JobChecker) Check(ctx context.Context) contract.HealthCheckResult {
	start := time.Now()

	if err := c.job.Health(ctx); err != nil {
		return contract.HealthCheckResult{
			Status:    contract.HealthStatusDown,
			Latency:   time.Since(start),
			Message:   "health check failed: " + err.Error(),
			Timestamp: time.Now(),
		}
	}

	return contract.HealthCheckResult{
		Status:    contract.HealthStatusUp,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

// CustomChecker is a helper for creating custom health checkers
type CustomChecker struct {
	name      string
	checkFunc func(ctx context.Context) contract.HealthCheckResult
}

// NewCustomChecker creates a new custom health checker
func NewCustomChecker(name string, check func(ctx context.Context) contract.HealthCheckResult) *CustomChecker {
	return &CustomChecker{
		name:      name,
		checkFunc: check,
	}
}

// Name returns the checker name
func (c *CustomChecker) Name() string {
	return c.name
}

// Check performs the custom health check
func (c *CustomChecker) Check(ctx context.Context) contract.HealthCheckResult {
	return c.checkFunc(ctx)
}
