# MongoDB

GOE provides comprehensive MongoDB integration with support for modern MongoDB features, connection pooling, multiple databases, and full observability.

## Overview

The MongoDB module provides:
- **Native MongoDB Driver**: Uses the official MongoDB Go driver v2
- **Multiple Connections**: Support for multiple named database connections
- **Connection Pooling**: Configurable connection pool settings
- **Command Monitoring**: Built-in logging and custom monitoring support
- **Type-Safe Operations**: Generic repository pattern for type safety
- **Observability**: Integrated metrics and distributed tracing
- **Lifecycle Management**: Automatic connection setup and teardown

## Configuration

Configure MongoDB through environment variables:

```bash
# Default Connection
MONGO_DB_URI=mongodb://localhost:27017
MONGO_DB_DB_NAME=myapp

# Connection Pool (Optional)
MONGO_DB_MIN_POOL_SIZE=5
MONGO_DB_MAX_POOL_SIZE=100
MONGO_DB_MAX_CONN_IDLE_TIME=30s

# Named Default Connection
MONGO_DB_CONNECTION=primary
MONGO_DB_PRIMARY_URI=mongodb://primary.example.com:27017
MONGO_DB_PRIMARY_DB_NAME=myapp

# Multiple Connections
MONGO_DB_CONNECTIONS=analytics,readonly
MONGO_DB_ANALYTICS_URI=mongodb://analytics.example.com:27017
MONGO_DB_ANALYTICS_DB_NAME=analytics
MONGO_DB_READONLY_URI=mongodb://readonly.example.com:27017
MONGO_DB_READONLY_DB_NAME=myapp_readonly
```

### Connection String Options

MongoDB connection strings support various options:

```bash
# Basic authentication
MONGO_DB_URI=mongodb://username:password@localhost:27017

# Replica set
MONGO_DB_URI=mongodb://host1:27017,host2:27017,host3:27017/?replicaSet=myRs

# Connection options
MONGO_DB_URI=mongodb://localhost:27017/?authSource=admin&ssl=true

# Atlas connection
MONGO_DB_URI=mongodb+srv://username:password@cluster.mongodb.net
```

## Enabling MongoDB Module

Enable the MongoDB module in your GOE application:

```go
package main

import "go.oease.dev/goe/v2"

func main() {
    goe.New(goe.Options{
        WithMongoDB: true,
        WithObservability: true, // Optional: for metrics and tracing
    })
    
    goe.Run()
}
```

## Accessing MongoDB

### Using Global Accessor

```go
import (
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/core/mongodb"
)

func someFunction() {
    // Get MongoDB instance
    mongoDB := goe.MongoDB()
    db := mongoDB.Instance()
    
    // Use the database
    collection := db.Collection("users")
    
    // Create a context with timeout
    ctx, cancel := mongodb.DefaultContext()
    defer cancel()
    
    // Perform operations
    result, err := collection.InsertOne(ctx, bson.M{
        "name": "John Doe",
        "email": "john@example.com",
    })
}
```

### Using Dependency Injection

```go
import (
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/mongodb"
)

type UserRepository struct {
    mongodb contract.MongoDB
}

func NewUserRepository(mongodb contract.MongoDB) *UserRepository {
    return &UserRepository{mongodb: mongodb}
}

func (r *UserRepository) CreateUser(user *User) error {
    db := r.mongodb.Instance()
    collection := db.Collection("users")
    
    ctx, cancel := mongodb.DefaultContext()
    defer cancel()
    
    _, err := collection.InsertOne(ctx, user)
    return err
}
```

## Models

Define your MongoDB models:

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name      string            `bson:"name" json:"name"`
    Email     string            `bson:"email" json:"email"`
    Password  string            `bson:"password" json:"-"`
    Active    bool              `bson:"active" json:"active"`
    Tags      []string          `bson:"tags" json:"tags"`
    Profile   UserProfile       `bson:"profile" json:"profile"`
    CreatedAt time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

type UserProfile struct {
    Bio       string            `bson:"bio" json:"bio"`
    AvatarURL string            `bson:"avatar_url" json:"avatar_url"`
    Location  string            `bson:"location" json:"location"`
    Social    map[string]string `bson:"social" json:"social"`
}

