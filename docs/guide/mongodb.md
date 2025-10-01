# MongoDB Integration

GOE provides seamless MongoDB integration with a clean, simple contract interface. The MongoDB module handles connection
management and provides direct access to the native MongoDB driver.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Configuration](#configuration)
3. [Basic Usage](#basic-usage)
4. [Multiple Connections](#multiple-connections)
5. [Utility Functions](#utility-functions)
6. [Best Practices](#best-practices)
7. [Examples](#examples)
8. [Troubleshooting](#troubleshooting)

## Quick Start

### Installation

Enable MongoDB support in your GOE application:

```go
package main

import (
	"go.oease.dev/goe/v2"
)

func main() {
	goe.New(goe.Options{
		WithMongoDB: true, // Enable MongoDB integration
	})

	goe.Run()
}
```

### Environment Configuration

Configure MongoDB connection via environment variables:

```env
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=myapp_db
```

### Basic Usage

```go
import (
"context"
"go.mongodb.org/mongo-driver/v2/bson"
"go.oease.dev/goe/v2/contract"
"go.oease.dev/goe/v2/core/mongodb"
)

type UserService struct {
db contract.MongoDB
}

func (s *UserService) CreateUser(name, email string) error {
ctx, cancel := mongodb.DefaultContext()
defer cancel()

// Direct access to MongoDB collection
collection := s.db.Collection("users")

user := bson.M{
"name":  name,
"email": email,
}

_, err := collection.InsertOne(ctx, user)
return err
}
```

## Configuration

### Environment Variables

| Variable              | Required | Default   | Description                                    |
|-----------------------|----------|-----------|------------------------------------------------|
| `MONGO_URI`           | Yes      | -         | MongoDB connection string                      |
| `MONGO_DB_NAME`       | Yes      | -         | Default database name                          |
| `MONGO_CONNECTION`    | No       | `default` | Default connection name                        |
| `MONGO_CONNECTIONS`   | No       | -         | Additional named connections (comma-separated) |
| `MONGO_TIMEOUT`       | No       | `10s`     | Connection timeout                             |
| `MONGO_MIN_POOL_SIZE` | No       | `5`       | Minimum connection pool size                   |
| `MONGO_MAX_POOL_SIZE` | No       | `100`     | Maximum connection pool size                   |

### Single Database Configuration

```env
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=myapp_production
```

### Multiple Connections Configuration

```env
# Default connection
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=myapp_production

# Additional connections
MONGO_CONNECTIONS=analytics,logs

# Analytics connection
MONGO_ANALYTICS_URI=mongodb://analytics.example.com:27017
MONGO_ANALYTICS_DATABASE=analytics_db

# Logs connection  
MONGO_LOGS_URI=mongodb://logs.example.com:27017
MONGO_LOGS_DATABASE=application_logs
```

## Basic Usage

The MongoDB contract provides clean, direct access to MongoDB functionality:

### Contract Interface

```go
type MongoDB interface {
// Client returns the underlying MongoDB client for advanced operations
Client() *mongo.Client

// Instance returns the default database instance
Instance() *mongo.Database

// Connection returns a specific database instance by connection name
Connection(name string) (*mongo.Database, error)

// Collection returns a collection from the default database
Collection(name string) *mongo.Collection

// CollectionFrom returns a collection from a specific database connection
CollectionFrom(connectionName, collectionName string) (*mongo.Collection, error)
}
```

### Basic CRUD Operations

```go
type UserService struct {
db contract.MongoDB
}

func (s *UserService) CreateUser(user *User) error {
ctx, cancel := mongodb.DefaultContext()
defer cancel()

collection := s.db.Collection("users")
_, err := collection.InsertOne(ctx, user)
return err
}

func (s *UserService) GetUser(id string) (*User, error) {
ctx, cancel := mongodb.DefaultContext()
defer cancel()

collection := s.db.Collection("users")
var user User
err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
if err != nil {
return nil, err
}
return &user, nil
}

func (s *UserService) UpdateUser(id string, updates bson.M) error {
ctx, cancel := mongodb.DefaultContext()
defer cancel()

collection := s.db.Collection("users")
_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
return err
}

func (s *UserService) DeleteUser(id string) error {
ctx, cancel := mongodb.DefaultContext()
defer cancel()

collection := s.db.Collection("users")
_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
return err
}
```

### Advanced Operations

```go
func (s *UserService) GetActiveUsers() ([]*User, error) {
ctx, cancel := mongodb.DefaultContext()
defer cancel()

collection := s.db.Collection("users")

// Complex query with options
opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(100)
cursor, err := collection.Find(ctx, bson.M{"active": true}, opts)
if err != nil {
return nil, err
}
defer cursor.Close(ctx)

var users []*User
if err := cursor.All(ctx, &users); err != nil {
return nil, err
}

return users, nil
}

func (s *UserService) GetUserStats() (*UserStats, error) {
ctx, cancel := mongodb.DefaultContext()
defer cancel()

collection := s.db.Collection("users")

// Aggregation pipeline
pipeline := mongo.Pipeline{
{{"$group", bson.D{
{"_id", nil},
{"total", bson.D{{"$sum", 1}}},
{"active", bson.D{{"$sum", bson.D{{"$cond", bson.A{"$active", 1, 0}}}}}},
}}},
}

cursor, err := collection.Aggregate(ctx, pipeline)
if err != nil {
return nil, err
}
defer cursor.Close(ctx)

var results []UserStats
if err := cursor.All(ctx, &results); err != nil {
return nil, err
}

if len(results) == 0 {
return &UserStats{}, nil
}

return &results[0], nil
}
```

## Multiple Connections

Use named connections for different databases:

```go
type AnalyticsService struct {
db contract.MongoDB
}

func (s *AnalyticsService) LogEvent(event *Event) error {
ctx, cancel := mongodb.DefaultContext()
defer cancel()

// Use analytics connection
collection, err := s.db.CollectionFrom("analytics", "events")
if err != nil {
return err
}

_, err = collection.InsertOne(ctx, event)
return err
}

func (s *AnalyticsService) GetDatabase() *mongo.Database {
// Get specific database connection
db, err := s.db.Connection("analytics")
if err != nil {
return nil
}
return db
}
```

## Utility Functions

GOE provides helpful utility functions for common MongoDB operations:

### Context Management

```go
import "go.oease.dev/goe/v2/core/mongodb"

// Create context with custom timeout
ctx, cancel := mongodb.NewContext(30 * time.Second)
defer cancel()

// Create context with default 10-second timeout
ctx, cancel := mongodb.DefaultContext()
defer cancel()
```

### Error Handling

```go
if err := collection.FindOne(ctx, filter).Decode(&result); err != nil {
if mongodb.IsNoDocumentsError(err) {
return nil, ErrUserNotFound
}
if mongodb.IsDuplicateKeyError(err) {
return nil, ErrUserExists
}
if mongodb.IsTimeout(err) {
return nil, ErrTimeout
}
return nil, err
}
```

### Index Management

```go
// Create simple index
indexModel := mongodb.BuildIndexModel(bson.D{{"email", 1}})

// Create unique index
uniqueIndex := mongodb.BuildUniqueIndex(bson.D{{"username", 1}}, "unique_username")

// Create text index for search
textIndex := mongodb.BuildTextIndex([]string{"title", "content"}, "search_index")

// Create compound index
compoundIndex := mongodb.BuildCompoundIndex(map[string]int{
"user_id": 1,
"created_at": -1,
}, "user_timeline")

// Create TTL index
ttlIndex := mongodb.BuildTTLIndex("expires_at", 24*time.Hour, "session_ttl")

// Apply indexes
err := mongodb.EnsureIndexes(ctx, collection, []mongo.IndexModel{
indexModel,
uniqueIndex,
textIndex,
})
```

### Bulk Operations

```go
// Create bulk write models
models := []mongo.WriteModel{
mongodb.CreateInsertOneModel(user1),
mongodb.CreateUpdateOneModel(
bson.M{"_id": user2.ID},
bson.M{"$set": bson.M{"last_login": time.Now()}},
),
mongodb.CreateDeleteOneModel(bson.M{"_id": oldUserID}),
}

// Execute bulk operation
result, err := collection.BulkWrite(ctx, models)
if err != nil {
return err
}

fmt.Printf("Inserted: %d, Modified: %d, Deleted: %d\n",
result.InsertedCount, result.ModifiedCount, result.DeletedCount)
```

### Transactions

```go
func (s *UserService) TransferCredits(fromID, toID string, amount int) error {
client := s.db.Client()
if client == nil {
return errors.New("MongoDB client not available")
}

return mongodb.WithTransaction(ctx, client, func (sessCtx context.Context) error {
users := s.db.Collection("users")

// Deduct from sender
_, err := users.UpdateOne(sessCtx,
bson.M{"_id": fromID},
bson.M{"$inc": bson.M{"credits": -amount}},
)
if err != nil {
return err
}

// Add to receiver
_, err = users.UpdateOne(sessCtx,
bson.M{"_id": toID},
bson.M{"$inc": bson.M{"credits": amount}},
)
return err
})
}
```

## Best Practices

### 1. Use Context with Timeouts

```go
// Always use contexts with timeouts
ctx, cancel := mongodb.DefaultContext() // 10 second timeout
defer cancel()

// Or custom timeout for long-running operations
ctx, cancel := mongodb.NewContext(60 * time.Second)
defer cancel()
```

### 2. Handle Errors Appropriately

```go
if err := collection.FindOne(ctx, filter).Decode(&result); err != nil {
if mongodb.IsNoDocumentsError(err) {
// Handle not found case
return nil, ErrNotFound
}
if mongodb.IsDuplicateKeyError(err) {
// Handle duplicate key
return nil, ErrAlreadyExists
}
// Log and return other errors
logger.Error("MongoDB operation failed", "error", err)
return nil, err
}
```

### 3. Leverage Dependency Injection

```go
// Inject the MongoDB contract
func NewUserService(db contract.MongoDB, logger contract.Logger) *UserService {
return &UserService{
db:     db,
logger: logger,
}
}

// Register with GOE
goe.New(goe.Options{
WithMongoDB: true,
Providers: []any{
NewUserService,
},
})
```

### 4. Use Proper Indexing

```go
func (s *UserService) SetupIndexes(ctx context.Context) error {
collection := s.db.Collection("users")

indexes := []mongo.IndexModel{
mongodb.BuildUniqueIndex(bson.D{{"email", 1}}, "unique_email"),
mongodb.BuildIndexModel(bson.D{{"created_at", -1}}),
mongodb.BuildTextIndex([]string{"name", "bio"}, "user_search"),
}

return mongodb.EnsureIndexes(ctx, collection, indexes)
}
```

## Examples

### Complete User Management Service

```go
package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.oease.dev/goe/v2/contract"
	"go.oease.dev/goe/v2/core/mongodb"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username"`
	Email     string             `bson:"email"`
	Active    bool               `bson:"active"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type UserService struct {
	db     contract.MongoDB
	logger contract.Logger
}

func NewUserService(db contract.MongoDB, logger contract.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

func (s *UserService) Initialize(ctx context.Context) error {
	collection := s.db.Collection("users")

	indexes := []mongo.IndexModel{
		mongodb.BuildUniqueIndex(bson.D{{"email", 1}}, "unique_email"),
		mongodb.BuildUniqueIndex(bson.D{{"username", 1}}, "unique_username"),
		mongodb.BuildIndexModel(bson.D{{"active", 1}, {"created_at", -1}}),
	}

	return mongodb.EnsureIndexes(ctx, collection, indexes)
}

func (s *UserService) Create(user *User) error {
	ctx, cancel := mongodb.DefaultContext()
	defer cancel()

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Active = true

	collection := s.db.Collection("users")
	_, err := collection.InsertOne(ctx, user)

	if mongodb.IsDuplicateKeyError(err) {
		return errors.New("user already exists")
	}

	return err
}

func (s *UserService) GetByID(id string) (*User, error) {
	ctx, cancel := mongodb.DefaultContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	collection := s.db.Collection("users")
	var user User

	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if mongodb.IsNoDocumentsError(err) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) Update(id string, updates bson.M) error {
	ctx, cancel := mongodb.DefaultContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	updates["updated_at"] = time.Now()

	collection := s.db.Collection("users")
	result, err := collection.UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (s *UserService) List(limit, offset int) ([]*User, error) {
	ctx, cancel := mongodb.DefaultContext()
	defer cancel()

	collection := s.db.Collection("users")

	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, bson.M{"active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
```

## Migration Guide

If upgrading from an older version with wrapper classes:

### Before (Old API)

```go
// Old wrapper approach
ops := db.Operations("users")
repo := db.Repository("users").(mongodb.RepositoryFactory).Create[User]()
```

### After (New Simplified API)

```go  
// Direct MongoDB driver access
collection := db.Collection("users")
client := db.Client()
database := db.Instance()
```

The new approach provides:

- ✅ Direct access to MongoDB driver features
- ✅ No wrapper overhead
- ✅ Full MongoDB functionality available
- ✅ Simpler, cleaner API
- ✅ Better performance
- ✅ Easier debugging

For more examples, see the [MongoDB CRUD Example](../examples/mongodb-crud.md).