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

GORM can automatically create or update database tables based on your model definitions. Goe's `contract.DB` interface exposes `AutoMigrate` methods.

*   `AutoMigrate(dst ...interface{}) error`: Operates on the **default** connection.
*   `AutoMigrateOnConnection(connectionName string, dst ...interface{}) error`: Operates on a **named** connection.

**How to use:**

This is typically done once during application startup, often within an Fx invoker.

```go
package main

import (
    // ... other imports
    "go.oease.dev/goe/v2/contract"
    "example.com/yourproject/internal/models" // Your models package
)

// ... main function with goe.New, including WithDB: true ...
// Add an invoker for migrations:
// Invokers: []any{RunMigrations},

// RunMigrations is an Fx invoker
func RunMigrations(db contract.DB, logger contract.Logger) error {
    logger.Info("Running database auto-migrations...")

    // Migrate models on the default connection
    err := db.AutoMigrate(
        &models.User{},
        &models.Product{},
        // ... other models for the default DB
    )
    if err != nil {
        logger.Error("Failed to auto-migrate default database tables", contract.NewField("error", err))
        return err // Returning an error will stop app startup if this is critical
    }
    logger.Info("Default database auto-migration successful.")

    // Example: Migrate models on a named connection "reporting_db"
    // Make sure "reporting_db" is configured and connected via DB_CONNECTIONS
    /*
    err = db.AutoMigrateOnConnection("reporting_db", &models.ReportSummary{})
    if err != nil {
        logger.Error("Failed to auto-migrate reporting_db tables", contract.NewField("error", err))
        return err
    }
    logger.Info("Reporting_db auto-migration successful.")
    */

    return nil
}
```

**Important Notes on Auto Migration:**

*   **Development vs. Production**: GORM's auto-migration is very convenient for development and testing. However, for production environments, it's generally safer and more controllable to use dedicated migration tools (like Goose, Atlas, Flyway, Liquibase, or GORM's own migrator tool). These tools offer versioning, rollbacks, and more fine-grained control over schema changes.
*   **Limitations**: Auto-migration might not handle all complex schema changes perfectly (e.g., renaming columns, changing column types with data preservation). Always test schema changes thoroughly.
*   **Configuration Hint**: Setting `DB_AUTO_MIGRATE=true` (or `DB_<NAME>_AUTO_MIGRATE=true`) primarily serves as a configuration hint that logs that this feature is enabled. The actual migration still needs to be triggered via code as shown above.

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

*   **Use Dependency Injection**: Inject `contract.DB` into your repositories or services rather than relying solely on the global `goe.DB()`.
*   **Repository Pattern**: Abstract database logic into repositories. Your services should call repository methods, not interact with GORM directly. This improves separation of concerns and testability.
*   **Error Handling**: Always check for errors returned by GORM operations, including `gorm.ErrRecordNotFound`.
*   **Connection Management**: Goe's DB module handles connection pooling and graceful shutdown. Ensure your `DB_*` pool settings are appropriate for your application's load.
*   **Query Optimization**: Use GORM's debugging features (`db.Debug()`) to inspect generated SQL. Write efficient queries and use database indexes.
*   **Production Migrations**: For production, use dedicated migration tools for schema changes.
*   **Security**: Be cautious of SQL injection if constructing raw SQL queries. Prefer GORM's query-building methods, which generally handle sanitization.

Goe's database module, powered by GORM, offers a powerful yet convenient way to manage data persistence in your applications.

Next, let's look at [Caching Strategies](09-caching.md).
```
