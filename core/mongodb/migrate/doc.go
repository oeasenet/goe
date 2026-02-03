// Package migrate provides MongoDB-specific database migrations for the GOE framework.
//
// This package offers a fluent, developer-friendly API for managing MongoDB schema
// changes, including indexes, collections, fields, and document-level versioning.
//
// # Features
//
//   - Schema-changing migrations (indexes, collections, field operations)
//   - Automatic document-level versioning via _goe_sv field
//   - Distributed locking for safe concurrent deployments
//   - Fluent builder API with powerful helpers
//   - Checksum verification to detect modified migrations
//   - Transaction support for replica sets
//   - Dry run mode for testing migrations
//
// # Quick Start
//
// Register migrations in an init function:
//
//	package migrations
//
//	import (
//	    "context"
//
//	    "go.mongodb.org/mongo-driver/v2/mongo"
//	    "go.oease.dev/goe/v2/core/mongodb/migrate"
//	)
//
//	func init() {
//	    migrate.MustRegister(
//	        // Custom migration function
//	        migrate.New(1, "create_users_indexes").
//	            Up(func(ctx context.Context, db *mongo.Database) error {
//	                _, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
//	                    Keys: bson.D{{Key: "email", Value: 1}},
//	                })
//	                return err
//	            }).
//	            Down(func(ctx context.Context, db *mongo.Database) error {
//	                return db.Collection("users").Indexes().DropOne(ctx, "email_1")
//	            }),
//
//	        // Using helper functions (one-liners)
//	        migrate.New(2, "add_email_unique").
//	            Up(migrate.CreateIndex("users", "email", migrate.Unique())).
//	            Down(migrate.DropIndex("users", "email_1")),
//
//	        // Schema-changing helper (auto-updates _goe_sv)
//	        migrate.New(3, "add_user_status").
//	            Up(migrate.AddFieldWithVersion("users", "status", "active", 3)).
//	            Down(migrate.RemoveField("users", "status")),
//	    )
//	}
//
// # GOE Integration
//
// Enable migrations in your GOE application:
//
//	app := goe.New(goe.Options{
//	    WithMongoDB: true,
//	    WithMigrate: true,
//	})
//
// Access the migrator:
//
//	migrator := goe.Migrate()
//	result, err := migrator.Up(ctx)
//
// # Configuration
//
// Configure via environment variables:
//
//	MONGODB_MIGRATE_COLLECTION=_goe_migrations   # State collection name (default: _goe_migrations)
//	MONGODB_MIGRATE_TIMEOUT=5m                   # Per-migration timeout (default: 5m)
//	MONGODB_MIGRATE_LOCK_TIMEOUT=30m             # Distributed lock timeout (default: 30m)
//	MONGODB_MIGRATE_LOCK_HEARTBEAT=30s           # Lock heartbeat interval (default: 30s)
//	MONGODB_MIGRATE_USE_TRANSACTIONS=true        # Use transactions if replica set available (default: true)
//	MONGODB_MIGRATE_VERIFY_CHECKSUMS=true        # Verify checksums on startup (default: true)
//	MONGODB_MIGRATE_AUTO=false                   # Auto-run pending migrations on app start (default: false)
//	MONGODB_MIGRATE_VERSION_SCHEME=sequential    # Version scheme: sequential or timestamp (default: sequential)
//	MONGODB_MIGRATE_SCHEMA_VERSION_FIELD=_goe_sv # Document schema version field name (default: _goe_sv)
//	MONGODB_MIGRATE_DRY_RUN=false                # Enable dry-run mode by default (default: false)
//
// # Migration Helpers
//
// The package provides many helper functions for common operations:
//
// Index Operations:
//
//	CreateIndex(collection, field, opts...)       // Create single-field index
//	CreateCompoundIndex(collection, fields, opts...) // Create compound index
//	CreateUniqueIndex(collection, field)          // Create unique index
//	CreateTTLIndex(collection, field, expireAfter) // Create TTL index
//	DropIndex(collection, indexName)              // Drop index by name
//
// Collection Operations:
//
//	CreateCollection(name, opts...)   // Create collection
//	DropCollection(name)              // Drop collection
//	RenameCollection(old, new)        // Rename collection
//	SetValidation(collection, schema) // Set JSON schema validation
//
// Field Operations (auto-update _goe_sv):
//
//	AddField(collection, field, default)     // Add field with default
//	AddFieldWithVersion(collection, field, default, version) // Add field with explicit version
//	RemoveField(collection, field)           // Remove field
//	RenameField(collection, old, new)        // Rename field
//	ConvertFieldType(collection, field, transformer) // Convert field type
//
// Schema Versioning:
//
//	SetSchemaVersion(collection, version, filter) // Set _goe_sv explicitly
//	BumpSchemaVersion(collection, from, to, transformer) // Transform docs v1→v2
//
// Utilities:
//
//	Sequence(fn1, fn2, ...)  // Run multiple ops in sequence
//	NoOp()                   // No operation (for down-only)
//	Log(message)             // Log a message (for debugging)
//
// # Index Options
//
//	Unique()           // Create unique index
//	Sparse()           // Create sparse index
//	ExpireAfter(d)     // Create TTL index
//	Name(n)            // Set custom index name
//	PartialFilter(f)   // Set partial filter expression
//	NoVersionBump()    // Opt-out of automatic _goe_sv update
//	WithSchemaVersion(v) // Set explicit schema version
//
// # Document-Level Schema Versioning
//
// The package supports document-level schema versioning through the _goe_sv field.
// This enables lazy migrations (transform on read if _goe_sv < currentVersion).
//
// When using schema-changing helpers like AddField, RemoveField, or RenameField,
// the _goe_sv field is automatically updated on affected documents.
//
// To opt-out of automatic versioning:
//
//	migrate.New(3, "backfill_legacy").
//	    Up(migrate.AddField("users", "legacy_flag", true, migrate.NoVersionBump()))
//
// # Migrator Methods
//
//	Up(ctx)             // Run all pending migrations
//	UpTo(ctx, version)  // Run up to specific version
//	Down(ctx, steps)    // Rollback N migrations
//	DownTo(ctx, version) // Rollback to specific version
//	Reset(ctx)          // Rollback all migrations
//	Redo(ctx, steps)    // Rollback then re-apply N migrations
//	Status(ctx)         // Get migration status
//	Pending(ctx)        // Get pending migrations
//	Version(ctx)        // Get current version
//	DryRun(ctx)         // Preview without applying
//
// # Safety Features
//
//   - Distributed Locking: MongoDB atomic findOneAndUpdate with heartbeat
//   - Checksum Verification: Detect modified migration functions
//   - Transaction Support: Wrap in transaction when replica set available
//   - Dirty State Handling: Detect and require manual intervention for failed migrations
//   - Timeout per Migration: Configurable, default 5 minutes
//
// # Error Handling
//
// The package provides specific error types for errors.Is/As:
//
//	ErrMigrationNotFound     // Migration version not found
//	ErrDuplicateVersion      // Duplicate migration version
//	ErrDirtyState            // Database in dirty state
//	ErrLockAcquisitionFailed // Failed to acquire lock
//	ErrLockLost              // Lock was lost
//	ErrChecksumMismatch      // Checksum mismatch
//	ErrMigrationTimeout      // Migration timed out
//
// For detailed error information, use type assertions:
//
//	if migErr, ok := err.(*migrate.MigrationError); ok {
//	    fmt.Printf("Migration %d (%s) failed: %v\n", migErr.Version, migErr.Name, migErr.Cause)
//	}
package migrate
