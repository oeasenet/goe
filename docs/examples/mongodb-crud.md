# MongoDB CRUD Operations

This example demonstrates how to build a complete CRUD application using GOE's MongoDB module with a blog system.

## Project Structure

```
blog-app/
├── main.go
├── models/
│   ├── user.go
│   └── post.go
├── repositories/
│   ├── user_repository.go
│   └── post_repository.go
├── services/
│   ├── user_service.go
│   └── post_service.go
└── handlers/
    ├── user_handler.go
    └── post_handler.go
```

## Models

### User Model

```go
// models/user.go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username  string            `bson:"username" json:"username"`
    Email     string            `bson:"email" json:"email"`
    Password  string            `bson:"password" json:"-"`
    FullName  string            `bson:"full_name" json:"full_name"`
    Bio       string            `bson:"bio" json:"bio"`
    Avatar    string            `bson:"avatar" json:"avatar"`
    Active    bool              `bson:"active" json:"active"`
    CreatedAt time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    FullName string `json:"full_name" validate:"required,min=2"`
    Bio      string `json:"bio" validate:"max=500"`
}

type UpdateUserRequest struct {
    FullName string `json:"full_name" validate:"min=2"`
    Bio      string `json:"bio" validate:"max=500"`
    Avatar   string `json:"avatar" validate:"url"`
}
```

### Post Model

```go
// models/post.go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type Post struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Title       string            `bson:"title" json:"title"`
    Slug        string            `bson:"slug" json:"slug"`
    Content     string            `bson:"content" json:"content"`
    Excerpt     string            `bson:"excerpt" json:"excerpt"`
    AuthorID    primitive.ObjectID `bson:"author_id" json:"author_id"`
    Author      *User             `bson:"author,omitempty" json:"author,omitempty"`
    Tags        []string          `bson:"tags" json:"tags"`
    Categories  []string          `bson:"categories" json:"categories"`
    Published   bool              `bson:"published" json:"published"`
    FeaturedImg string            `bson:"featured_img" json:"featured_img"`
    Views       int64             `bson:"views" json:"views"`
    Likes       []primitive.ObjectID `bson:"likes" json:"likes"`
    CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
    PublishedAt *time.Time        `bson:"published_at,omitempty" json:"published_at,omitempty"`
}

type CreatePostRequest struct {
    Title       string   `json:"title" validate:"required,min=5,max=200"`
    Content     string   `json:"content" validate:"required,min=10"`
    Excerpt     string   `json:"excerpt" validate:"max=500"`
    Tags        []string `json:"tags" validate:"max=10"`
    Categories  []string `json:"categories" validate:"max=5"`
    FeaturedImg string   `json:"featured_img" validate:"url"`
}

type UpdatePostRequest struct {
    Title       string   `json:"title" validate:"min=5,max=200"`
    Content     string   `json:"content" validate:"min=10"`
    Excerpt     string   `json:"excerpt" validate:"max=500"`
    Tags        []string `json:"tags" validate:"max=10"`
    Categories  []string `json:"categories" validate:"max=5"`
    FeaturedImg string   `json:"featured_img" validate:"url"`
    Published   *bool    `json:"published"`
}
```

## Repositories

### User Repository

