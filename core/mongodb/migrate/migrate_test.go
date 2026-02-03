package migrate

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestMigrationBuilder(t *testing.T) {
	// Test basic migration creation
	m := New(1, "test_migration")
	if m.Version != 1 {
		t.Errorf("expected version 1, got %d", m.Version)
	}
	if m.Name != "test_migration" {
		t.Errorf("expected name test_migration, got %s", m.Name)
	}
	if m.HasUp() {
		t.Error("expected HasUp to be false")
	}
	if m.HasDown() {
		t.Error("expected HasDown to be false")
	}
}

func TestMigrationBuilderWithFunctions(t *testing.T) {
	upCalled := false
	downCalled := false

	m := New(2, "test_with_funcs").
		Up(func(ctx context.Context, db *mongo.Database) error {
			upCalled = true
			return nil
		}).
		Down(func(ctx context.Context, db *mongo.Database) error {
			downCalled = true
			return nil
		})

	if !m.HasUp() {
		t.Error("expected HasUp to be true")
	}
	if !m.HasDown() {
		t.Error("expected HasDown to be true")
	}

	// Test running the functions
	ctx := context.Background()
	if err := m.RunUp(ctx, nil); err != nil {
		t.Errorf("unexpected error running up: %v", err)
	}
	if !upCalled {
		t.Error("expected up function to be called")
	}

	if err := m.RunDown(ctx, nil); err != nil {
		t.Errorf("unexpected error running down: %v", err)
	}
	if !downCalled {
		t.Error("expected down function to be called")
	}
}

func TestMigrationChecksum(t *testing.T) {
	m1 := New(1, "test").
		Up(func(ctx context.Context, db *mongo.Database) error { return nil })

	m2 := New(1, "test").
		Up(func(ctx context.Context, db *mongo.Database) error { return nil })

	// Checksums should be computed
	if m1.Checksum() == "" {
		t.Error("expected checksum to be computed")
	}
	if m2.Checksum() == "" {
		t.Error("expected checksum to be computed")
	}
}

func TestRegistry(t *testing.T) {
	// Clear the registry before testing
	Clear()

	// Register migrations
	err := Register(
		New(1, "first").Up(NoOp()),
		New(2, "second").Up(NoOp()),
	)
	if err != nil {
		t.Fatalf("unexpected error registering migrations: %v", err)
	}

	// Check count
	if Count() != 2 {
		t.Errorf("expected 2 migrations, got %d", Count())
	}

	// Check versions
	versions := Versions()
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
	if versions[0] != 1 || versions[1] != 2 {
		t.Errorf("expected versions [1, 2], got %v", versions)
	}

	// Get specific migration
	m, err := GetMigration(1)
	if err != nil {
		t.Fatalf("unexpected error getting migration: %v", err)
	}
	if m.Name != "first" {
		t.Errorf("expected name 'first', got %s", m.Name)
	}

	// Get non-existent migration
	_, err = GetMigration(999)
	if !errors.Is(err, ErrMigrationNotFound) {
		t.Errorf("expected ErrMigrationNotFound, got %v", err)
	}

	// Clear and check
	Clear()
	if Count() != 0 {
		t.Errorf("expected 0 migrations after clear, got %d", Count())
	}
}

func TestRegistryDuplicateVersion(t *testing.T) {
	Clear()

	// Register first migration
	err := Register(New(1, "first").Up(NoOp()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to register duplicate version
	err = Register(New(1, "duplicate").Up(NoOp()))
	if !errors.Is(err, ErrDuplicateVersion) {
		t.Errorf("expected ErrDuplicateVersion, got %v", err)
	}

	Clear()
}

func TestRegistryInvalidVersion(t *testing.T) {
	Clear()

	err := Register(New(0, "invalid").Up(NoOp()))
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}

	err = Register(New(-1, "negative").Up(NoOp()))
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}

	Clear()
}

func TestMustRegisterPanics(t *testing.T) {
	Clear()

	// Register first
	MustRegister(New(1, "first").Up(NoOp()))

	// Should panic on duplicate
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for duplicate version")
		}
	}()

	MustRegister(New(1, "duplicate").Up(NoOp()))
}

func TestGetMigrationsAfter(t *testing.T) {
	Clear()

	Register(
		New(1, "first").Up(NoOp()),
		New(2, "second").Up(NoOp()),
		New(3, "third").Up(NoOp()),
	)

	migrations := GetMigrationsAfter(1)
	if len(migrations) != 2 {
		t.Errorf("expected 2 migrations after version 1, got %d", len(migrations))
	}

	migrations = GetMigrationsAfter(3)
	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations after version 3, got %d", len(migrations))
	}

	Clear()
}

