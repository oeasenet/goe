# MongoDB CRUD Operations Example

This example demonstrates how to build a complete CRUD service using GOE's simplified MongoDB integration with clean, direct access to the MongoDB driver.

## Overview

This example shows:
- Setting up MongoDB with GOE's clean contract interface
- Creating models with proper BSON tags
- Implementing service layer with dependency injection
- CRUD operations using the native MongoDB driver directly
- Error handling and validation
- Index management
- Testing approaches

## Project Structure

```
example/
├── main.go                 # Application entry point
├── models/
│   ├── user.go            # User model
│   └── post.go            # Post model
├── services/
│   ├── user_service.go    # User service implementation
│   └── post_service.go    # Post service implementation
└── .env                   # Environment configuration
```

## Environment Configuration

```.env
# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=blog_db

# Optional: Connection pool settings
MONGO_MIN_POOL_SIZE=5
MONGO_MAX_POOL_SIZE=100
MONGO_TIMEOUT=10s
```

## Models

### User Model

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username  string            `bson:"username" json:"username" validate:"required,min=3,max=20"`
    Email     string            `bson:"email" json:"email" validate:"required,email"`
    FullName  string            `bson:"full_name" json:"full_name" validate:"required"`
    Active    bool              `bson:"active" json:"active"`
    CreatedAt time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=20"`
    Email    string `json:"email" validate:"required,email"`
    FullName string `json:"full_name" validate:"required"`
}

type UpdateUserRequest struct {
    Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=20"`
    Email    *string `json:"email,omitempty" validate:"omitempty,email"`
    FullName *string `json:"full_name,omitempty"`
    Active   *bool   `json:"active,omitempty"`
}
```

## Service Layer

### User Service

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
    
    "your-app/models"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id string) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, id string, updates bson.M) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]*models.User, error)
    Count(ctx context.Context) (int64, error)
}

type userRepository struct {
    db     contract.MongoDB
    logger contract.Logger
}

func NewUserRepository(mongoDB contract.MongoDB, logger contract.Logger) UserRepository {
    return &userRepository{
        db:     mongoDB,
        logger: logger,
    }
}

func (r *userRepository) Initialize(ctx context.Context) error {
    collection := r.db.Collection("users")
    
    indexes := []mongo.IndexModel{
        mongodb.BuildUniqueIndex(bson.D{{"email", 1}}, "unique_email"),
        mongodb.BuildUniqueIndex(bson.D{{"username", 1}}, "unique_username"),
        mongodb.BuildIndexModel(bson.D{{"active", 1}, {"created_at", -1}}),
        mongodb.BuildTextIndex([]string{"username", "full_name"}, "user_search"),
    }
    
    return mongodb.EnsureIndexes(ctx, collection, indexes)
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    user.ID = primitive.NewObjectID()
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()
    user.Active = true
    
    collection := r.db.Collection("users")
    _, err := collection.InsertOne(ctx, user)
    
    if mongodb.IsDuplicateKeyError(err) {
        return errors.New("user with this email or username already exists")
    }
    
    return err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, errors.New("invalid user ID format")
    }
    
    collection := r.db.Collection("users")
    var user models.User
    
    err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
    if mongodb.IsNoDocumentsError(err) {
        return nil, errors.New("user not found")
    }
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    collection := r.db.Collection("users")
    var user models.User
    
    err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if mongodb.IsNoDocumentsError(err) {
        return nil, errors.New("user not found")
    }
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (r *userRepository) Update(ctx context.Context, id string, updates bson.M) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return errors.New("invalid user ID format")
    }
    
    // Always update the UpdatedAt field
    updates["updated_at"] = time.Now()
    
    collection := r.db.Collection("users")
    result, err := collection.UpdateOne(ctx, 
        bson.M{"_id": objectID},
        bson.M{"$set": updates},
    )
    
    if err != nil {
        if mongodb.IsDuplicateKeyError(err) {
            return errors.New("email or username already exists")
        }
        return err
    }
    
    if result.MatchedCount == 0 {
        return errors.New("user not found")
    }
    
    return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return errors.New("invalid user ID format")
    }
    
    collection := r.db.Collection("users")
    result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
    
    if err != nil {
        return err
    }
    
    if result.DeletedCount == 0 {
        return errors.New("user not found")
    }
    
    return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
    collection := r.db.Collection("users")
    
    opts := options.Find().
        SetSort(bson.D{{"created_at", -1}}).
        SetLimit(int64(limit)).
        SetSkip(int64(offset))
    
    cursor, err := collection.Find(ctx, bson.M{"active": true}, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var users []*models.User
    if err := cursor.All(ctx, &users); err != nil {
        return nil, err
    }
    
    return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
    collection := r.db.Collection("users")
    return collection.CountDocuments(ctx, bson.M{"active": true})
}
```