```go
// repositories/user_repository.go
package repositories

import (
    "context"
    "fmt"
    "strings"
    "time"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    
    "your-app/models"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/mongodb"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByID(ctx context.Context, id string) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, error)
    FindByUsername(ctx context.Context, username string) (*models.User, error)
    Update(ctx context.Context, id string, updates bson.M) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, page, limit int64, search string) (*mongodb.PaginatedResult[models.User], error)
    GetStats(ctx context.Context) (map[string]interface{}, error)
}

type userRepository struct {
    repo   *mongodb.Repository[models.User]
    logger contract.Logger
}

func NewUserRepository(mongoDB contract.MongoDB, logger contract.Logger) UserRepository {
    db := mongoDB.Instance()
    repo := mongodb.NewRepository[models.User](db, "users")
    
    return &userRepository{
        repo:   repo,
        logger: logger,
    }
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    user.ID = primitive.NewObjectID()
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()
    user.Active = true
    
    _, err := r.repo.Insert(ctx, user)
    if err != nil {
        if mongodb.IsDuplicateKeyError(err) {
            return fmt.Errorf("user with email or username already exists")
        }
        r.logger.Error("Failed to create user", "error", err)
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    r.logger.Info("User created", "user_id", user.ID.Hex(), "email", user.Email)
    return nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, fmt.Errorf("invalid user ID: %w", err)
    }
    
    user, err := r.repo.FindByID(ctx, objectID)
    if err != nil {
        if mongodb.IsNoDocumentsError(err) {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    
    return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    user, err := r.repo.FindOne(ctx, bson.M{"email": strings.ToLower(email)})
    if err != nil {
        if mongodb.IsNoDocumentsError(err) {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    
    return user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
    user, err := r.repo.FindOne(ctx, bson.M{"username": username})
    if err != nil {
        if mongodb.IsNoDocumentsError(err) {
            return nil, fmt.Errorf("user not found")
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    
    return user, nil
}

func (r *userRepository) Update(ctx context.Context, id string, updates bson.M) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("invalid user ID: %w", err)
    }
    
    updates["updated_at"] = time.Now()
    
    result, err := r.repo.UpdateByID(ctx, objectID, bson.M{"$set": updates})
    if err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }
    
    if result.MatchedCount == 0 {
        return fmt.Errorf("user not found")
    }
    
    r.logger.Info("User updated", "user_id", id)
    return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("invalid user ID: %w", err)
    }
    
    result, err := r.repo.DeleteByID(ctx, objectID)
    if err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    
    if result.DeletedCount == 0 {
        return fmt.Errorf("user not found")
    }
    
    r.logger.Info("User deleted", "user_id", id)
    return nil
}

func (r *userRepository) List(ctx context.Context, page, limit int64, search string) (*mongodb.PaginatedResult[models.User], error) {
    filter := bson.M{"active": true}
    
    if search != "" {
        filter["$or"] = []bson.M{
            {"username": bson.M{"$regex": search, "$options": "i"}},
            {"full_name": bson.M{"$regex": search, "$options": "i"}},
            {"email": bson.M{"$regex": search, "$options": "i"}},
        }
    }
    
    opts := options.Find().
        SetSort(bson.D{{"created_at", -1}}).
        SetProjection(bson.D{{"password", 0}}) // Exclude password
    
    return r.repo.Paginate(ctx, filter, page, limit, opts)
}

func (r *userRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
    pipeline := mongo.Pipeline{
        {{"$group", bson.D{
            {"_id", nil},
            {"total_users", bson.D{{"$sum", 1}}},
            {"active_users", bson.D{{"$sum", bson.D{{"$cond", []interface{}{"$active", 1, 0}}}}}},
            {"inactive_users", bson.D{{"$sum", bson.D{{"$cond", []interface{}{"$active", 0, 1}}}}}},
        }}},
    }
    
    results, err := r.repo.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("failed to get user stats: %w", err)
    }
    
    if len(results) == 0 {
        return map[string]interface{}{
            "total_users":    0,
            "active_users":   0,
            "inactive_users": 0,
        }, nil
    }
    
    stats := map[string]interface{}{}
    // Convert result to map (implementation depends on your needs)
    return stats, nil
}
```

### Post Repository