func TestGetMigrationsInRange(t *testing.T) {
	Clear()

	Register(
		New(1, "first").Up(NoOp()),
		New(2, "second").Up(NoOp()),
		New(3, "third").Up(NoOp()),
		New(4, "fourth").Up(NoOp()),
	)

	migrations := GetMigrationsInRange(2, 3)
	if len(migrations) != 2 {
		t.Errorf("expected 2 migrations in range [2,3], got %d", len(migrations))
	}

	Clear()
}

func TestGetLatestVersion(t *testing.T) {
	Clear()

	if GetLatestVersion() != 0 {
		t.Error("expected 0 for empty registry")
	}

	Register(
		New(5, "five").Up(NoOp()),
		New(2, "two").Up(NoOp()),
		New(10, "ten").Up(NoOp()),
	)

	if GetLatestVersion() != 10 {
		t.Errorf("expected latest version 10, got %d", GetLatestVersion())
	}

	Clear()
}

func TestConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Collection != "_goe_migrations" {
		t.Errorf("expected collection _goe_migrations, got %s", cfg.Collection)
	}
	if cfg.AutoMigrate != false {
		t.Error("expected AutoMigrate to be false by default")
	}
	if cfg.VerifyChecksums != true {
		t.Error("expected VerifyChecksums to be true by default")
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check defaults were applied
	if cfg.SchemaVersionField != "_goe_sv" {
		t.Errorf("expected _goe_sv, got %s", cfg.SchemaVersionField)
	}
}

func TestErrors(t *testing.T) {
	// Test MigrationError
	migErr := NewMigrationError(1, "test", DirectionUp, errors.New("test error"))
	if migErr.Version != 1 {
		t.Errorf("expected version 1, got %d", migErr.Version)
	}
	if migErr.Direction != DirectionUp {
		t.Errorf("expected direction up, got %s", migErr.Direction)
	}
	if !errors.Is(migErr, migErr.Cause) {
		t.Error("expected Unwrap to return cause")
	}

	// Test ChecksumError
	checksumErr := NewChecksumError(1, "test", "expected", "actual")
	if !errors.Is(checksumErr, ErrChecksumMismatch) {
		t.Error("expected ChecksumError to match ErrChecksumMismatch")
	}

	// Test DirtyStateError
	dirtyErr := NewDirtyStateError(1, "test", "error msg", "2024-01-01T00:00:00Z")
	if !errors.Is(dirtyErr, ErrDirtyState) {
		t.Error("expected DirtyStateError to match ErrDirtyState")
	}
}

func TestSequenceHelper(t *testing.T) {
	callOrder := []int{}

	seq := Sequence(
		func(ctx context.Context, db *mongo.Database) error {
			callOrder = append(callOrder, 1)
			return nil
		},
		func(ctx context.Context, db *mongo.Database) error {
			callOrder = append(callOrder, 2)
			return nil
		},
		func(ctx context.Context, db *mongo.Database) error {
			callOrder = append(callOrder, 3)
			return nil
		},
	)

	err := seq(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(callOrder) != 3 {
		t.Errorf("expected 3 calls, got %d", len(callOrder))
	}
	for i, v := range callOrder {
		if v != i+1 {
			t.Errorf("expected call order %d at position %d, got %d", i+1, i, v)
		}
	}
}

func TestSequenceStopsOnError(t *testing.T) {
	callCount := 0
	testErr := errors.New("test error")

	seq := Sequence(
		func(ctx context.Context, db *mongo.Database) error {
			callCount++
			return nil
		},
		func(ctx context.Context, db *mongo.Database) error {
			callCount++
			return testErr
		},
		func(ctx context.Context, db *mongo.Database) error {
			callCount++
			return nil
		},
	)

	err := seq(context.Background(), nil)
	if !errors.Is(err, testErr) {
		t.Errorf("expected testErr, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls before error, got %d", callCount)
	}
}

func TestNoOp(t *testing.T) {
	noop := NoOp()
	err := noop(context.Background(), nil)
	if err != nil {
		t.Errorf("expected nil error from NoOp, got %v", err)
	}
}

func TestIndexOptions(t *testing.T) {
	cfg := buildIndexConfig([]IndexOption{
		Unique(),
		Sparse(),
		Name("test_index"),
	})

	if !cfg.unique {
		t.Error("expected unique to be true")
	}
	if !cfg.sparse {
		t.Error("expected sparse to be true")
	}
	if cfg.name != "test_index" {
		t.Errorf("expected name 'test_index', got %s", cfg.name)
	}
}

func TestNoVersionBumpOption(t *testing.T) {
	cfg := buildIndexConfig([]IndexOption{
		NoVersionBump(),
		WithSchemaVersion(5),
	})

	if !cfg.noVersionBump {
		t.Error("expected noVersionBump to be true")
	}
	if cfg.schemaVersion != 5 {
		t.Errorf("expected schemaVersion 5, got %d", cfg.schemaVersion)
	}
}
