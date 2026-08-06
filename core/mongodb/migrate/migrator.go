package migrate

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/contract"
)

// Type aliases for contract types to ensure interface compatibility
type (
	// MigrationResult is an alias for contract.MigrationResult
	MigrationResult = contract.MigrationResult

	// MigrationStatusEntry is an alias for contract.MigrationStatusEntry
	MigrationStatusEntry = contract.MigrationStatusEntry

	// MigrationInfo is an alias for contract.MigrationInfo
	MigrationInfo = contract.MigrationInfo
)

// MigratorOption configures the Migrator
type MigratorOption func(*Migrator)

// WithLogger sets a custom logger
func WithLogger(logger contract.Logger) MigratorOption {
	return func(m *Migrator) {
		m.logger = logger
	}
}

// WithConfig sets custom configuration
func WithConfig(config *Config) MigratorOption {
	return func(m *Migrator) {
		m.config = config
	}
}

// WithDryRun enables dry run mode
func WithDryRun(enabled bool) MigratorOption {
	return func(m *Migrator) {
		m.dryRun = enabled
	}
}

// WithHostname sets a custom hostname for lock acquisition
func WithHostname(hostname string) MigratorOption {
	return func(m *Migrator) {
		m.hostname = hostname
	}
}

// Migrator is the main migration executor
type Migrator struct {
	db       *mongo.Database
	config   *Config
	logger   contract.Logger
	state    *MongoStateStore
	lock     *MongoLock
	hostname string
	dryRun   bool
}

// Compile-time check that Migrator implements contract.Migrator
var _ contract.Migrator = (*Migrator)(nil)

// NewMigrator creates a new Migrator instance
// Panics if db is nil to fail fast during initialization
func NewMigrator(db *mongo.Database, opts ...MigratorOption) *Migrator {
	if db == nil {
		panic("migrate: database cannot be nil")
	}

	m := &Migrator{
		db:     db,
		config: DefaultConfig(),
	}

	// Apply options
	for _, opt := range opts {
		opt(m)
	}

	// Create state store
	m.state = NewMongoStateStore(db, m.config.Collection, m.hostname)

	// Create lock
	lockCollection := m.config.Collection + "_lock"
	m.lock = NewMongoLock(db, lockCollection, m.config.LockTimeout, m.config.LockHeartbeat)

	return m
}

// Init initializes the migrator (creates collections, indexes, etc.)
func (m *Migrator) Init(ctx context.Context) error {
	if err := m.state.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize state store: %w", err)
	}

	if err := m.lock.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize lock: %w", err)
	}

	return nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) (*MigrationResult, error) {
	return m.upTo(ctx, 0)
}

// UpTo runs migrations up to and including the specified version
func (m *Migrator) UpTo(ctx context.Context, version int64) (*MigrationResult, error) {
	return m.upTo(ctx, version)
}