```go
// repositories/post_repository.go
package repositories

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/bson/primitive"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    
    "your-app/models"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/mongodb"
)

type PostRepository interface {
    Create(ctx context.Context, post *models.Post) error
    FindByID(ctx context.Context, id string) (*models.Post, error)
    FindBySlug(ctx context.Context, slug string) (*models.Post, error)
    Update(ctx context.Context, id string, updates bson.M) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, page, limit int64, published *bool) (*mongodb.PaginatedResult[models.Post], error)
    FindByAuthor(ctx context.Context, authorID string, page, limit int64) (*mongodb.PaginatedResult[models.Post], error)
    FindByTags(ctx context.Context, tags []string, page, limit int64) (*mongodb.PaginatedResult[models.Post], error)
    Search(ctx context.Context, query string, page, limit int64) (*mongodb.PaginatedResult[models.Post], error)
    IncrementViews(ctx context.Context, id string) error
    ToggleLike(ctx context.Context, postID, userID string) error
}

type postRepository struct {
    repo   *mongodb.Repository[models.Post]
    logger contract.Logger
}

func NewPostRepository(mongoDB contract.MongoDB, logger contract.Logger) PostRepository {
    db := mongoDB.Instance()
    repo := mongodb.NewRepository[models.Post](db, "posts")
    
    return &postRepository{
        repo:   repo,
        logger: logger,
    }
}

func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
    post.ID = primitive.NewObjectID()
    post.CreatedAt = time.Now()
    post.UpdatedAt = time.Now()
    post.Views = 0
    post.Likes = []primitive.ObjectID{}
    
    if post.Published {
        now := time.Now()
        post.PublishedAt = &now
    }
    
    _, err := r.repo.Insert(ctx, post)
    if err != nil {
        r.logger.Error("Failed to create post", "error", err)
        return fmt.Errorf("failed to create post: %w", err)
    }
    
    r.logger.Info("Post created", "post_id", post.ID.Hex(), "title", post.Title)
    return nil
}

func (r *postRepository) FindByID(ctx context.Context, id string) (*models.Post, error) {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, fmt.Errorf("invalid post ID: %w", err)
    }
    
    // Use aggregation to populate author
    pipeline := mongo.Pipeline{
        {{"$match", bson.D{{"_id", objectID}}}},
        {{"$lookup", bson.D{
            {"from", "users"},
            {"localField", "author_id"},
            {"foreignField", "_id"},
            {"as", "author"},
            {"pipeline", mongo.Pipeline{
                {{"$project", bson.D{{"password", 0}}}},
            }},
        }}},
        {{"$unwind", bson.D{
            {"path", "$author"},
            {"preserveNullAndEmptyArrays", true},
        }}},
    }
    
    results, err := r.repo.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("failed to find post: %w", err)
    }
    
    if len(results) == 0 {
        return nil, fmt.Errorf("post not found")
    }
    
    return results[0], nil
}

func (r *postRepository) List(ctx context.Context, page, limit int64, published *bool) (*mongodb.PaginatedResult[models.Post], error) {
    filter := bson.M{}
    if published != nil {
        filter["published"] = *published
    }
    
    opts := options.Find().SetSort(bson.D{{"created_at", -1}})
    
    return r.repo.Paginate(ctx, filter, page, limit, opts)
}

func (r *postRepository) Search(ctx context.Context, query string, page, limit int64) (*mongodb.PaginatedResult[models.Post], error) {
    filter := bson.M{
        "$and": []bson.M{
            {"published": true},
            {"$or": []bson.M{
                {"title": bson.M{"$regex": query, "$options": "i"}},
                {"content": bson.M{"$regex": query, "$options": "i"}},
                {"excerpt": bson.M{"$regex": query, "$options": "i"}},
                {"tags": bson.M{"$in": []string{query}}},
            }},
        },
    }
    
    opts := options.Find().SetSort(bson.D{{"created_at", -1}})
    
    return r.repo.Paginate(ctx, filter, page, limit, opts)
}

func (r *postRepository) IncrementViews(ctx context.Context, id string) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("invalid post ID: %w", err)
    }
    
    _, err = r.repo.UpdateByID(ctx, objectID, bson.M{
        "$inc": bson.M{"views": 1},
    })
    
    return err
}

func (r *postRepository) ToggleLike(ctx context.Context, postID, userID string) error {
    postObjectID, err := primitive.ObjectIDFromHex(postID)
    if err != nil {
        return fmt.Errorf("invalid post ID: %w", err)
    }
    
    userObjectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return fmt.Errorf("invalid user ID: %w", err)
    }
    
    // Check if user already liked the post
    post, err := r.repo.FindByID(ctx, postObjectID)
    if err != nil {
        return fmt.Errorf("post not found: %w", err)
    }
    
    var update bson.M
    liked := false
    for _, likeID := range post.Likes {
        if likeID == userObjectID {
            liked = true
            break
        }
    }
    
    if liked {
        // Remove like
        update = bson.M{"$pull": bson.M{"likes": userObjectID}}
    } else {
        // Add like
        update = bson.M{"$addToSet": bson.M{"likes": userObjectID}}
    }
    
    _, err = r.repo.UpdateByID(ctx, postObjectID, update)
    return err
}
```

## Services

### User Service

