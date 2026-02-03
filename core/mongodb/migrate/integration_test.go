//go:build integration

package migrate

import (
	"context"
	"errors"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var testDB *mongo.Database

func TestMain(m *testing.M) {
	// Setup
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic("Failed to connect to MongoDB: " + err.Error())
	}

	// Use a unique database name for tests
	testDB = client.Database("goe_migrate_test")

	// Run tests
	code := m.Run()

	// Teardown
	testDB.Drop(context.Background())
	client.Disconnect(context.Background())

	os.Exit(code)
}

func setupMigrator(t *testing.T) *Migrator {
	t.Helper()

	// Clear registry for each test
	Clear()

	// Drop all collections to start fresh
	ctx := context.Background()
	collections, _ := testDB.ListCollectionNames(ctx, bson.M{})
	for _, coll := range collections {
		testDB.Collection(coll).Drop(ctx)
	}

	cfg := DefaultConfig()
	cfg.Collection = "_test_migrations"

	return NewMigrator(testDB, WithConfig(cfg))
}

func TestIntegration_MigratorInit(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	err := migrator.Init(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize migrator: %v", err)
	}

	// Verify the collection was created
	collections, err := testDB.ListCollectionNames(ctx, bson.M{"name": "_test_migrations"})
	if err != nil {
		t.Fatalf("Failed to list collections: %v", err)
	}

	// Collection may not exist until first insert, that's OK
	t.Logf("Found collections: %v", collections)
}

func TestIntegration_UpMigrations(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	// Register migrations
	err := Register(
		New(1, "create_users_collection").
			Up(func(ctx context.Context, db *mongo.Database) error {
				return db.CreateCollection(ctx, "users")
			}).
			Down(func(ctx context.Context, db *mongo.Database) error {
				return db.Collection("users").Drop(ctx)
			}),
		New(2, "add_email_index").
			Up(CreateIndex("users", "email", Unique())).
			Down(DropIndex("users", "email_1")),
		New(3, "add_name_field").
			Up(func(ctx context.Context, db *mongo.Database) error {
				// Insert a test document first
				_, err := db.Collection("users").InsertOne(ctx, bson.M{
					"email": "test@example.com",
				})
				return err
			}).
			Down(func(ctx context.Context, db *mongo.Database) error {
				_, err := db.Collection("users").DeleteMany(ctx, bson.M{})
				return err
			}),
	)
	if err != nil {
		t.Fatalf("Failed to register migrations: %v", err)
	}

	// Initialize
	err = migrator.Init(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Run migrations
	result, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	if len(result.Applied) != 3 {
		t.Errorf("Expected 3 applied migrations, got %d", len(result.Applied))
	}

	t.Logf("Applied migrations: %v in %v", result.Applied, result.Duration)

	// Verify current version
	version, err := migrator.Version(ctx)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}
	if version != 3 {
		t.Errorf("Expected version 3, got %d", version)
	}

	// Verify no pending migrations
	pending, err := migrator.Pending(ctx)
	if err != nil {
		t.Fatalf("Failed to get pending: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("Expected 0 pending migrations, got %d", len(pending))
	}

	// Verify users collection exists with index
	indexes, err := testDB.Collection("users").Indexes().List(ctx)
	if err != nil {
		t.Fatalf("Failed to list indexes: %v", err)
	}
	var indexNames []string
	for indexes.Next(ctx) {
		var idx bson.M
		indexes.Decode(&idx)
		if name, ok := idx["name"].(string); ok {
			indexNames = append(indexNames, name)
		}
	}
	t.Logf("User indexes: %v", indexNames)

	// Check for email index
	hasEmailIndex := false
	for _, name := range indexNames {
		if name == "email_1" {
			hasEmailIndex = true
			break
		}
	}
	if !hasEmailIndex {
		t.Error("Expected email_1 index to exist")
	}
}

func TestIntegration_DownMigrations(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	// Register migrations
	Register(
		New(1, "first").
			Up(func(ctx context.Context, db *mongo.Database) error {
				return db.CreateCollection(ctx, "test_coll_1")
			}).
			Down(func(ctx context.Context, db *mongo.Database) error {
				return db.Collection("test_coll_1").Drop(ctx)
			}),
		New(2, "second").
			Up(func(ctx context.Context, db *mongo.Database) error {
				return db.CreateCollection(ctx, "test_coll_2")
			}).
			Down(func(ctx context.Context, db *mongo.Database) error {
				return db.Collection("test_coll_2").Drop(ctx)
			}),
		New(3, "third").
			Up(func(ctx context.Context, db *mongo.Database) error {
				return db.CreateCollection(ctx, "test_coll_3")
			}).
			Down(func(ctx context.Context, db *mongo.Database) error {
				return db.Collection("test_coll_3").Drop(ctx)
			}),
	)

	migrator.Init(ctx)

	// Run all migrations
	_, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Failed to run Up: %v", err)
	}

	// Rollback 1 step
	result, err := migrator.Down(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to run Down: %v", err)
	}

	if len(result.RolledBack) != 1 {
		t.Errorf("Expected 1 rollback, got %d", len(result.RolledBack))
	}
	if result.RolledBack[0] != 3 {
		t.Errorf("Expected version 3 to be rolled back, got %d", result.RolledBack[0])
	}

	// Verify version
	version, err := migrator.Version(ctx)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}
	if version != 2 {
		t.Errorf("Expected version 2, got %d", version)
	}

	// Verify test_coll_3 is dropped
	collections, _ := testDB.ListCollectionNames(ctx, bson.M{"name": "test_coll_3"})
	if len(collections) > 0 {
		t.Error("Expected test_coll_3 to be dropped")
	}
}

