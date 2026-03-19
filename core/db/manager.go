package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.oease.dev/goe/v2/contract"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// connect initializes a GORM DB connection based on the provided configuration prefix.
// The configuration keys are expected to be like:
// DB_DRIVER, DB_HOST, DB_PORT, DB_DATABASE, DB_USERNAME, DB_PASSWORD
// For a named connection "foo", the keys would be:
// DB_FOO_DRIVER, DB_FOO_HOST, etc.
func (dbm *DatabaseModule) connect(name string) (*gorm.DB, error) {
	configPrefix := "DB_"
	if name != "default" && name != "" {
		configPrefix = fmt.Sprintf("DB_%s_", strings.ToUpper(name))
	}

	driver := dbm.config.GetString(configPrefix + "DRIVER")
	if driver == "" {
		// If no specific driver for a named connection, try falling back to the default driver config.
		// This might be useful if only the DSN parts change but the driver type is the same.
		// However, for clarity, it's better if each named connection specifies its driver.
		// For now, if DB_FOO_DRIVER is not set, it's an error.
		return nil, fmt.Errorf("database driver not configured for connection '%s' (expected key %sDRIVER)", name, configPrefix)
	}

	dsn, err := dbm.buildDSN(name, driver, configPrefix)
	if err != nil {
		return nil, err
	}

	// GORM logger configuration - use Warn level to reduce verbosity
	// Detailed SQL logs will still appear at debug level through Trace method
	gormLogLevel := gormlogger.Silent
	if dbm.config.GetBool(configPrefix+"LOG_MODE") || dbm.config.GetBool("DB_LOG_MODE") { // Allow global and per-connection log mode
		gormLogLevel = gormlogger.Warn
	}

	newLogger := gormlogger.New(
		// Use goe logger for gorm messages
		NewGoeGormLogger(dbm.logger),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond, // Can be made configurable
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: dbm.config.GetBool(configPrefix + "IGNORE_RECORD_NOT_FOUND_ERROR"), // Default false
			Colorful:                  false,                                                              // Usually true for dev, false for prod. Let's keep it false for structured logging.
		},
	)

	gormConfig := &gorm.Config{
		Logger: newLogger,
		// Add more GORM configs here as needed, e.g., NamingStrategy, DryRun, etc.
		// Example:
		// NamingStrategy: schema.NamingStrategy{
		// SingularTable: dbm.config.GetBool(configPrefix + "SINGULAR_TABLE"),
		// },
		DisableForeignKeyConstraintWhenMigrating: dbm.config.GetBool(configPrefix + "DISABLE_FOREIGN_KEY_CONSTRAINT_WHEN_MIGRATING"), // Default false
	}

	var dialector gorm.Dialector
	switch strings.ToLower(driver) {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "pgsql", "postgres", "postgresql":
		dialector = postgres.Open(dsn)
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(dsn) // DSN is usually just the file path for SQLite
	case "sqlserver", "mssql":
		dialector = sqlserver.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s for connection '%s'", driver, name)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		dbm.logger.Error("Failed to connect to database",
			"connection", name,
			"driver", driver,
			"error", err.Error(), // Log only error message, not the full error struct
		)
		return nil, fmt.Errorf("failed to connect to %s database '%s': %w", driver, name, err)
	}

	// Configure connection pool (can be made configurable)
	sqlDB, err := db.DB()
	if err != nil {
		dbm.logger.Error("Failed to get underlying sql.DB for connection pool setup", "connection", name, "error", err)
		// Not returning error here, as the connection itself was successful. Log and proceed.
	} else {
		maxIdleConns := dbm.config.GetInt(configPrefix + "MAX_IDLE_CONNS")
		if maxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(maxIdleConns)
		} else if dbm.config.Has(configPrefix + "MAX_IDLE_CONNS") { // Check if explicitly set to 0 or less
			// use default from database/sql package
		} else {
			sqlDB.SetMaxIdleConns(10) // Default
		}

		maxOpenConns := dbm.config.GetInt(configPrefix + "MAX_OPEN_CONNS")
		if maxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(maxOpenConns)
		} else if dbm.config.Has(configPrefix + "MAX_OPEN_CONNS") {
			// use default
		} else {
			sqlDB.SetMaxOpenConns(100) // Default
		}

		connMaxLifetime := dbm.config.GetDuration(configPrefix + "CONN_MAX_LIFETIME")
		if connMaxLifetime > 0 {
			sqlDB.SetConnMaxLifetime(connMaxLifetime)
		}

		connMaxIdleTime := dbm.config.GetDuration(configPrefix + "CONN_MAX_IDLE_TIME")
		if connMaxIdleTime > 0 {
			sqlDB.SetConnMaxIdleTime(connMaxIdleTime)
		}
	}

	dbm.logger.Info("Database connection established successfully", "connection", name, "driver", driver)
	return db, nil
}

