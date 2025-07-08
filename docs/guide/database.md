# Database

GOE provides seamless database integration through GORM, supporting multiple database drivers and offering features like connection pooling, migrations, and lifecycle management.

## Overview

The database module provides:
- **GORM Integration**: Full ORM capabilities with associations, hooks, and transactions
- **Multiple Drivers**: Support for PostgreSQL, MySQL, SQLite, and SQL Server
- **Connection Pooling**: Configurable connection pool settings
- **Auto-Migration**: Automatic schema migration in development
- **Lifecycle Management**: Proper connection setup and teardown

## Supported Databases

- **PostgreSQL** (`postgres`, `pgsql`, `postgresql`)
- **MySQL** (`mysql`)
- **SQLite** (`sqlite`, `sqlite3`)
- **SQL Server** (`sqlserver`, `mssql`)

## Configuration

Configure your database through environment variables:

```bash
# PostgreSQL
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=secret
DB_SSL_MODE=disable
DB_TIMEZONE=UTC

# MySQL
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=myapp
DB_USER=root
DB_PASSWORD=secret
DB_CHARSET=utf8mb4
DB_PARSE_TIME=true
DB_TIMEZONE=Local

# SQLite
DB_DRIVER=sqlite
DB_PATH=./database.db

# SQL Server
DB_DRIVER=sqlserver
DB_HOST=localhost
DB_PORT=1433
DB_NAME=myapp
DB_USER=sa
DB_PASSWORD=secret
```

## Enabling Database Module

Enable the database module in your GOE application:

```go
package main

import "go.oease.dev/goe/v2"

func main() {
    goe.New(goe.Options{
        WithDB: true,
    })
    
    goe.Run()
}
```

## Accessing the Database

### Using Global Accessor

```go
import "go.oease.dev/goe/v2"

func someFunction() {
    db := goe.DB()
    
    var users []User
    db.Instance().Find(&users)
}
```

### Using Dependency Injection

```go
import "go.oease.dev/goe/v2/contract"

type UserRepository struct {
    db contract.DB
}

func NewUserRepository(db contract.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindAll() ([]User, error) {
    var users []User
    err := r.db.Instance().Find(&users).Error
    return users, err
}
```

## Models

Define your database models using GORM conventions:

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:100;not null"`
    Email     string    `gorm:"size:100;uniqueIndex;not null"`
    Password  string    `gorm:"size:255;not null"`
    Active    bool      `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // Associations
    Posts []Post `gorm:"foreignKey:UserID"`
}

type Post struct {
    ID        uint      `gorm:"primaryKey"`
    Title     string    `gorm:"size:200;not null"`
    Content   string    `gorm:"type:text"`
    UserID    uint      `gorm:"not null"`
    User      User      `gorm:"foreignKey:UserID"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## CRUD Operations

### Create

```go
func (r *UserRepository) Create(user *User) error {
    return r.db.Instance().Create(user).Error
}

func (r *UserRepository) CreateInBatches(users []User) error {
    return r.db.Instance().CreateInBatches(users, 100).Error
}
```

### Read

```go
func (r *UserRepository) FindByID(id uint) (*User, error) {
    var user User
    err := r.db.Instance().First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*User, error) {
    var user User
    err := r.db.Instance().Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) FindAll() ([]User, error) {
    var users []User
    err := r.db.Instance().Find(&users).Error
    return users, err
}

func (r *UserRepository) FindActiveUsers() ([]User, error) {
    var users []User
    err := r.db.Instance().Where("active = ?", true).Find(&users).Error
    return users, err
}
```

### Update

```go
func (r *UserRepository) Update(user *User) error {
    return r.db.Instance().Save(user).Error
}

func (r *UserRepository) UpdateFields(id uint, updates map[string]interface{}) error {
    return r.db.Instance().Model(&User{}).Where("id = ?", id).Updates(updates).Error
}
```

### Delete

```go
func (r *UserRepository) Delete(id uint) error {
    return r.db.Instance().Delete(&User{}, id).Error
}

func (r *UserRepository) SoftDelete(id uint) error {
    return r.db.Instance().Delete(&User{}, id).Error // GORM soft delete
}

func (r *UserRepository) HardDelete(id uint) error {
    return r.db.Instance().Unscoped().Delete(&User{}, id).Error
}
```

## Query Building

### Advanced Queries

```go
func (r *UserRepository) FindUsersWithPosts() ([]User, error) {
    var users []User
    err := r.db.Instance().
        Preload("Posts").
        Find(&users).Error
    return users, err
}

func (r *UserRepository) FindUsersByAgeRange(minAge, maxAge int) ([]User, error) {
    var users []User
    err := r.db.Instance().
        Where("age BETWEEN ? AND ?", minAge, maxAge).
        Order("age DESC").
        Find(&users).Error
    return users, err
}

func (r *UserRepository) FindUsersWithPagination(page, limit int) ([]User, int64, error) {
    var users []User
    var total int64
    
    offset := (page - 1) * limit
    
    err := r.db.Instance().Model(&User{}).Count(&total).Error
    if err != nil {
        return nil, 0, err
    }
    
    err = r.db.Instance().
        Offset(offset).
        Limit(limit).
        Find(&users).Error
    
    return users, total, err
}
```

### Raw Queries

```go
func (r *UserRepository) GetUserStats() (map[string]interface{}, error) {
    var result map[string]interface{}
    
    err := r.db.Instance().
        Raw("SELECT COUNT(*) as total_users, AVG(age) as avg_age FROM users").
        Scan(&result).Error
    
    return result, err
}
```

## Transactions