func TestIntegration_DownToVersion(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	// Register migrations
	Register(
		New(1, "first").Up(NoOp()).Down(NoOp()),
		New(2, "second").Up(NoOp()).Down(NoOp()),
		New(3, "third").Up(NoOp()).Down(NoOp()),
		New(4, "fourth").Up(NoOp()).Down(NoOp()),
	)

	migrator.Init(ctx)
	migrator.Up(ctx)

	// Rollback to version 2 (exclusive - versions 3 and 4 should be rolled back)
	result, err := migrator.DownTo(ctx, 2)
	if err != nil {
		t.Fatalf("Failed to DownTo: %v", err)
	}

	if len(result.RolledBack) != 2 {
		t.Errorf("Expected 2 rollbacks, got %d: %v", len(result.RolledBack), result.RolledBack)
	}

	version, _ := migrator.Version(ctx)
	if version != 2 {
		t.Errorf("Expected version 2, got %d", version)
	}
}

func TestIntegration_Reset(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	Register(
		New(1, "first").Up(NoOp()).Down(NoOp()),
		New(2, "second").Up(NoOp()).Down(NoOp()),
	)

	migrator.Init(ctx)
	migrator.Up(ctx)

	result, err := migrator.Reset(ctx)
	if err != nil {
		t.Fatalf("Failed to Reset: %v", err)
	}

	if len(result.RolledBack) != 2 {
		t.Errorf("Expected 2 rollbacks, got %d", len(result.RolledBack))
	}

	version, _ := migrator.Version(ctx)
	if version != 0 {
		t.Errorf("Expected version 0 after reset, got %d", version)
	}
}

func TestIntegration_Redo(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	callCount := 0
	Register(
		New(1, "first").
			Up(func(ctx context.Context, db *mongo.Database) error {
				callCount++
				return nil
			}).
			Down(NoOp()),
	)

	migrator.Init(ctx)
	migrator.Up(ctx)

	if callCount != 1 {
		t.Errorf("Expected 1 up call, got %d", callCount)
	}

	// Redo should rollback and reapply
	_, err := migrator.Redo(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to Redo: %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 up calls after redo, got %d", callCount)
	}
}

func TestIntegration_Status(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	Register(
		New(1, "first").Up(NoOp()).Down(NoOp()),
		New(2, "second").Up(NoOp()).Down(NoOp()),
		New(3, "third").Up(NoOp()).Down(NoOp()),
	)

	migrator.Init(ctx)

	// Apply only first 2
	migrator.UpTo(ctx, 2)

	status, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if len(status) != 3 {
		t.Errorf("Expected 3 status entries, got %d", len(status))
	}

	for _, s := range status {
		t.Logf("Migration %d (%s): %s", s.Version, s.Name, s.Status)
	}

	// Check statuses
	appliedCount := 0
	pendingCount := 0
	for _, s := range status {
		switch s.Status {
		case "applied":
			appliedCount++
		case "pending":
			pendingCount++
		}
	}

	if appliedCount != 2 {
		t.Errorf("Expected 2 applied, got %d", appliedCount)
	}
	if pendingCount != 1 {
		t.Errorf("Expected 1 pending, got %d", pendingCount)
	}
}

