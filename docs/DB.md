# Database Module (GORM)

The Goe framework includes a powerful and flexible database module built upon [GORM](https://gorm.io/), allowing developers to interact with various SQL databases in a Go-idiomatic way. This module is designed to be configuration-driven and easily accessible throughout your application.

## Enabling the DB Module

To use the database module, you need to enable it when initializing your Goe application:

```go
package main

import (
	"go.oease.dev/goe/v2"
	// other imports
)

func main() {
	app := goe.New(goe.Options{
		WithDB: true, // Enable the DB module
		// ... other options
	})

	// ... start your application
	goe.Run()
}
```

## Configuration

Database connections are configured via environment variables. The module supports a default connection and can be extended to support multiple named connections (though automatic startup of multiple named connections requires further enhancement in the `OnStart` method).

### Default Connection

The primary connection is configured using `DB_*` prefixed variables. If `DB_CONNECTION` is not set, it defaults to a connection named `"default"`.

**Required for most drivers:**

*   `DB_DRIVER`: The type of database driver. Supported values:
    *   `mysql`
    *   `postgres` (or `pgsql`, `postgresql`)
    *   `sqlite` (or `sqlite3`)
    *   `sqlserver` (or `mssql`)
*   `DB_HOST`: Database server host (e.g., `localhost`, `127.0.0.1`).
*   `DB_DATABASE`: Database name. For SQLite, this is the path to the database file (e.g., `mydatabase.db`). If not specified for SQLite, it defaults to an in-memory database (`:memory:`), which is useful for testing.
*   `DB_USERNAME`: Database username.
*   `DB_PASSWORD`: Database password.

**Optional / Driver-Specific:**

*   `DB_PORT`: Database server port (e.g., `3306` for MySQL, `5432` for PostgreSQL). Defaults to standard ports if not set.
*   `DB_DSN`: A full Data Source Name (DSN) string. If provided, this will be used directly, and individual components like host, port, etc., for the default connection might be ignored.
*   `DB_CHARSET`: (MySQL) Character set (e.g., `utf8mb4`). Defaults to `utf8mb4`.
*   `DB_TIMEZONE`: (MySQL, PostgreSQL) Timezone for the connection (e.g., `UTC`, `Local`). Defaults to `Local` for MySQL, `UTC` for PostgreSQL.
*   `DB_SSLMODE`: (PostgreSQL) SSL mode (e.g., `disable`, `require`). Defaults to `disable`.
*   `DB_ENCRYPT`: (SQLServer) Encryption mode (e.g., `disable`, `true`). Defaults to `disable`.
*   `DB_TRUST_SERVER_CERTIFICATE`: (SQLServer) Whether to trust the server certificate (e.g., `true`, `false`).

**Connection Pool:**

*   `DB_MAX_IDLE_CONNS`: Maximum number of connections in the idle connection pool. (Default: 10)
*   `DB_MAX_OPEN_CONNS`: Maximum number of open connections to the database. (Default: 100)
*   `DB_CONN_MAX_LIFETIME`: Maximum amount of time a connection may be reused (e.g., `1h`, `30m`).
*   `DB_CONN_MAX_IDLE_TIME`: Maximum amount of time a connection may be idle before being closed (e.g., `5m`).

**Logging & GORM Behavior:**

*   `DB_LOG_MODE`: Set to `true` to enable GORM's logger (prints SQL queries). (Default: `false`)
*   `DB_IGNORE_RECORD_NOT_FOUND_ERROR`: Set to `true` to make GORM treat `ErrRecordNotFound` as a normal empty result, not an error. (Default: `false`)
*   `DB_DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING`: Set to `true` to disable foreign key constraints during GORM's auto-migration. (Default: `false`)
*   `DB_AUTO_MIGRATE`: If set to `true`, the module will log that auto-migration is enabled. Actual migration of specific models needs to be triggered explicitly (see [Auto Migration](#auto-migration)).

### Named Connections

You can configure multiple database connections by prefixing the configuration keys with `DB_<NAME>_`. For example, for a connection named `reporting`:

*   `DB_REPORTING_DRIVER=postgres`
*   `DB_REPORTING_HOST=reports.example.com`
*   ...and so on.

To make the application use a named connection as the *default* connection (the one returned by `goe.DB().Instance()`), set the `DB_CONNECTION` variable:

*   `DB_CONNECTION=reporting`

**Note:** Currently, the `DatabaseModule`'s `OnStart` method only automatically establishes the connection specified by `DB_CONNECTION` (or "default"). To use other named connections, they would need to be connected programmatically after startup, or the `OnStart` method would need to be enhanced to parse a list of connections to pre-connect (e.g., from a `DB_CONNECTIONS` environment variable).

## Usage

### Accessing the DB Instance

You can get the default GORM database instance anywhere in your application using the global accessor:

```go
import "go.oease.dev/goe/v2"

func GetDB() *gorm.DB {
	return goe.DB().Instance()
}

// To access a named connection (if configured and connected):
func GetNamedDB(name string) (*gorm.DB, error) {
    return goe.DB().Connection(name)
}
```

### Defining GORM Models

Define your GORM models as you normally would.

```go
package models // or your preferred package

import "gorm.io/gorm"

type User struct {
	gorm.Model // Includes ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string
	Email      string `gorm:"uniqueIndex"`
	Age        uint8
}

type Product struct {
	ID     uint `gorm:"primaryKey"`
	Code   string
	Price  uint
	UserID uint   // Foreign key for User
	User   User   // Belongs to User
}
```

### Auto Migration

While `DB_AUTO_MIGRATE=true` enables the *possibility* of auto-migration, you still need to tell GORM which models to migrate. This is typically done once, often during application startup or via a dedicated migration command.

The `DatabaseModule` provides an `AutoMigrate` helper method. You can access the underlying `*db.DatabaseModule` instance via dependency injection if you need to call its methods directly, or ideally, wrap this in a service.

```go
// Example of how you might trigger migration after app setup
// (This assumes you have a way to get the concrete *db.DatabaseModule instance,
// or this functionality is exposed through contract.DB or a dedicated migration service)

// In your main setup or a specific invoker:
// var dbService *db.DatabaseModule // Injected by Fx
// err := goe.App().Container().Invoke(func(d contract.DB) {
//     // If contract.DB is the concrete *db.DatabaseModule type or wraps AutoMigrate:
//     // d.AutoMigrate(&models.User{}, &models.Product{})
//     // Otherwise, you might need to resolve the concrete module.
// })

// A more direct way using the default instance:
dbInstance := goe.DB().Instance()
if dbInstance != nil {
    err := dbInstance.AutoMigrate(&models.User{}, &models.Product{})
    if err != nil {
        goe.Log().Error("Failed to auto-migrate database tables", contract.NewField("error", err))
    } else {
        goe.Log().Info("Database auto-migration successful for User and Product tables.")
    }
}
```

**Note:** For production environments, consider using dedicated migration tools (like GORM's migrator, Goose, Atlas, etc.) for more control over schema changes.

### CRUD Operations

Perform CRUD operations using standard GORM syntax:

```go
import (
	"go.oease.dev/goe/v2"
	"your_project/models" // Assuming your models are here
	"gorm.io/gorm"
)

func CreateUser(name, email string, age uint8) (*models.User, error) {
	db := goe.DB().Instance()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	user := models.User{Name: name, Email: email, Age: age}
	result := db.Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func GetUserByID(id uint) (*models.User, error) {
	db := goe.DB().Instance()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var user models.User
	result := db.First(&user, id) // Find user with integer primary key
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Or a custom "not found" error
		}
		return nil, result.Error
	}
	return &user, nil
}

func UpdateUserEmail(id uint, newEmail string) error {
	db := goe.DB().Instance()
	if db == nil {
		return errors.New("database not initialized")
	}
	// Update attributes with map or struct
	result := db.Model(&models.User{}).Where("id = ?", id).Update("email", newEmail)
	return result.Error
}

func DeleteUser(id uint) error {
	db := goe.DB().Instance()
	if db == nil {
		return errors.New("database not initialized")
	}
	result := db.Delete(&models.User{}, id)
	return result.Error
}
```

## Advanced GORM Features

The `*gorm.DB` instance obtained from `goe.DB().Instance()` is a standard GORM DB object, so you can use all of GORM's features:
*   Transactions
*   Associations (Has One, Belongs To, Has Many, Many To Many)
*   Scopes
*   Raw SQL / SQL Builder
*   Hooks
*   And more.

Refer to the [official GORM documentation](https://gorm.io/docs/) for detailed information on these features.

## Testing

For testing, it's common to use an in-memory SQLite database:

```env
# .env.test or set these before running tests
DB_DRIVER=sqlite
DB_DATABASE=:memory:
DB_LOG_MODE=false
# DB_AUTO_MIGRATE=true (if your tests rely on it, but usually explicit migration in tests is better)
```

Then, in your tests, ensure `goe.Options{WithDB: true}` is set. You can perform migrations for your test models in a setup function. See `tests/db_test.go` for examples.
```