// upTo is the internal implementation of Up/UpTo
func (m *Migrator) upTo(ctx context.Context, targetVersion int64) (*MigrationResult, error) {
	start := time.Now()
	result := &MigrationResult{DryRun: m.dryRun}

	// Acquire lock
	if !m.dryRun {
		if err := m.lock.Acquire(ctx); err != nil {
			return nil, err
		}
		defer func() { _ = m.lock.Release(ctx) }()
	}

	// Check for dirty state
	dirty, err := m.state.GetDirty(ctx)
	if err != nil {
		return nil, err
	}
	if dirty != nil {
		return nil, NewDirtyStateError(dirty.Version, dirty.Name, dirty.Error, dirty.AppliedAt.Format(time.RFC3339))
	}

	// Get applied versions
	appliedVersions, err := m.state.GetAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	// Get migrations to run
	migrations := GetMigrations()
	if len(migrations) == 0 {
		result.Duration = time.Since(start)
		return result, nil
	}

	// Get next batch number
	batch, err := m.state.GetNextBatch(ctx)
	if err != nil {
		return nil, err
	}

	// Filter migrations to run
	var toRun []*Migration
	for _, mig := range migrations {
		// Skip if already applied
		if appliedVersions[mig.Version] {
			result.Skipped = append(result.Skipped, mig.Version)
			continue
		}

		// Skip if beyond target version
		if targetVersion > 0 && mig.Version > targetVersion {
			continue
		}

		toRun = append(toRun, mig)
	}

	// Run migrations
	for _, mig := range toRun {
		m.log("Running migration", "version", mig.Version, "name", mig.Name, "direction", "up")

		if m.dryRun {
			m.log("(dry run) Would apply migration", "version", mig.Version, "name", mig.Name)
			result.Applied = append(result.Applied, mig.Version)
			continue
		}

		migStart := time.Now()

		// Run the migration
		if err := m.runMigration(ctx, mig, DirectionUp, batch); err != nil {
			// Mark as failed
			_ = m.state.MarkFailed(ctx, mig.Version, mig.Name, err, time.Since(migStart).Milliseconds())
			return result, NewMigrationError(mig.Version, mig.Name, DirectionUp, err)
		}

		result.Applied = append(result.Applied, mig.Version)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Down rolls back N migrations
func (m *Migrator) Down(ctx context.Context, steps int) (*MigrationResult, error) {
	if steps <= 0 {
		steps = 1
	}

	start := time.Now()
	result := &MigrationResult{DryRun: m.dryRun}

	// Acquire lock
	if !m.dryRun {
		if err := m.lock.Acquire(ctx); err != nil {
			return nil, err
		}
		defer func() { _ = m.lock.Release(ctx) }()
	}

	// Get applied migrations in reverse order
	applied, err := m.state.GetApplied(ctx)
	if err != nil {
		return nil, err
	}

	if len(applied) == 0 {
		result.Duration = time.Since(start)
		return result, nil
	}

	// Sort in reverse order
	slices.SortFunc(applied, func(a, b MigrationRecord) int {
		return cmp.Compare(b.Version, a.Version) // newest first
	})

	// Limit to steps
	if steps > len(applied) {
		steps = len(applied)
	}

	toRollback := applied[:steps]

	for _, record := range toRollback {
		mig, err := GetMigration(record.Version)
		if err != nil {
			// Migration not found in registry, skip
			m.log("Migration not found in registry, skipping", "version", record.Version)
			result.Skipped = append(result.Skipped, record.Version)
			continue
		}

		if !mig.HasDown() {
			m.log("Migration has no down function, skipping", "version", mig.Version, "name", mig.Name)
			result.Skipped = append(result.Skipped, mig.Version)
			continue
		}

		m.log("Rolling back migration", "version", mig.Version, "name", mig.Name, "direction", "down")

		if m.dryRun {
			m.log("(dry run) Would rollback migration", "version", mig.Version, "name", mig.Name)
			result.RolledBack = append(result.RolledBack, mig.Version)
			continue
		}

		if err := m.runMigration(ctx, mig, DirectionDown, 0); err != nil {
			return result, NewMigrationError(mig.Version, mig.Name, DirectionDown, err)
		}

		result.RolledBack = append(result.RolledBack, mig.Version)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// DownTo rolls back to a specific version (exclusive - the target version remains applied)
func (m *Migrator) DownTo(ctx context.Context, version int64) (*MigrationResult, error) {
	start := time.Now()
	result := &MigrationResult{DryRun: m.dryRun}

	// Acquire lock
	if !m.dryRun {
		if err := m.lock.Acquire(ctx); err != nil {
			return nil, err
		}
		defer func() { _ = m.lock.Release(ctx) }()
	}

	// Get applied migrations in reverse order
	applied, err := m.state.GetApplied(ctx)
	if err != nil {
		return nil, err
	}

	// Sort in reverse order
	slices.SortFunc(applied, func(a, b MigrationRecord) int {
		return cmp.Compare(b.Version, a.Version) // newest first
	})

	for _, record := range applied {
		if record.Version <= version {
			break
		}

		mig, err := GetMigration(record.Version)
		if err != nil {
			result.Skipped = append(result.Skipped, record.Version)
			continue
		}

		if !mig.HasDown() {
			m.log("Migration has no down function, skipping", "version", mig.Version, "name", mig.Name)
			result.Skipped = append(result.Skipped, mig.Version)
			continue
		}

		m.log("Rolling back migration", "version", mig.Version, "name", mig.Name, "direction", "down")

		if m.dryRun {
			result.RolledBack = append(result.RolledBack, mig.Version)
			continue
		}

		if err := m.runMigration(ctx, mig, DirectionDown, 0); err != nil {
			return result, NewMigrationError(mig.Version, mig.Name, DirectionDown, err)
		}

		result.RolledBack = append(result.RolledBack, mig.Version)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Reset rolls back all applied migrations
func (m *Migrator) Reset(ctx context.Context) (*MigrationResult, error) {
	return m.DownTo(ctx, 0)
}

// Redo rolls back then re-applies the last N migrations
func (m *Migrator) Redo(ctx context.Context, steps int) (*MigrationResult, error) {
	start := time.Now()

	// First, rollback
	downResult, err := m.Down(ctx, steps)
	if err != nil {
		return downResult, err
	}

	// Then, reapply
	upResult, err := m.Up(ctx)
	if err != nil {
		// Combine results
		downResult.Applied = upResult.Applied
		downResult.Duration = time.Since(start)
		return downResult, err
	}

	// Combine results
	result := &MigrationResult{
		Applied:    upResult.Applied,
		RolledBack: downResult.RolledBack,
		Skipped:    append(downResult.Skipped, upResult.Skipped...),
		Duration:   time.Since(start),
		DryRun:     m.dryRun,
	}

	return result, nil
}

// Status returns the status of all migrations
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatusEntry, error) {
	applied, err := m.state.GetApplied(ctx)
	if err != nil {
		return nil, err
	}

	// Build map of applied migrations
	appliedMap := make(map[int64]MigrationRecord)
	for _, r := range applied {
		appliedMap[r.Version] = r
	}

	// Check for dirty state
	dirty, err := m.state.GetDirty(ctx)
	if err != nil {
		return nil, err
	}

	// Get all registered migrations
	migrations := GetMigrations()

	var entries []MigrationStatusEntry
	for _, mig := range migrations {
		entry := MigrationStatusEntry{
			Version: mig.Version,
			Name:    mig.Name,
		}

		if dirty != nil && dirty.Version == mig.Version {
			entry.Status = "dirty"
			entry.AppliedAt = &dirty.AppliedAt
		} else if record, ok := appliedMap[mig.Version]; ok {
			entry.Status = "applied"
			entry.AppliedAt = &record.AppliedAt
			entry.Batch = record.Batch
		} else {
			entry.Status = "pending"
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Pending returns migrations that haven't been applied yet
func (m *Migrator) Pending(ctx context.Context) ([]MigrationInfo, error) {
	appliedVersions, err := m.state.GetAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	migrations := GetMigrations()
	var pending []MigrationInfo

	for _, mig := range migrations {
		if !appliedVersions[mig.Version] {
			pending = append(pending, MigrationInfo{
				Version: mig.Version,
				Name:    mig.Name,
			})
		}
	}

	return pending, nil
}

// Version returns the current schema version
func (m *Migrator) Version(ctx context.Context) (int64, error) {
	return m.state.GetVersion(ctx)
}

// DryRun simulates migrations without applying them
func (m *Migrator) DryRun(ctx context.Context) (*MigrationResult, error) {
	// Temporarily enable dry run
	oldDryRun := m.dryRun
	m.dryRun = true
	defer func() { m.dryRun = oldDryRun }()

	return m.Up(ctx)
}

// VerifyChecksums verifies all applied migrations' checksums
func (m *Migrator) VerifyChecksums(ctx context.Context) error {
	applied, err := m.state.GetApplied(ctx)
	if err != nil {
		return err
	}

	for _, record := range applied {
		mig, err := GetMigration(record.Version)
		if err != nil {
			continue // Migration not in registry, can't verify
		}

		if record.Checksum != "" && record.Checksum != mig.Checksum() {
			return NewChecksumError(record.Version, record.Name, record.Checksum, mig.Checksum())
		}
	}

	return nil
}

// ClearDirty clears the dirty state for manual intervention
func (m *Migrator) ClearDirty(ctx context.Context, version int64) error {
	return m.state.ClearDirty(ctx, version)
}

// runMigration runs a single migration with timeout and state tracking
func (m *Migrator) runMigration(ctx context.Context, mig *Migration, direction Direction, batch int) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	migStart := time.Now()

	var runErr error
	if direction == DirectionUp {
		runErr = mig.RunUp(ctx, m.db)
	} else {
		runErr = mig.RunDown(ctx, m.db)
	}

	durationMs := time.Since(migStart).Milliseconds()

	if runErr != nil {
		return runErr
	}

	// Update state
	if direction == DirectionUp {
		record := &MigrationRecord{
			Version:    mig.Version,
			Name:       mig.Name,
			AppliedAt:  time.Now(),
			DurationMs: durationMs,
			Checksum:   mig.Checksum(),
			Batch:      batch,
		}
		if err := m.state.MarkApplied(ctx, record); err != nil {
			return err
		}
	} else {
		if err := m.state.MarkRolledBack(ctx, mig.Version); err != nil {
			return err
		}
	}

	return nil
}

// log logs a message if a logger is available
func (m *Migrator) log(msg string, args ...any) {
	if m.logger != nil {
		m.logger.Info(msg, args...)
	}
}

// GetAppliedMigrations returns all applied migrations
func (m *Migrator) GetAppliedMigrations(ctx context.Context) ([]MigrationRecord, error) {
	return m.state.GetApplied(ctx)
}

// ForceClearLock forcefully clears the migration lock (for emergency use)
func (m *Migrator) ForceClearLock(ctx context.Context) error {
	return m.lock.ForceClear(ctx)
}

// GetLockInfo returns information about the current lock holder
func (m *Migrator) GetLockInfo(ctx context.Context) (*MigrationLock, error) {
	return m.lock.GetLockInfo(ctx)
}

// RunInTransaction runs a function within a MongoDB transaction
// This is useful for migrations that need transaction support
func (m *Migrator) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if !m.config.UseTransactions {
		return fn(ctx)
	}

	session, err := m.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		return nil, fn(ctx)
	})

	return err
}

// Refresh returns a refreshed view of migration status (useful for watch operations)
func (m *Migrator) Refresh(ctx context.Context) (*MigrationResult, error) {
	version, err := m.Version(ctx)
	if err != nil {
		return nil, err
	}

	pending, err := m.Pending(ctx)
	if err != nil {
		return nil, err
	}

	var pendingVersions []int64
	for _, p := range pending {
		pendingVersions = append(pendingVersions, p.Version)
	}

	return &MigrationResult{
		Applied: []int64{version},
		Skipped: pendingVersions,
	}, nil
}

// SetSchemaVersionOnCollection sets the _goe_sv field on all documents in a collection
func (m *Migrator) SetSchemaVersionOnCollection(ctx context.Context, collection string, version int64, filter bson.M) error {
	if filter == nil {
		filter = bson.M{}
	}

	update := bson.M{"$set": bson.M{m.config.SchemaVersionField: version}}
	_, err := m.db.Collection(collection).UpdateMany(ctx, filter, update)
	return err
}