## Application Setup

### Main Application

```go
package main

import (
    "context"
    "log"
    
    "go.oease.dev/goe/v2"
    "your-app/services"
)

func main() {
    goe.New(goe.Options{
        WithMongoDB: true,
        Providers: []any{
            services.NewUserRepository,
        },
        Invokers: []any{
            initializeRepositories,
        },
    })
    
    goe.Run()
}

func initializeRepositories(userRepo services.UserRepository) {
    ctx := context.Background()
    
    // Initialize indexes for users
    if userRepoWithInit, ok := userRepo.(*services.userRepository); ok {
        if err := userRepoWithInit.Initialize(ctx); err != nil {
            log.Printf("Failed to initialize user indexes: %v", err)
        }
    }
    
    log.Println("Application initialized successfully")
}
```

## Advanced Examples

### Aggregation Pipeline Example

```go
func (r *userRepository) GetUserStats(ctx context.Context) (*UserStats, error) {
    collection := r.db.Collection("users")
    
    pipeline := mongo.Pipeline{
        {{"$group", bson.D{
            {"_id", nil},
            {"total_users", bson.D{{"$sum", 1}}},
            {"active_users", bson.D{{"$sum", bson.D{{"$cond", bson.A{"$active", 1, 0}}}}}},
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

type UserStats struct {
    TotalUsers  int64 `bson:"total_users" json:"total_users"`
    ActiveUsers int64 `bson:"active_users" json:"active_users"`
}
```

### Transaction Example

```go
func (s *UserService) TransferCredits(ctx context.Context, fromID, toID string, amount int) error {
    client := s.db.Client()
    if client == nil {
        return errors.New("MongoDB client not available")
    }
    
    return mongodb.WithTransaction(ctx, client, func(sessCtx context.Context) error {
        collection := s.db.Collection("users")
        
        // Deduct from sender
        _, err := collection.UpdateOne(sessCtx,
            bson.M{"_id": fromID},
            bson.M{"$inc": bson.M{"credits": -amount}},
        )
        if err != nil {
            return err
        }
        
        // Add to receiver
        _, err = collection.UpdateOne(sessCtx,
            bson.M{"_id": toID}, 
            bson.M{"$inc": bson.M{"credits": amount}},
        )
        return err
    })
}
```

### Bulk Operations Example

```go
func (r *userRepository) BulkUpdateStatus(ctx context.Context, updates []BulkStatusUpdate) error {
    collection := r.db.Collection("users")
    
    var models []mongo.WriteModel
    
    for _, update := range updates {
        objectID, err := primitive.ObjectIDFromHex(update.UserID)
        if err != nil {
            continue // Skip invalid IDs
        }
        
        model := mongodb.CreateUpdateOneModel(
            bson.M{"_id": objectID},
            bson.M{"$set": bson.M{
                "active": update.Active,
                "updated_at": time.Now(),
            }},
        )
        models = append(models, model)
    }
    
    if len(models) == 0 {
        return errors.New("no valid updates provided")
    }
    
    result, err := collection.BulkWrite(ctx, models)
    if err != nil {
        return err
    }
    
    log.Printf("Bulk update completed: %d modified", result.ModifiedCount)
    return nil
}

type BulkStatusUpdate struct {
    UserID string `json:"user_id"`
    Active bool   `json:"active"`
}
```

## Key Benefits of Simplified API

The new simplified MongoDB contract provides:

### ✅ **Direct MongoDB Driver Access**
```go
// Direct access to native MongoDB features
collection := db.Collection("users")
cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"created_at", -1}}))
```

### ✅ **No Wrapper Overhead**
```go
// No abstraction layers - full MongoDB functionality available
client := db.Client()
database := db.Instance()
collection := db.Collection("users")
```

### ✅ **Simpler, Cleaner Code**
```go
// Before: Complex wrapper patterns
// repo := db.Repository("users").(mongodb.RepositoryFactory).Create[User]()

// After: Simple, direct access  
collection := db.Collection("users")
```

### ✅ **Better Performance**
- No wrapper object creation
- Direct MongoDB driver calls
- Zero abstraction overhead

### ✅ **Full MongoDB Features**
- Access to all MongoDB driver capabilities
- Advanced aggregation pipelines
- Change streams
- GridFS
- Transactions with full control

This approach gives you the full power of the MongoDB Go driver while maintaining clean dependency injection and connection management through GOE's contract system.