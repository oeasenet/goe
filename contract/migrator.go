package contract

import (
	"context"
	"time"
)

// Migrator defines the interface for MongoDB migrations
type Migrator interface {
	// Up runs all pending migrations
	Up(ctx context.Context) (*MigrationResult, error)

	// UpTo runs migrations up to and including the specified version
	UpTo(ctx context.Context, version int64) (*MigrationResult, error)

	// Down rolls back N migrations
	Down(ctx context.Context, steps int) (*MigrationResult, error)

	// DownTo rolls back to a specific version (exclusive)
	DownTo(ctx context.Context, version int64) (*MigrationResult, error)

	// Reset rolls back all applied migrations
	Reset(ctx context.Context) (*MigrationResult, error)

	// Redo rolls back then re-applies the last N migrations
	Redo(ctx context.Context, steps int) (*MigrationResult, error)

	// Status returns the status of all migrations
	Status(ctx context.Context) ([]MigrationStatusEntry, error)

	// Pending returns migrations that haven't been applied yet
	Pending(ctx context.Context) ([]MigrationInfo, error)

	// Version returns the current schema version
	Version(ctx context.Context) (int64, error)

	// DryRun simulates migrations without applying
	DryRun(ctx context.Context) (*MigrationResult, error)

	// VerifyChecksums verifies all applied migrations' checksums
	VerifyChecksums(ctx context.Context) error

	// ClearDirty clears the dirty state for manual intervention
	ClearDirty(ctx context.Context, version int64) error
}

// MigrationResult contains the result of a migration operation
type MigrationResult struct {
	// Applied contains the versions that were successfully applied
	Applied []int64

	// RolledBack contains the versions that were rolled back
	RolledBack []int64

	// Skipped contains the versions that were skipped
	Skipped []int64

	// Duration is the total time taken for the operation
	Duration time.Duration

	// DryRun indicates if this was a dry run
	DryRun bool
}

// MigrationStatusEntry represents the status of a single migration
type MigrationStatusEntry struct {
	Version   int64
	Name      string
	Status    string // "applied", "pending", "dirty"
	AppliedAt *time.Time
	Batch     int
}

// MigrationInfo provides basic information about a migration
type MigrationInfo struct {
	Version int64
	Name    string
}