func TestIntegration_DryRun(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	executed := false
	Register(
		New(1, "test").
			Up(func(ctx context.Context, db *mongo.Database) error {
				executed = true
				return nil
			}).
			Down(NoOp()),
	)

	migrator.Init(ctx)

	result, err := migrator.DryRun(ctx)
	if err != nil {
		t.Fatalf("Failed dry run: %v", err)
	}

	if !result.DryRun {
		t.Error("Expected DryRun flag to be true")
	}

	if len(result.Applied) != 1 {
		t.Errorf("Expected 1 migration in dry run result, got %d", len(result.Applied))
	}

	if executed {
		t.Error("Migration should not have been executed in dry run mode")
	}

	// Verify nothing was actually applied
	version, _ := migrator.Version(ctx)
	if version != 0 {
		t.Errorf("Expected version 0 after dry run, got %d", version)
	}
}

func TestIntegration_DirtyState(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	testErr := errors.New("intentional failure")
	Register(
		New(1, "failing_migration").
			Up(func(ctx context.Context, db *mongo.Database) error {
				return testErr
			}).
			Down(NoOp()),
	)

	migrator.Init(ctx)

	// This should fail
	_, err := migrator.Up(ctx)
	if err == nil {
		t.Fatal("Expected migration to fail")
	}

	// Check dirty state
	dirty, err := migrator.state.GetDirty(ctx)
	if err != nil {
		t.Fatalf("Failed to get dirty state: %v", err)
	}
	if dirty == nil {
		t.Fatal("Expected dirty state to be set")
	}

	t.Logf("Dirty state: version=%d, error=%s", dirty.Version, dirty.Error)

	// Trying to run Up again should fail with dirty state error
	_, err = migrator.Up(ctx)
	if !errors.Is(err, ErrDirtyState) {
		t.Errorf("Expected ErrDirtyState, got %v", err)
	}

	// Clear dirty state
	err = migrator.ClearDirty(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to clear dirty: %v", err)
	}

	// Now dirty should be nil
	dirty, _ = migrator.state.GetDirty(ctx)
	if dirty != nil {
		t.Error("Expected dirty state to be cleared")
	}
}

