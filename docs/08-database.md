# 8. Database Interactions with GORM 💾

Goe provides a robust database module that integrates [GORM](https://gorm.io/), a developer-friendly ORM library for Go. This module simplifies database interactions, connection management, and allows you to work with various SQL databases.

## Enabling the Database Module

To use the database features, you must enable the `DB` module when initializing your Goe application:

```go
package main

import (
	"go.oease.dev/goe/v2"
)

func main() {
	app := goe.New(goe.Options{
		WithDB: true, // Enable the Database module
		// ... other options
	})
	goe.Run()
}
```

## Configuration

Database connections are configured via environment variables, primarily prefixed with `DB_`.

### Default Connection

The main connection used by `goe.DB().Instance()` is determined by the `DB_CONNECTION` environment variable. If `DB_CONNECTION` is not set, it defaults to a connection named `"default"`.

The configuration keys for this connection would be:

*   **`DB_DRIVER`**: Specifies the database driver.
    *   Supported: `mysql`, `postgres` (or `pgsql`, `postgresql`), `sqlite` (or `sqlite3`), `sqlserver` (or `mssql`).
*   **`DB_HOST`**: Database server hostname or IP address.
*   **`DB_PORT`**: Database server port. (Defaults to standard ports if not set, e.g., MySQL: 3306, PostgreSQL: 5432).
*   **`DB_DATABASE`**: The name of the database.
    *   For SQLite, this is the path to the database file (e.g., `mydatabase.db`). If not specified, SQLite defaults to an in-memory database (`file::memory:?cache=shared` or `:memory:` depending on GORM driver specifics), useful for testing.
*   **`DB_USERNAME`**: Username for database authentication.
*   **`DB_PASSWORD`**: Password for database authentication.
*   **`DB_DSN`**: A full Data Source Name (DSN) string. If provided, GORM might use this directly, potentially overriding individual parameters like host, port, etc. The exact behavior depends on the GORM driver.

**Driver-Specific Options:**

*   **MySQL:**
    *   `DB_CHARSET`: Character set (e.g., `utf8mb4`). Defaults to `utf8mb4`.
    *   `DB_TIMEZONE`: Timezone for the connection (e.g., `UTC`, `Local`). Defaults to `Local`.
*   **PostgreSQL:**
    *   `DB_SSLMODE`: SSL mode (e.g., `disable`, `require`). Defaults to `disable`.
    *   `DB_TIMEZONE`: Timezone (e.g., `UTC`, `Asia/Shanghai`). Defaults to `UTC`.
*   **SQL Server:**
    *   `DB_ENCRYPT`: Encryption mode (e.g., `disable`, `true`).
    *   `DB_TRUST_SERVER_CERTIFICATE`: (`true`/`false`).

**Connection Pool Settings:**

*   `DB_MAX_IDLE_CONNS`: Maximum number of connections in the idle connection pool. Default: `10`.
*   `DB_MAX_OPEN_CONNS`: Maximum number of open connections to the database. Default: `100`.
*   `DB_CONN_MAX_LIFETIME`: Maximum amount of time a connection may be reused (e.g., `1h`, `30m`).
*   `DB_CONN_MAX_IDLE_TIME`: Maximum amount of time a connection may be idle (e.g., `5m`).

**GORM Behavior & Logging:**

*   `DB_LOG_MODE`: Set to `true` to enable GORM's built-in logger (prints SQL queries). Default: `false`.
*   `DB_IGNORE_RECORD_NOT_FOUND_ERROR`: Set to `true` to make GORM treat `ErrRecordNotFound` as a normal empty result rather than an error. Default: `false`.
*   `DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING`: (`true`/`false`) Disables foreign key constraints during GORM's auto-migration. Default: `false`.
*   `DB_AUTO_MIGRATE`: If `true`, logs that auto-migration is enabled. Actual model migration requires explicit calls (see [Auto Migration](#auto-migration)).
    *   This specific key applies if the default connection name is `default`.
    *   For a custom default connection name (e.g., `DB_CONNECTION=maindb`), the key would be `DB_MAINDB_AUTO_MIGRATE=true`.
*   `DB_AUTO_MIGRATE_ANY`: If `true`, enables the auto-migration hint for *any* configured connection if its specific `DB_<NAME>_AUTO_MIGRATE` is also true.

### Multiple Named Connections

Goe supports configuring multiple database connections. To configure an additional connection named `reporting`, you would use keys like:

*   `DB_REPORTING_DRIVER=postgres`
*   `DB_REPORTING_HOST=reporting.db.example.com`
*   ...and so on for all necessary parameters.

To activate these connections on startup, list their names (comma-separated) in the `DB_CONNECTIONS` environment variable:

*   `DB_CONNECTIONS=reporting,another_db`

The `DatabaseModule`'s `OnStart` method will attempt to connect to the default connection and any connections listed in `DB_CONNECTIONS`.

## Defining GORM Models

You define your GORM models as Go structs, typically embedding `gorm.Model` or defining your own primary keys and timestamp fields.

```go
package models // Or your preferred package, e.g., internal/domain/user

import "gorm.io/gorm"

// User represents a user in the system
type User struct {
	gorm.Model        // Adds ID, CreatedAt, UpdatedAt, DeletedAt
	Name         string
	Email        string `gorm:"uniqueIndex;not null"`
	Age          uint8
	MemberNumber string `gorm:"unique;not null"`
}

// Product represents a product
type Product struct {
	ID          uint   `gorm:"primaryKey"`
	Code        string `gorm:"uniqueIndex"`
	Price       uint
	Description string
	gorm.CreatedAt
	gorm.UpdatedAt
}
```
Refer to [GORM Model Definition](https://gorm.io/docs/models.html) for more details.

## Accessing Database Instances

### Default Connection

*   **Global Accessor**: `goe.DB().Instance() *gorm.DB`
*   **Dependency Injection**: Inject `contract.DB` and call `Instance()`.

```go
// Global access
db := goe.DB().Instance()
if db != nil {
    // Use db for GORM operations
}

// Dependency Injection
type MyRepository struct {
    db contract.DB
    logger contract.Logger
}

func NewMyRepository(db contract.DB, logger contract.Logger) *MyRepository {
    return &MyRepository{db: db, logger: logger}
}

func (r *MyRepository) GetUser(id uint) (*models.User, error) {
    var user models.User
    // Use r.db.Instance() to get the *gorm.DB for the default connection
    if err := r.db.Instance().First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            r.logger.Warn("User not found", contract.NewField("user_id", id))
            return nil, nil // Or a custom not found error
        }
        r.logger.Error("Failed to fetch user", contract.NewField("user_id", id), contract.NewField("error", err))
        return nil, err
    }
    return &user, nil
}
```

### Named Connections

To access a named connection (e.g., "reporting" configured as above):

*   Call `goe.DB().Connection("reporting") (*gorm.DB, error)`
*   Or, if `contract.DB` is injected: `dbService.Connection("reporting")`

```go
// Global access for a named connection
reportingDB, err := goe.DB().Connection("reporting")
if err != nil {
    goe.Log().Error("Failed to get reporting DB connection", contract.NewField("error", err))
    // Handle error
} else {
    // Use reportingDB
}

// Via injected contract.DB
func (r *MyRepository) GetReportData() ([]someReportModel, error) {
    conn, err := r.db.Connection("reporting")
    if err != nil {
        r.logger.Error("Reporting DB connection error", contract.NewField("error", err))
        return nil, err
    }
    var results []someReportModel
    // Use conn for GORM operations on the "reporting" database
    if err := conn.Find(&results).Error; err != nil {
        return nil, err
    }
    return results, nil
}
```

## CRUD Operations

Use standard GORM methods for Create, Read, Update, and Delete operations.

```go
// Assuming db is a *gorm.DB instance (e.g., from goe.DB().Instance())

// Create
newUser := models.User{Name: "Alice", Email: "alice@example.com", Age: 30}
result := db.Create(&newUser) // newUser.ID will be populated
if result.Error != nil {
    // Handle error
}
goe.Log().Info("Created user", contract.NewField("user_id", newUser.ID), contract.NewField("rows_affected", result.RowsAffected))

// Read
var user models.User
// Find user with primary key
db.First(&user, newUser.ID)
// Find user by attribute
db.First(&user, "email = ?", "alice@example.com")

var users []models.User
// Get all users
db.Find(&users)
// Get users with condition
db.Where("age > ?", 25).Find(&users)

// Update
// Update user's age
db.Model(&user).Update("Age", 31)
// Update multiple attributes
db.Model(&user).Updates(models.User{Name: "Alicia", Age: 32}) // Updates non-zero fields
db.Model(&user).Updates(map[string]interface{}{"Name": "Alicia II", "Age": 33})

// Delete
// Soft delete (if gorm.Model or gorm.DeletedAt is used)
db.Delete(&user, user.ID)
// Check GORM docs for permanent deletion if needed.
```

## Auto Migration

GORM can automatically create or update database tables based on your model definitions. Goe's database module provides two approaches for auto-migration:

1. **Model Registration System (Recommended)** - Pre-register models for automatic migration during startup
2. **Manual Migration** - Explicitly call migration methods after startup

### Model Registration System (Recommended)

The model registration system solves the common timing issue where developers tried to migrate models before database connections were established. With this approach, you register your models during application initialization, and they are automatically migrated when the database connections are ready.

**Available Methods:**

*   `RegisterModelsForMigration(dst ...interface{})`: Pre-registers models for automatic migration on the **default** connection
*   `RegisterModelsForMigrationOnConnection(connectionName string, dst ...interface{})`: Pre-registers models for automatic migration on a **named** connection
*   `AutoMigrate(dst ...interface{}) error`: Manual migration on the **default** connection
*   `AutoMigrateOnConnection(connectionName string, dst ...interface{}) error`: Manual migration on a **named** connection

**How to use Model Registration:**

```go
package main

import (
    "log"
    "go.oease.dev/goe/v2"
    "go.oease.dev/goe/v2/contract"
    "example.com/yourproject/internal/models" // Your models package
)

func main() {
    // Create a GOE application with DB enabled and register models via Invokers
    goe.New(goe.Options{
        WithDB: true,
        // SOLUTION: Use Invokers to register models for migration BEFORE starting the application
        // This is the key improvement - you can now pre-register models
        // and they will be automatically migrated when the DB connections are established
        Invokers: []any{
            func(db contract.DB) {
                // Register models for automatic migration on the default connection
                // These will be migrated automatically during startup if auto-migration is enabled
                db.RegisterModelsForMigration(
                    &models.User{},
                    &models.Product{},
                    &models.Order{},
                    // ... other models for the default DB
                )

                // You can also register models for specific connections
                db.RegisterModelsForMigrationOnConnection("analytics", &models.AnalyticsEvent{})
                db.RegisterModelsForMigrationOnConnection("reporting", &models.ReportSummary{})

                log.Println("Models registered for auto-migration")
            },
        },
    })

    // Start the application - this will:
    // 1. Establish database connections
    // 2. Automatically migrate registered models (if auto-migration is enabled in config)
    // 3. Start other services
    goe.Run()

    log.Println("Application started successfully with auto-migrated models")
}
```

**Configuration for Auto-Migration:**

To enable automatic migration of registered models, set the appropriate configuration:

```bash
# Enable auto-migration for the default connection
DB_AUTO_MIGRATE=true

# Or enable auto-migration for all connections
DB_AUTO_MIGRATE_ANY=true

# Enable auto-migration for specific named connections
DB_ANALYTICS_AUTO_MIGRATE=true
DB_REPORTING_AUTO_MIGRATE=true

# Database connection settings
DB_DRIVER=sqlite
DB_DATABASE=./app.db
```

### Manual Migration (Legacy Approach)

If you prefer manual control over when migrations occur, you can still use the traditional approach:

```go
// RunMigrations is an Fx invoker for manual migration
func RunMigrations(db contract.DB, logger contract.Logger) error {
    logger.Info("Running database auto-migrations...")

    // Migrate models on the default connection
    err := db.AutoMigrate(
        &models.User{},
        &models.Product{},
        // ... other models for the default DB
    )
    if err != nil {
        logger.Error("Failed to auto-migrate default database tables", "error", err)
        return err // Returning an error will stop app startup if this is critical
    }
    logger.Info("Default database auto-migration successful.")

    // Example: Migrate models on a named connection "reporting_db"
    err = db.AutoMigrateOnConnection("reporting_db", &models.ReportSummary{})
    if err != nil {
        logger.Error("Failed to auto-migrate reporting_db tables", "error", err)
        return err
    }
    logger.Info("Reporting_db auto-migration successful.")

    return nil
}
```

**Key Benefits of Model Registration System:**

1. **Timing Issue Solved**: Models are registered before DB connections are established, then automatically migrated during the startup phase when connections are ready.

2. **Clean DI Integration**: Works seamlessly with the GOE framework's dependency injection system without requiring manual timing control.

3. **Backward Compatibility**: Existing `AutoMigrate()` methods still work for manual migration.

4. **Configuration-Driven**: Auto-migration only happens if enabled in configuration.

5. **Multi-Connection Support**: Can register different models for different database connections.

6. **No More Errors**: Eliminates the "database connection 'default' not found" error that occurred when trying to migrate before connections were established.

**Important Notes on Auto Migration:**

*   **Development vs. Production**: GORM's auto-migration is very convenient for development and testing. However, for production environments, it's generally safer and more controllable to use dedicated migration tools (like Goose, Atlas, Flyway, Liquibase, or GORM's own migrator tool). These tools offer versioning, rollbacks, and more fine-grained control over schema changes.
*   **Limitations**: Auto-migration might not handle all complex schema changes perfectly (e.g., renaming columns, changing column types with data preservation). Always test schema changes thoroughly.
*   **Registration Timing**: Model registration must happen during application initialization (in Invokers) before `goe.Run()` is called.
*   **Configuration Control**: Auto-migration only occurs if explicitly enabled in configuration. Models are registered but not migrated unless the appropriate config flags are set.

## Transactions

GORM supports database transactions for performing multiple operations atomically.

```go
func CreateUserWithProfile(db *gorm.DB, userName, userEmail string, profileInfo string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Create User
		user := models.User{Name: userName, Email: userEmail}
		if err := tx.Create(&user).Error; err != nil {
			// Any error returned here will roll back the transaction
			return err
		}

		// Create UserProfile (assuming a UserProfile model exists)
		// profile := models.UserProfile{UserID: user.ID, Info: profileInfo}
		// if err := tx.Create(&profile).Error; err != nil {
		//     return err
		// }

		// If all operations succeed, the transaction is committed when the function returns nil
		return nil
	})
}

// Usage:
// defaultDB := goe.DB().Instance()
// err := CreateUserWithProfile(defaultDB, "John Doe", "john@example.com", "Loves Go")
// if err != nil {
//     goe.Log().Error("Transaction failed", contract.NewField("error", err))
// }
```
Refer to [GORM Transactions](https://gorm.io/docs/transactions.html) for more details.

## Best Practices

### Model Registration and Migration

*   **Use Model Registration**: Prefer the model registration system over manual migration for better timing control and cleaner code organization.
*   **Register Early**: Always register models in Invokers during application initialization, before `goe.Run()` is called.
*   **Configuration-Driven Migration**: Use configuration flags (`DB_AUTO_MIGRATE`, `DB_AUTO_MIGRATE_ANY`) to control when auto-migration occurs, especially useful for different environments.
*   **Group Related Models**: Register related models together for the same connection to ensure proper foreign key relationships are established.

### Architecture and Design

*   **Use Dependency Injection**: Inject `contract.DB` into your repositories or services rather than relying solely on the global `goe.DB()`.
*   **Repository Pattern**: Abstract database logic into repositories. Your services should call repository methods, not interact with GORM directly. This improves separation of concerns and testability.
*   **Separate Model Packages**: Organize your GORM models in dedicated packages (e.g., `internal/models`, `domain/entities`) for better code organization.

### Error Handling and Reliability

*   **Error Handling**: Always check for errors returned by GORM operations, including `gorm.ErrRecordNotFound`.
*   **Connection Validation**: Check if database instances are `nil` before using them, especially in early application lifecycle.
*   **Graceful Degradation**: Design your application to handle database connection failures gracefully.

### Performance and Optimization

*   **Connection Management**: Goe's DB module handles connection pooling and graceful shutdown. Ensure your `DB_*` pool settings are appropriate for your application's load.
*   **Query Optimization**: Use GORM's debugging features (`db.Debug()`) to inspect generated SQL. Write efficient queries and use database indexes.
*   **Batch Operations**: Use GORM's batch operations for bulk inserts/updates to improve performance.
*   **Preloading**: Use GORM's `Preload` feature to avoid N+1 query problems when loading related data.

### Security and Production Considerations

*   **Production Migrations**: For production, use dedicated migration tools for schema changes rather than auto-migration.
*   **Security**: Be cautious of SQL injection if constructing raw SQL queries. Prefer GORM's query-building methods, which generally handle sanitization.
*   **Environment-Specific Configuration**: Use different database configurations for development, testing, and production environments.
*   **Backup Strategy**: Ensure proper backup and recovery procedures are in place before running migrations in production.

### Example Repository Pattern with Model Registration

```go
// internal/models/user.go
package models

import "gorm.io/gorm"

type User struct {
    gorm.Model
    Name  string `gorm:"size:255;not null"`
    Email string `gorm:"size:255;uniqueIndex;not null"`
}

// internal/repository/user_repository.go
package repository

import (
    "go.oease.dev/goe/v2/contract"
    "yourproject/internal/models"
)

type UserRepository struct {
    db     contract.DB
    logger contract.Logger
}

func NewUserRepository(db contract.DB, logger contract.Logger) *UserRepository {
    return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository) Create(user *models.User) error {
    if err := r.db.Instance().Create(user).Error; err != nil {
        r.logger.Error("Failed to create user", "error", err)
        return err
    }
    return nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Instance().Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // User not found
        }
        r.logger.Error("Failed to find user by email", "email", email, "error", err)
        return nil, err
    }
    return &user, nil
}

// main.go
func main() {
    goe.New(goe.Options{
        WithDB: true,
        Providers: []any{
            repository.NewUserRepository,
        },
        Invokers: []any{
            func(db contract.DB) {
                // Register models for auto-migration
                db.RegisterModelsForMigration(&models.User{})
            },
            func(userRepo *repository.UserRepository) {
                // Use the repository in your application logic
            },
        },
    })
    goe.Run()
}
```

## Troubleshooting

### Common Issues and Solutions

#### "database connection 'default' not found or not configured"

This error typically occurs when trying to access the database before connections are established. 

**Solution**: Use the model registration system instead of manual migration:

```go
// ❌ WRONG - This can cause timing issues
func BadExample(db contract.DB) {
    // This might run before DB connections are established
    db.AutoMigrate(&models.User{})
}

// ✅ CORRECT - Register models for automatic migration
func GoodExample(db contract.DB) {
    // Register models - they'll be migrated when connections are ready
    db.RegisterModelsForMigration(&models.User{})
}
```

#### Models Not Being Migrated Automatically

If your registered models aren't being migrated:

1. **Check Configuration**: Ensure auto-migration is enabled:
   ```bash
   DB_AUTO_MIGRATE=true
   # or
   DB_AUTO_MIGRATE_ANY=true
   ```

2. **Verify Registration Timing**: Models must be registered in Invokers before `goe.Run()`:
   ```go
   goe.New(goe.Options{
       WithDB: true,
       Invokers: []any{
           func(db contract.DB) {
               db.RegisterModelsForMigration(&models.User{})
           },
       },
   })
   ```

3. **Check Logs**: Look for migration-related log messages during startup.

#### Connection Pool Issues

If you're experiencing connection pool exhaustion:

1. **Adjust Pool Settings**:
   ```bash
   DB_MAX_OPEN_CONNS=50
   DB_MAX_IDLE_CONNS=10
   DB_CONN_MAX_LIFETIME=1h
   DB_CONN_MAX_IDLE_TIME=5m
   ```

2. **Ensure Proper Connection Cleanup**: Always close transactions and don't hold connections longer than necessary.

#### Foreign Key Constraint Errors During Migration

If you encounter foreign key errors during auto-migration:

1. **Register Related Models Together**:
   ```go
   // Register parent models before child models
   db.RegisterModelsForMigration(
       &models.User{},      // Parent
       &models.Profile{},   // Child with foreign key to User
   )
   ```

2. **Disable Foreign Key Constraints During Migration** (if needed):
   ```bash
   DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING=true
   ```

### Performance Tips

#### Optimizing Database Queries

1. **Use Indexes**: Define appropriate indexes in your GORM models:
   ```go
   type User struct {
       gorm.Model
       Email string `gorm:"uniqueIndex"`
       Name  string `gorm:"index"`
   }
   ```

2. **Use Preloading for Relationships**:
   ```go
   // Load users with their profiles in a single query
   var users []models.User
   db.Preload("Profile").Find(&users)
   ```

3. **Use Select for Specific Fields**:
   ```go
   // Only load specific fields
   var users []models.User
   db.Select("id", "name", "email").Find(&users)
   ```

#### Batch Operations

For bulk operations, use GORM's batch features:

```go
// Batch insert
users := []models.User{
    {Name: "User1", Email: "user1@example.com"},
    {Name: "User2", Email: "user2@example.com"},
}
db.CreateInBatches(users, 100) // Insert in batches of 100
```

### Environment-Specific Configurations

#### Development Environment
```bash
DB_DRIVER=sqlite
DB_DATABASE=./dev.db
DB_AUTO_MIGRATE=true
DB_LOG_MODE=true  # Enable SQL logging for debugging
```

#### Testing Environment
```bash
DB_DRIVER=sqlite
DB_DATABASE=:memory:  # In-memory database for fast tests
DB_AUTO_MIGRATE=true
```

#### Production Environment
```bash
DB_DRIVER=postgres
DB_HOST=prod-db.example.com
DB_DATABASE=myapp_prod
DB_USERNAME=myapp_user
DB_PASSWORD=secure_password
DB_AUTO_MIGRATE=false  # Use dedicated migration tools in production
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=1h
```

Goe's database module, powered by GORM, offers a powerful yet convenient way to manage data persistence in your applications. The new model registration system eliminates common timing issues and provides a clean, configuration-driven approach to database migrations.

Next, let's look at [Caching Strategies](09-caching.md).