```go
// services/user_service.go
package services

import (
    "context"
    "fmt"
    "strings"
    "time"

    "golang.org/x/crypto/bcrypt"
    "go.mongodb.org/mongo-driver/v2/bson"
    
    "your-app/models"
    "your-app/repositories"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/mongodb"
)

type UserService interface {
    CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
    GetUser(ctx context.Context, id string) (*models.User, error)
    UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.User, error)
    DeleteUser(ctx context.Context, id string) error
    ListUsers(ctx context.Context, page, limit int64, search string) (*mongodb.PaginatedResult[models.User], error)
    AuthenticateUser(ctx context.Context, email, password string) (*models.User, error)
}

type userService struct {
    userRepo repositories.UserRepository
    logger   contract.Logger
}

func NewUserService(userRepo repositories.UserRepository, logger contract.Logger) UserService {
    return &userService{
        userRepo: userRepo,
        logger:   logger,
    }
}

func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
    // Check if user already exists
    existing, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err == nil && existing != nil {
        return nil, fmt.Errorf("user with email already exists")
    }
    
    existing, err = s.userRepo.FindByUsername(ctx, req.Username)
    if err == nil && existing != nil {
        return nil, fmt.Errorf("username already taken")
    }
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }
    
    user := &models.User{
        Username: req.Username,
        Email:    strings.ToLower(req.Email),
        Password: string(hashedPassword),
        FullName: req.FullName,
        Bio:      req.Bio,
    }
    
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Remove password from response
    user.Password = ""
    return user, nil
}

func (s *userService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("invalid credentials")
    }
    
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return nil, fmt.Errorf("invalid credentials")
    }
    
    if !user.Active {
        return nil, fmt.Errorf("account is deactivated")
    }
    
    // Remove password from response
    user.Password = ""
    return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.User, error) {
    updates := bson.M{}
    
    if req.FullName != "" {
        updates["full_name"] = req.FullName
    }
    if req.Bio != "" {
        updates["bio"] = req.Bio
    }
    if req.Avatar != "" {
        updates["avatar"] = req.Avatar
    }
    
    if len(updates) == 0 {
        return nil, fmt.Errorf("no updates provided")
    }
    
    if err := s.userRepo.Update(ctx, id, updates); err != nil {
        return nil, err
    }
    
    return s.userRepo.FindByID(ctx, id)
}
```

## HTTP Handlers

### User Handler

```go
// handlers/user_handler.go
package handlers

import (
    "strconv"

    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
    
    "your-app/models"
    "your-app/services"
)

type UserHandler struct {
    userService services.UserService
    logger      contract.Logger
}

func NewUserHandler(userService services.UserService, logger contract.Logger) *UserHandler {
    return &UserHandler{
        userService: userService,
        logger:      logger,
    }
}

func (h *UserHandler) RegisterRoutes(app fiber.Router) {
    users := app.Group("/users")
    
    users.Post("/", h.CreateUser)
    users.Get("/", h.ListUsers)
    users.Get("/:id", h.GetUser)
    users.Put("/:id", h.UpdateUser)
    users.Delete("/:id", h.DeleteUser)
    
    // Authentication
    app.Post("/auth/login", h.Login)
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    var req models.CreateUserRequest
    if err := c.Bind().Body(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    user, err := h.userService.CreateUser(c.Context(), &req)
    if err != nil {
        h.logger.Error("Failed to create user", "error", err)
        return c.Status(400).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    return c.Status(201).JSON(fiber.Map{
        "message": "User created successfully",
        "user":    user,
    })
}

func (h *UserHandler) ListUsers(c fiber.Ctx) error {
    page, _ := strconv.ParseInt(c.Query("page", "1"), 10, 64)
    limit, _ := strconv.ParseInt(c.Query("limit", "20"), 10, 64)
    search := c.Query("search", "")
    
    result, err := h.userService.ListUsers(c.Context(), page, limit, search)
    if err != nil {
        h.logger.Error("Failed to list users", "error", err)
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch users",
        })
    }
    
    return c.JSON(result)
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
    userID := c.Params("id")
    
    user, err := h.userService.GetUser(c.Context(), userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{
            "error": "User not found",
        })
    }
    
    return c.JSON(user)
}

func (h *UserHandler) Login(c fiber.Ctx) error {
    var req struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required"`
    }
    
    if err := c.Bind().Body(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    user, err := h.userService.AuthenticateUser(c.Context(), req.Email, req.Password)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{
            "error": "Invalid credentials",
        })
    }
    
    // Generate JWT token here
    token := "generated-jwt-token"
    
    return c.JSON(fiber.Map{
        "message": "Login successful",
        "token":   token,
        "user":    user,
    })
}
```

## Main Application

```go
// main.go
package main