func TestIntegration_Lock(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	Register(New(1, "test").Up(NoOp()).Down(NoOp()))
	migrator.Init(ctx)

	// Manually acquire lock
	err := migrator.lock.Acquire(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Get lock info
	info, err := migrator.GetLockInfo(ctx)
	if err != nil {
		t.Fatalf("Failed to get lock info: %v", err)
	}
	if info == nil {
		t.Fatal("Expected lock info")
	}
	t.Logf("Lock held by: %s at %v", info.LockedBy, info.LockedAt)

	// Release lock
	err = migrator.lock.Release(ctx)
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Lock should be released
	info, _ = migrator.GetLockInfo(ctx)
	if info != nil {
		t.Error("Expected lock to be released")
	}
}

func TestIntegration_Checksums(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	fn := func(ctx context.Context, db *mongo.Database) error { return nil }

	Register(New(1, "test").Up(fn).Down(NoOp()))
	migrator.Init(ctx)
	migrator.Up(ctx)

	// Verify checksums should pass
	err := migrator.VerifyChecksums(ctx)
	if err != nil {
		t.Fatalf("Checksum verification failed: %v", err)
	}
}

func TestIntegration_HelperCreateIndex(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	Register(
		New(1, "create_collection").
			Up(CreateCollection("helper_test")).
			Down(DropCollection("helper_test")),
		New(2, "create_index").
			Up(CreateIndex("helper_test", "field1", Unique(), Sparse())).
			Down(DropIndex("helper_test", "field1_1")),
		New(3, "create_compound").
			Up(CreateCompoundIndex("helper_test", []string{"field2", "field3"}, Name("compound_idx"))).
			Down(DropIndex("helper_test", "compound_idx")),
	)

	migrator.Init(ctx)
	result, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	t.Logf("Applied: %v", result.Applied)

	// Verify indexes
	indexes, _ := testDB.Collection("helper_test").Indexes().List(ctx)
	var indexNames []string
	for indexes.Next(ctx) {
		var idx bson.M
		indexes.Decode(&idx)
		indexNames = append(indexNames, idx["name"].(string))
	}
	t.Logf("Indexes: %v", indexNames)

	// Check indexes exist
	hasField1 := false
	hasCompound := false
	for _, name := range indexNames {
		if name == "field1_1" {
			hasField1 = true
		}
		if name == "compound_idx" {
			hasCompound = true
		}
	}

	if !hasField1 {
		t.Error("Expected field1_1 index")
	}
	if !hasCompound {
		t.Error("Expected compound_idx index")
	}
}

func TestIntegration_HelperAddRemoveField(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	Register(
		New(1, "setup").
			Up(func(ctx context.Context, db *mongo.Database) error {
				_, err := db.Collection("field_test").InsertMany(ctx, []any{
					bson.M{"name": "doc1"},
					bson.M{"name": "doc2"},
				})
				return err
			}).
			Down(DropCollection("field_test")),
		New(2, "add_status").
			Up(AddFieldWithVersion("field_test", "status", "active", 2)).
			Down(RemoveField("field_test", "status")),
	)

	migrator.Init(ctx)
	migrator.Up(ctx)

	// Check that status field was added
	cursor, _ := testDB.Collection("field_test").Find(ctx, bson.M{})
	var docs []bson.M
	cursor.All(ctx, &docs)

	for _, doc := range docs {
		if doc["status"] != "active" {
			t.Errorf("Expected status=active, got %v", doc["status"])
		}
		t.Logf("Doc: %v", doc)
	}

	// Rollback
	migrator.Down(ctx, 1)

	// Check status field was removed
	cursor, _ = testDB.Collection("field_test").Find(ctx, bson.M{})
	docs = nil
	cursor.All(ctx, &docs)

	for _, doc := range docs {
		if _, hasStatus := doc["status"]; hasStatus {
			t.Error("Expected status field to be removed")
		}
	}
}

func TestIntegration_SequenceHelper(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	Register(
		New(1, "sequence_test").
			Up(Sequence(
				CreateCollection("seq_test"),
				CreateIndex("seq_test", "field1"),
				CreateIndex("seq_test", "field2"),
			)).
			Down(DropCollection("seq_test")),
	)

	migrator.Init(ctx)
	_, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Verify collection and indexes
	indexes, _ := testDB.Collection("seq_test").Indexes().List(ctx)
	count := 0
	for indexes.Next(ctx) {
		count++
	}
	// Should have _id, field1, field2
	if count != 3 {
		t.Errorf("Expected 3 indexes, got %d", count)
	}
}

func TestIntegration_MigrationWithoutDown(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	Register(
		New(1, "no_down").Up(NoOp()), // No Down function
	)

	migrator.Init(ctx)
	migrator.Up(ctx)

	// Rollback should skip this migration
	result, err := migrator.Down(ctx, 1)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if len(result.Skipped) != 1 {
		t.Errorf("Expected 1 skipped, got %d", len(result.Skipped))
	}
}

func TestIntegration_UpToVersion(t *testing.T) {
	migrator := setupMigrator(t)
	ctx := context.Background()

	Register(
		New(1, "first").Up(NoOp()).Down(NoOp()),
		New(2, "second").Up(NoOp()).Down(NoOp()),
		New(3, "third").Up(NoOp()).Down(NoOp()),
		New(4, "fourth").Up(NoOp()).Down(NoOp()),
	)

	migrator.Init(ctx)

	// Only apply up to version 2
	result, err := migrator.UpTo(ctx, 2)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if len(result.Applied) != 2 {
		t.Errorf("Expected 2 applied, got %d", len(result.Applied))
	}

	version, _ := migrator.Version(ctx)
	if version != 2 {
		t.Errorf("Expected version 2, got %d", version)
	}

	// Pending should show 2 migrations
	pending, _ := migrator.Pending(ctx)
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending, got %d", len(pending))
	}
}

func TestIntegration_ConcurrentLockPrevention(t *testing.T) {
	migrator1 := setupMigrator(t)
	ctx := context.Background()

	// Create second migrator pointing to same DB
	cfg := DefaultConfig()
	cfg.Collection = "_test_migrations"
	migrator2 := NewMigrator(testDB, WithConfig(cfg))

	Register(New(1, "test").Up(NoOp()).Down(NoOp()))

	migrator1.Init(ctx)
	migrator2.Init(ctx)

	// First migrator acquires lock
	err := migrator1.lock.Acquire(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	defer migrator1.lock.Release(ctx)

	// Second migrator should fail to acquire
	err = migrator2.lock.Acquire(ctx)
	if err == nil {
		t.Error("Expected second lock acquisition to fail")
		migrator2.lock.Release(ctx)
	} else {
		t.Logf("Correctly failed to acquire second lock: %v", err)
	}
}