```go
func (r *UserRepository) CreateUserWithPost(user *User, post *Post) error {
    return r.db.Instance().Transaction(func(tx *gorm.DB) error {
        // Create user
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        // Set user ID for post
        post.UserID = user.ID
        
        // Create post
        if err := tx.Create(post).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

## Migrations

### Auto-Migration

GOE can automatically migrate your models:

```go
func RegisterMigrations(db contract.DB) {
    // Auto-migrate models
    db.Instance().AutoMigrate(
        &User{},
        &Post{},
        &Category{},
    )
}

func main() {
    goe.New(goe.Options{
        WithDB: true,
        Invokers: []any{RegisterMigrations},
    })
    
    goe.Run()
}
```

### Manual Migrations

For production, use manual migrations:

```go
func RunMigrations(db contract.DB) error {
    // Create users table
    if !db.Instance().Migrator().HasTable(&User{}) {
        if err := db.Instance().Migrator().CreateTable(&User{}); err != nil {
            return err
        }
    }
    
    // Add column
    if !db.Instance().Migrator().HasColumn(&User{}, "last_login") {
        if err := db.Instance().Migrator().AddColumn(&User{}, "last_login"); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Connection Pooling

Configure connection pool settings:

```bash
# Connection pool settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_MAX_CONN_LIFETIME=300s
DB_MAX_IDLE_TIME=60s
```

## Database Service Pattern

Create a database service for complex operations:

```go
type UserService struct {
    db     contract.DB
    logger contract.Logger
}

func NewUserService(db contract.DB, logger contract.Logger) *UserService {
    return &UserService{db: db, logger: logger}
}

func (s *UserService) CreateUser(name, email, password string) (*User, error) {
    s.logger.Info("Creating user", "email", email)
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        s.logger.Error("Failed to hash password", "error", err)
        return nil, err
    }
    
    user := &User{
        Name:     name,
        Email:    email,
        Password: string(hashedPassword),
        Active:   true,
    }
    
    if err := s.db.Instance().Create(user).Error; err != nil {
        s.logger.Error("Failed to create user", "error", err, "email", email)
        return nil, err
    }
    
    s.logger.Info("User created successfully", "user_id", user.ID, "email", email)
    return user, nil
}

func (s *UserService) AuthenticateUser(email, password string) (*User, error) {
    var user User
    
    err := s.db.Instance().Where("email = ? AND active = ?", email, true).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrInvalidCredentials
        }
        return nil, err
    }
    
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return nil, ErrInvalidCredentials
    }
    
    return &user, nil
}
```

## HTTP Integration

Use database services in HTTP handlers:

```go
type UserHandler struct {
    userService *UserService
    logger      contract.Logger
}

func NewUserHandler(userService *UserService, logger contract.Logger) *UserHandler {
    return &UserHandler{
        userService: userService,
        logger:      logger,
    }
}

func (h *UserHandler) GetUsers(c fiber.Ctx) error {
    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 10)
    
    users, total, err := h.userService.GetUsersWithPagination(page, limit)
    if err != nil {
        h.logger.Error("Failed to get users", "error", err)
        return c.Status(500).JSON(fiber.Map{
            "error": "Internal server error",
        })
    }
    
    return c.JSON(fiber.Map{
        "data":  users,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
    var req CreateUserRequest
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    user, err := h.userService.CreateUser(req.Name, req.Email, req.Password)
    if err != nil {
        h.logger.Error("Failed to create user", "error", err)
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to create user",
        })
    }
    
    return c.Status(201).JSON(user)
}
```

## Testing

### Database Testing

```go
func TestUserRepository(t *testing.T) {
    // Setup test database
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    // Auto-migrate
    err = db.AutoMigrate(&User{})
    require.NoError(t, err)
    
    // Create repository
    mockDB := &MockDB{instance: db}
    repo := NewUserRepository(mockDB)
    
    // Test create
    user := &User{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    err = repo.Create(user)
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
    
    // Test find
    found, err := repo.FindByEmail("john@example.com")
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", found.Name)
}
```

### Mock Database

```go
type MockDB struct {
    instance *gorm.DB
}

func (m *MockDB) Instance() *gorm.DB {
    return m.instance
}

func (m *MockDB) IsConnected() bool {
    return true
}
```

## Best Practices

### 1. Use Repository Pattern

```go
type UserRepository interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
    FindByEmail(email string) (*User, error)
    Update(user *User) error
    Delete(id uint) error
}

type userRepository struct {
    db contract.DB
}

func NewUserRepository(db contract.DB) UserRepository {
    return &userRepository{db: db}
}
```

### 2. Handle Errors Properly

```go
func (r *userRepository) FindByID(id uint) (*User, error) {
    var user User
    err := r.db.Instance().First(&user, id).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    
    return &user, nil
}
```

### 3. Use Transactions for Related Operations

```go
func (s *UserService) CreateUserWithProfile(userData UserData, profileData ProfileData) error {
    return s.db.Instance().Transaction(func(tx *gorm.DB) error {
        user := &User{Name: userData.Name, Email: userData.Email}
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        profile := &Profile{UserID: user.ID, Bio: profileData.Bio}
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### 4. Use Indexes for Performance

```go
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"size:100;uniqueIndex;not null"`
    Name  string `gorm:"size:100;index"`
}
```

## Performance Optimization

### 1. Use Preloading

```go
// Preload associations
var users []User
db.Preload("Posts").Find(&users)

// Preload with conditions
db.Preload("Posts", "published = ?", true).Find(&users)
```

### 2. Use Select

```go
// Select specific fields
var users []User
db.Select("id", "name", "email").Find(&users)
```

### 3. Use Batching

```go
// Batch operations
db.CreateInBatches(users, 100)
```

## Next Steps

- [**Caching**](./caching.md) - Add caching to database operations
- [**Testing**](./testing.md) - Test database operations
- [**Best Practices**](./best-practices.md) - Database development best practices