import (
    "context"

    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "go.oease.dev/goe/v2/core/mongodb"
    
    "your-app/handlers"
    "your-app/repositories"
    "your-app/services"
)

func main() {
    app := goe.New(goe.Options{
        WithHTTP:          true,
        WithMongoDB:       true,
        WithObservability: true,
        
        Providers: []any{
            repositories.NewUserRepository,
            repositories.NewPostRepository,
            services.NewUserService,
            services.NewPostService,
            handlers.NewUserHandler,
            handlers.NewPostHandler,
        },
        
        Invokers: []any{
            SetupIndexes,
            RegisterRoutes,
        },
    })
    
    goe.Run()
}

func SetupIndexes(mongoDB contract.MongoDB) error {
    db := mongoDB.Instance()
    ctx, cancel := mongodb.DefaultContext()
    defer cancel()
    
    // User indexes
    userIndexes := []mongo.IndexModel{
        mongodb.BuildUniqueIndex(bson.D{{"email", 1}}, "unique_email"),
        mongodb.BuildUniqueIndex(bson.D{{"username", 1}}, "unique_username"),
        mongodb.BuildIndexModel(bson.D{{"created_at", -1}}),
        mongodb.BuildIndexModel(bson.D{{"active", 1}}),
    }
    
    if err := mongodb.CreateIndexes(ctx, db.Collection("users"), userIndexes); err != nil {
        return err
    }
    
    // Post indexes
    postIndexes := []mongo.IndexModel{
        mongodb.BuildIndexModel(bson.D{{"author_id", 1}}),
        mongodb.BuildIndexModel(bson.D{{"published", 1}, {"created_at", -1}}),
        mongodb.BuildIndexModel(bson.D{{"slug", 1}}),
        mongodb.BuildIndexModel(bson.D{{"tags", 1}}),
        mongodb.BuildTextIndex([]string{"title", "content", "excerpt"}, "post_search"),
    }
    
    return mongodb.CreateIndexes(ctx, db.Collection("posts"), postIndexes)
}

func RegisterRoutes(
    http contract.HTTPKernel,
    userHandler *handlers.UserHandler,
    postHandler *handlers.PostHandler,
) {
    api := http.App().Group("/api/v1")
    
    userHandler.RegisterRoutes(api)
    postHandler.RegisterRoutes(api)
}
```

## Environment Configuration

```bash
# .env
# MongoDB Configuration
MONGO_DB_URI=mongodb://localhost:27017
MONGO_DB_DB_NAME=blog_app
MONGO_DB_MIN_POOL_SIZE=5
MONGO_DB_MAX_POOL_SIZE=50

# HTTP Configuration
HTTP_PORT=8080

# Observability
OTEL_ENABLED=true
OTEL_SERVICE_NAME=blog-api

# Logging
LOG_LEVEL=info
```

## Testing

```go
// main_test.go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "your-app/models"
)

func TestCreateUser(t *testing.T) {
    // Setup test app (use test containers for MongoDB)
    app := setupTestApp(t)
    
    req := models.CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
        FullName: "Test User",
        Bio:      "Test bio",
    }
    
    body, _ := json.Marshal(req)
    
    request := httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
    request.Header.Set("Content-Type", "application/json")
    
    response, err := app.Test(request)
    require.NoError(t, err)
    
    assert.Equal(t, 201, response.StatusCode)
    
    var result map[string]interface{}
    json.NewDecoder(response.Body).Decode(&result)
    
    assert.Equal(t, "User created successfully", result["message"])
    assert.NotNil(t, result["user"])
}
```

This comprehensive example demonstrates:

1. **Complete CRUD operations** for users and posts
2. **Generic repository pattern** using MongoDB utilities
3. **Service layer** with business logic
4. **HTTP handlers** with validation
5. **Index management** for performance
6. **Error handling** with MongoDB utilities
7. **Pagination** support
8. **Search functionality**
9. **Testing** setup

The example follows GOE's patterns and best practices while showcasing MongoDB's capabilities for modern applications.