// buildDSN constructs the DSN string based on the driver and configuration.
func (dbm *DatabaseModule) buildDSN(name, driver, configPrefix string) (string, error) {
	// Allow providing a full DSN directly
	directDSN := dbm.config.GetString(configPrefix + "DSN")
	if directDSN != "" {
		dbm.logger.Info("Using direct DSN for connection", "connection", name)
		return directDSN, nil
	}

	// Standard DSN components
	host := dbm.config.GetString(configPrefix + "HOST")
	port := dbm.config.GetString(configPrefix + "PORT") // GetString because port can be non-numeric for sockets etc.
	dbname := dbm.config.GetString(configPrefix + "DATABASE")
	user := dbm.config.GetString(configPrefix + "USERNAME")
	pass := dbm.config.GetString(configPrefix + "PASSWORD")
	charset := dbm.config.GetString(configPrefix + "CHARSET")
	timezone := dbm.config.GetString(configPrefix + "TIMEZONE") // URL encode this if needed

	// SQLite specific: dbname is the file path
	if strings.ToLower(driver) == "sqlite" || strings.ToLower(driver) == "sqlite3" {
		if dbname == "" {
			// Default to an in-memory database if no path is provided, common for testing
			// Or you could make this an error: return "", fmt.Errorf("database path (DB_DATABASE or DB_%s_DATABASE) not configured for SQLite connection '%s'", strings.ToUpper(name), name)
			dbm.logger.Info("SQLite database path not specified, using in-memory database.", "connection", name)
			return ":memory:", nil
		}
		// TODO: Add support for query params for SQLite if needed, e.g., "file:path?cache=shared&mode=memory"
		return dbname, nil
	}

	// Check for required fields for other drivers
	if host == "" {
		return "", fmt.Errorf("database host not configured for connection '%s' (expected key %sHOST)", name, configPrefix)
	}
	// dbname, user might not be strictly required for all DSN formats or scenarios, but are typical.

	switch strings.ToLower(driver) {
	case "mysql":
		// dsn := "user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		if charset == "" {
			charset = "utf8mb4"
		}
		if port == "" {
			port = "3306"
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True", user, pass, host, port, dbname, charset)
		if timezone != "" {
			// MySQL's loc parameter needs to be URL encoded if it contains special characters.
			// For simplicity, assuming common values like "Local" or "UTC" don't need encoding here.
			// Consider using net/url.QueryEscape for robust timezone handling.
			dsn = fmt.Sprintf("%s&loc=%s", dsn, timezone)
		} else {
			dsn = fmt.Sprintf("%s&loc=Local", dsn) // GORM default
		}
		return dsn, nil
	case "pgsql", "postgres", "postgresql":
		// dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/Shanghai"
		if port == "" {
			port = "5432"
		}
		sslmode := dbm.config.GetString(configPrefix + "SSLMODE")
		if sslmode == "" {
			sslmode = "disable" // Default, adjust as needed for production
		}
		if timezone == "" {
			timezone = "UTC" // Default
		}
		// Note: password might need to be escaped if it contains spaces or special characters.
		// GORM pq driver handles this, but good to be aware.
		return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s TimeZone=%s",
			host, port, user, dbname, pass, sslmode, timezone), nil
	case "sqlserver", "mssql":
		// dsn := "sqlserver://username:password@host:port?database=dbname&encrypt=disable" // encrypt=disable for dev, true for prod with proper certs
		if port == "" {
			port = "1433"
		}
		encrypt := dbm.config.GetString(configPrefix + "ENCRYPT")
		if encrypt == "" {
			encrypt = "disable" // Default for dev
		}
		// TrustServerCertificate might be needed for self-signed certs: &trustservercertificate=true
		trustCert := dbm.config.GetString(configPrefix + "TRUST_SERVER_CERTIFICATE")
		connString := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=%s",
			user, pass, host, port, dbname, encrypt)
		if trustCert != "" {
			connString += "&trustservercertificate=" + trustCert
		}
		return connString, nil
	default:
		return "", fmt.Errorf("cannot build DSN for unsupported driver: %s for connection '%s'", driver, name)
	}
}