type Post struct {
    ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
    Title     string              `bson:"title" json:"title"`
    Content   string              `bson:"content" json:"content"`
    AuthorID  primitive.ObjectID   `bson:"author_id" json:"author_id"`
    Tags      []string            `bson:"tags" json:"tags"`
    Views     int64               `bson:"views" json:"views"`
    Published bool                `bson:"published" json:"published"`
    CreatedAt time.Time           `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time           `bson:"updated_at" json:"updated_at"`
}
```

## CRUD Operations

### Using Direct Collection Access

```go
func (r *UserRepository) Create(user *User) error {
    collection := r.mongodb.Instance().Collection("users")
    
    ctx, cancel := mongodb.DefaultContext()
    defer cancel()
    
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()
    
    result, err := collection.InsertOne(ctx, user)
    if err != nil {
        return err
    }
    
    user.ID = result.InsertedID.(primitive.ObjectID)
    return nil
}

func (r *UserRepository) FindByID(id string) (*User, error) {
    collection := r.mongodb.Instance().Collection("users")
    
    ctx, cancel := mongodb.DefaultContext()
    defer cancel()
    
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    
    var user User
    err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
    if err != nil {
        if mongodb.IsNoDocumentsError(err) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    
    return &user, nil
}

func (r *UserRepository) Update(user *User) error {
    collection := r.mongodb.Instance().Collection("users")
    
    ctx, cancel := mongodb.DefaultContext()
    defer cancel()
    
    user.UpdatedAt = time.Now()
    
    filter := bson.M{"_id": user.ID}
    update := bson.M{"$set": user}
    
    result, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }
    
    if result.MatchedCount == 0 {
        return ErrUserNotFound
    }
    
    return nil
}

func (r *UserRepository) Delete(id string) error {
    collection := r.mongodb.Instance().Collection("users")
    
    ctx, cancel := mongodb.DefaultContext()
    defer cancel()
    
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return err
    }
    
    result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
    if err != nil {
        return err
    }
    
    if result.DeletedCount == 0 {
        return ErrUserNotFound
    }
    
    return nil
}
```

### Using Generic Repository

```go
// Create a typed repository
userRepo := mongodb.NewRepository[User](db, "users")

// Create
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
}

result, err := userRepo.Insert(ctx, user)

// Find by ID
found, err := userRepo.FindByID(ctx, objectID)

// Find with filter
users, err := userRepo.Find(ctx, bson.M{"active": true})

// Update
updateResult, err := userRepo.UpdateByID(ctx, user.ID, bson.M{
    "$set": bson.M{"name": "Jane Doe"},
})

// Delete
deleteResult, err := userRepo.DeleteByID(ctx, user.ID)

// Pagination
page, err := userRepo.Paginate(ctx, bson.M{}, 1, 20)
fmt.Printf("Total: %d, Page: %d/%d\n", page.Total, page.Page, page.TotalPages)
```

## Query Building

### Basic Queries

```go
// Find all active users
cursor, err := collection.Find(ctx, bson.M{"active": true})
defer cursor.Close(ctx)

var users []User
err = cursor.All(ctx, &users)

// Find with options
opts := options.Find().
    SetSort(bson.D{{"created_at", -1}}).
    SetLimit(10).
    SetProjection(bson.D{{"password", 0}})

cursor, err := collection.Find(ctx, bson.M{}, opts)

// Find one with sorting
opts := options.FindOne().SetSort(bson.D{{"created_at", -1}})
var latestUser User
err := collection.FindOne(ctx, bson.M{}, opts).Decode(&latestUser)
```

### Complex Queries

```go
// Text search
filter := bson.M{
    "$text": bson.M{"$search": "mongodb"},
}
cursor, err := collection.Find(ctx, filter)

// Regular expression
filter := bson.M{
    "email": bson.M{"$regex": "@example.com$"},
}

// Array operations
filter := bson.M{
    "tags": bson.M{"$in": []string{"mongodb", "database"}},
}

// Compound conditions
filter := bson.M{
    "$and": []bson.M{
        {"age": bson.M{"$gte": 18}},
        {"age": bson.M{"$lte": 65}},
        {"active": true},
    },
}

// Nested documents
filter := bson.M{
    "profile.location": "New York",
}
```

## Aggregation

```go
// Simple aggregation
pipeline := mongo.Pipeline{
    {{"$match", bson.D{{"active", true}}}},
    {{"$group", bson.D{
        {"_id", "$location"},
        {"count", bson.D{{"$sum", 1}}},
        {"avg_age", bson.D{{"$avg", "$age"}}},
    }}},
    {{"$sort", bson.D{{"count", -1}}}},
}

cursor, err := collection.Aggregate(ctx, pipeline)

// Complex aggregation with lookup
pipeline := mongo.Pipeline{
    {{"$match", bson.D{{"published", true}}}},
    {{"$lookup", bson.D{
        {"from", "users"},
        {"localField", "author_id"},
        {"foreignField", "_id"},
        {"as", "author"},
    }}},
    {{"$unwind", "$author"}},
    {{"$project", bson.D{
        {"title", 1},
        {"content", 1},
        {"author.name", 1},
        {"author.email", 1},
    }}},
}

// Using typed repository
results, err := postRepo.Aggregate(ctx, pipeline)
```

## Indexes

### Creating Indexes

```go
// Simple index
index := mongodb.BuildIndexModel(bson.D{{"email", 1}})
name, err := collection.Indexes().CreateOne(ctx, index)

// Unique index
uniqueIndex := mongodb.BuildUniqueIndex(
    bson.D{{"email", 1}},
    "unique_email",
)

// Compound index
compoundIndex := mongodb.BuildCompoundIndex(
    map[string]int{"category": 1, "created_at": -1},
    "category_created",
)

// Text index for search
textIndex := mongodb.BuildTextIndex(
    []string{"title", "content"},
    "content_search",
)

// TTL index for expiration
ttlIndex := mongodb.BuildTTLIndex(
    "expires_at",
    0, // Documents expire at the time specified in expires_at field
    "document_ttl",
)

// Create multiple indexes
indexes := []mongo.IndexModel{
    uniqueIndex,
    compoundIndex,
    textIndex,
}
err := mongodb.CreateIndexes(ctx, collection, indexes)
```

### Managing Indexes

```go
// List indexes
cursor, err := collection.Indexes().List(ctx)
var indexes []bson.M
err = cursor.All(ctx, &indexes)

// Drop index
_, err := collection.Indexes().DropOne(ctx, "index_name")

// Drop all indexes
_, err := collection.Indexes().DropAll(ctx)
```

## Transactions

```go
// Using the transaction helper
err := mongodb.WithTransaction(ctx, db.Client(), func(sessCtx context.Context) error {
    // All operations in this function are part of the transaction
    
    // Insert user
    usersCollection := db.Collection("users")
    userResult, err := usersCollection.InsertOne(sessCtx, user)
    if err != nil {
        return err // Transaction will be aborted
    }
    
    // Insert related post
    post.AuthorID = userResult.InsertedID.(primitive.ObjectID)
    postsCollection := db.Collection("posts")
    _, err = postsCollection.InsertOne(sessCtx, post)
    if err != nil {
        return err // Transaction will be aborted
    }
    
    return nil // Transaction will be committed
})

// Manual transaction control
session, err := db.Client().StartSession()
if err != nil {
    return err
}
defer session.EndSession(ctx)

err = session.StartTransaction()
if err != nil {
    return err
}

// Perform operations
err = performOperations(session)
if err != nil {
    session.AbortTransaction(ctx)
    return err
}

err = session.CommitTransaction(ctx)
```

## Bulk Operations

```go
// Bulk write
models := []mongo.WriteModel{
    mongo.NewInsertOneModel().SetDocument(bson.M{
        "name": "User 1",
        "email": "user1@example.com",
    }),
    mongo.NewUpdateOneModel().
        SetFilter(bson.M{"email": "user2@example.com"}).
        SetUpdate(bson.M{"$set": bson.M{"active": true}}),
    mongo.NewDeleteOneModel().
        SetFilter(bson.M{"email": "user3@example.com"}),
}

opts := mongodb.BulkWriteOptions()
result, err := collection.BulkWrite(ctx, models, opts)

fmt.Printf("Inserted: %d, Updated: %d, Deleted: %d\n",
    result.InsertedCount,
    result.ModifiedCount,
    result.DeletedCount,
)
```

## Change Streams

```go
// Watch for changes
pipeline := mongo.Pipeline{
    {{"$match", bson.D{
        {"operationType", bson.D{{"$in", []string{"insert", "update"}}}},
    }}},
}

stream, err := collection.Watch(ctx, pipeline)
if err != nil {
    return err
}
defer stream.Close(ctx)

for stream.Next(ctx) {
    var event bson.M
    if err := stream.Decode(&event); err != nil {
        log.Printf("Error decoding event: %v", err)
        continue
    }
    
    log.Printf("Change event: %v", event)
}
```

## Error Handling

```go
// Using MongoDB utility functions
err := collection.FindOne(ctx, filter).Decode(&result)
if err != nil {
    if mongodb.IsNoDocumentsError(err) {
        return nil, ErrNotFound
    }
    if mongodb.IsNetworkError(err) {
        return nil, ErrConnectionFailed
    }
    if mongodb.IsTimeout(err) {
        return nil, ErrTimeout
    }
    return nil, err
}

// Duplicate key errors
_, err = collection.InsertOne(ctx, document)
if err != nil {
    if mongodb.IsDuplicateKeyError(err) {
        return ErrEmailAlreadyExists
    }
    return err
}
```

## Service Pattern

```go
type UserService struct {
    mongodb contract.MongoDB
    logger  contract.Logger
}

func NewUserService(mongodb contract.MongoDB, logger contract.Logger) *UserService {
    return &UserService{
        mongodb: mongodb,
        logger:  logger,
    }
}

func (s *UserService) CreateUser(name, email, password string) (*User, error) {
    s.logger.Info("Creating user", "email", email)
    
    // Check if user exists
    collection := s.mongodb.Instance().Collection("users")
    
    ctx, cancel := mongodb.NewContext(5 * time.Second)
    defer cancel()
    
    exists, err := mongodb.Exists(ctx, collection, bson.M{"email": email})
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrEmailAlreadyExists
    }
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    
    // Create user
    user := &User{
        ID:        primitive.NewObjectID(),
        Name:      name,
        Email:     email,
        Password:  string(hashedPassword),
        Active:    true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    _, err = collection.InsertOne(ctx, user)
    if err != nil {
        s.logger.Error("Failed to create user", "error", err)
        return nil, err
    }
    
    s.logger.Info("User created successfully", "user_id", user.ID.Hex())
    return user, nil
}

func (s *UserService) GetUsersPaginated(page, pageSize int64, filter bson.M) (*mongodb.PaginatedResult[User], error) {
    repo := mongodb.NewRepository[User](s.mongodb.Instance(), "users")
    
    ctx, cancel := mongodb.DefaultContext()
    defer cancel()
    
    // Add sorting
    opts := options.Find().SetSort(bson.D{{"created_at", -1}})
    
    return repo.Paginate(ctx, filter, page, pageSize, opts)
}
```

## HTTP Integration

```go
type UserHandler struct {
    userService *UserService
}

func NewUserHandler(userService *UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUsers(c fiber.Ctx) error {
    page := int64(c.QueryInt("page", 1))
    pageSize := int64(c.QueryInt("page_size", 20))
    
    // Build filter from query params
    filter := bson.M{}
    if active := c.Query("active"); active != "" {
        filter["active"] = active == "true"
    }
    
    result, err := h.userService.GetUsersPaginated(page, pageSize, filter)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch users",
        })
    }
    
    return c.JSON(result)
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
    userID := c.Params("id")
    
    user, err := h.userService.GetUserByID(userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return c.Status(404).JSON(fiber.Map{
                "error": "User not found",
            })
        }
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch user",
        })
    }
    
    return c.JSON(user)
}
```

## Testing

### Unit Testing

```go
func TestUserRepository(t *testing.T) {
    ctx := context.Background()
    
    // Setup test database using testcontainers
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "mongo:6",
            ExposedPorts: []string{"27017/tcp"},
            WaitingFor:   wait.ForListeningPort("27017/tcp"),
        },
        Started: true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    // Get connection string
    host, err := container.Host(ctx)
    require.NoError(t, err)
    
    port, err := container.MappedPort(ctx, "27017")
    require.NoError(t, err)
    
    uri := fmt.Sprintf("mongodb://%s:%s", host, port.Port())
    
    // Connect to MongoDB
    client, err := mongo.Connect(options.Client().ApplyURI(uri))
    require.NoError(t, err)
    defer client.Disconnect(ctx)
    
    db := client.Database("test")
    repo := mongodb.NewRepository[User](db, "users")
    
    // Test operations
    t.Run("Insert", func(t *testing.T) {
        user := &User{
            Name:  "Test User",
            Email: "test@example.com",
        }
        
        result, err := repo.Insert(ctx, user)
        assert.NoError(t, err)
        assert.NotNil(t, result.InsertedID)
    })
    
    t.Run("Find", func(t *testing.T) {
        users, err := repo.Find(ctx, bson.M{})
        assert.NoError(t, err)
        assert.Len(t, users, 1)
        assert.Equal(t, "Test User", users[0].Name)
    })
}
```

### Mock MongoDB

```go
type MockMongoDB struct {
    mock.Mock
}

func (m *MockMongoDB) Instance() *mongo.Database {
    args := m.Called()
    return args.Get(0).(*mongo.Database)
}

func (m *MockMongoDB) Connection(name string) (*mongo.Database, error) {
    args := m.Called(name)
    return args.Get(0).(*mongo.Database), args.Error(1)
}
```

## Best Practices

### 1. Use Context with Timeouts

```go
// Always use contexts with timeouts
ctx, cancel := mongodb.NewContext(5 * time.Second)
defer cancel()

// For long operations
ctx, cancel := mongodb.NewContext(30 * time.Second)
defer cancel()
```

### 2. Handle Errors Appropriately

```go
func (r *UserRepository) FindByEmail(email string) (*User, error) {
    var user User
    err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
    
    if err != nil {
        if mongodb.IsNoDocumentsError(err) {
            return nil, ErrUserNotFound
        }
        if mongodb.IsNetworkError(err) {
            r.logger.Error("Network error", "error", err)
            return nil, ErrTemporaryFailure
        }
        return nil, fmt.Errorf("find user: %w", err)
    }
    
    return &user, nil
}
```

### 3. Use Indexes for Performance

```go
func SetupIndexes(db *mongo.Database) error {
    // Users collection indexes
    usersIndexes := []mongo.IndexModel{
        mongodb.BuildUniqueIndex(bson.D{{"email", 1}}, "unique_email"),
        mongodb.BuildIndexModel(bson.D{{"created_at", -1}}),
        mongodb.BuildTextIndex([]string{"name", "bio"}, "user_search"),
    }
    
    ctx, cancel := mongodb.NewContext(30 * time.Second)
    defer cancel()
    
    return mongodb.CreateIndexes(ctx, db.Collection("users"), usersIndexes)
}
```

### 4. Use Typed Repositories

```go
// Define repository interface
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    Paginate(ctx context.Context, filter bson.M, page, size int64) (*PaginatedResult[User], error)
}

// Implement with typed repository
type userRepository struct {
    repo *mongodb.Repository[User]
}

func NewUserRepository(db *mongo.Database) UserRepository {
    return &userRepository{
        repo: mongodb.NewRepository[User](db, "users"),
    }
}
```

## Performance Optimization

### 1. Use Projection

```go
// Only fetch required fields
opts := options.Find().SetProjection(bson.D{
    {"name", 1},
    {"email", 1},
    {"active", 1},
})
users, err := repo.Find(ctx, bson.M{}, opts)
```

### 2. Use Proper Pool Settings

```bash
# Configure connection pool
MONGO_DB_MIN_POOL_SIZE=10
MONGO_DB_MAX_POOL_SIZE=100
MONGO_DB_MAX_CONN_IDLE_TIME=60s
```

### 3. Use Bulk Operations

```go
// Instead of multiple individual operations
var models []mongo.WriteModel
for _, user := range users {
    model := mongo.NewInsertOneModel().SetDocument(user)
    models = append(models, model)
}

result, err := collection.BulkWrite(ctx, models)
```

## Observability

When observability is enabled, the MongoDB module provides:

### Metrics
- `mongodb_operations_total` - Total MongoDB operations
- `mongodb_operation_duration_seconds` - Operation duration
- `mongodb_errors_total` - Total errors
- `mongodb_commands_total` - Command counts
- `mongodb_command_duration_seconds` - Command duration

### Tracing
- Automatic span creation for all operations
- Command-level tracing
- Error recording and status tracking

## Migration from SQL Databases

If migrating from SQL databases:

```go
// SQL: SELECT * FROM users WHERE age > 18 ORDER BY name
filter := bson.M{"age": bson.M{"$gt": 18}}
opts := options.Find().SetSort(bson.D{{"name", 1}})
cursor, err := collection.Find(ctx, filter, opts)

// SQL: SELECT COUNT(*) FROM users WHERE active = true
count, err := collection.CountDocuments(ctx, bson.M{"active": true})

// SQL: UPDATE users SET active = false WHERE last_login < ?
filter := bson.M{"last_login": bson.M{"$lt": thirtyDaysAgo}}
update := bson.M{"$set": bson.M{"active": false}}
result, err := collection.UpdateMany(ctx, filter, update)

// SQL: DELETE FROM users WHERE created_at < ?
filter := bson.M{"created_at": bson.M{"$lt": oneYearAgo}}
result, err := collection.DeleteMany(ctx, filter)
```

## Next Steps

- [**Caching**](./caching.md) - Add caching to MongoDB operations
- [**Event System**](./event-system.md) - Integrate with event system
- [**Testing**](./testing.md) - Test MongoDB operations
- [**Best Practices**](./best-practices.md) - MongoDB development best practices