// --- GORM Logger Wrapper ---

// GoeGormLogger wraps goe.Logger to be used as a gorm.logger.Interface
// All GORM logs are converted to use GOE's structured logging format with "module": "gorm"
type GoeGormLogger struct {
	goeLogger contract.Logger
}

func (l *GoeGormLogger) Printf(s string, i ...any) {
	var message string
	if i != nil {
		message = fmt.Sprintf(s, i...)
	} else {
		message = s
	}

	// Parse GORM's formatted log messages and convert to structured logging
	if strings.Contains(message, "SLOW SQL") {
		// Extract slow query information and log with structured format
		l.goeLogger.Debug("GORM slow query detected",
			"module", "gorm",
			"message", message,
		)
		return
	}

	// For other GORM log messages, use structured logging
	l.goeLogger.Debug("GORM log",
		"module", "gorm",
		"message", message,
	)
}

// NewGoeGormLogger creates a new GoeGormLogger
func NewGoeGormLogger(logger contract.Logger) *GoeGormLogger {
	return &GoeGormLogger{goeLogger: logger}
}

// LogMode sets log level (not directly used here as GORM's own config controls it)
func (l *GoeGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	// GORM will pass its configured level. We can create a new logger if we want to respect it,
	// but for now, our goeLogger's level is independent.
	return l
}

// Info prints info messages
func (l *GoeGormLogger) Info(ctx context.Context, msg string, data ...any) {
	args := []any{"module", "gorm"}
	args = append(args, convertGormLogDataToArgs(data)...)
	l.goeLogger.Debug(msg, args...)
}

// Warn prints warning messages
func (l *GoeGormLogger) Warn(ctx context.Context, msg string, data ...any) {
	args := []any{"module", "gorm"}
	args = append(args, convertGormLogDataToArgs(data)...)
	l.goeLogger.Warn(msg, args...)
}

// Error prints error messages
func (l *GoeGormLogger) Error(ctx context.Context, msg string, data ...any) {
	args := []any{"module", "gorm"}
	args = append(args, convertGormLogDataToArgs(data)...)
	l.goeLogger.Error(msg, args...)
}

// Trace prints SQL query execution information
func (l *GoeGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	elapsedMs := float64(elapsed.Nanoseconds()) / 1e6

	if err != nil && err != gorm.ErrRecordNotFound { // Don't log RecordNotFound as an error from Trace, GORM handles it.
		l.goeLogger.Error("GORM SQL error",
			"module", "gorm",
			"elapsed_ms", fmt.Sprintf("%.3f", elapsedMs),
			"sql", sql,
			"rows", rows,
			"error", err.Error(),
		)
		return
	}

	// Log slow queries at warn level (threshold from GORM config is 200ms)
	slowThreshold := 200 * time.Millisecond
	if elapsed > slowThreshold {
		l.goeLogger.Debug("GORM slow query",
			"module", "gorm",
			"elapsed_ms", fmt.Sprintf("%.3f", elapsedMs),
			"sql", sql,
			"rows", rows,
		)
		return
	}

	// Regular queries at debug level
	l.goeLogger.Debug("GORM SQL query",
		"module", "gorm",
		"elapsed_ms", fmt.Sprintf("%.3f", elapsedMs),
		"sql", sql,
		"rows", rows,
	)
}

// convertGormLogDataToArgs converts GORM's variadic data to key-value pairs
func convertGormLogDataToArgs(data []any) []any {
	args := make([]any, 0, len(data))
	for i := 0; i < len(data); i += 2 {
		key, ok := data[i].(string)
		if !ok {
			continue // Should not happen with GORM's internal logging format
		}
		args = append(args, key)
		if i+1 < len(data) {
			args = append(args, data[i+1])
		} else {
			args = append(args, nil)
		}
	}
	return args
}

// Update DatabaseModule's OnStart to use the connect method
// This needs to be done by modifying core/db/db.go after this subtask.
// For now, this file only defines the manager logic.
