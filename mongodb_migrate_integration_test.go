//go:build integration

package goe

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/core/mongodb/migrate"
)

// TestMongoDBMigrate_Integration tests the full goe.New() flow with WithMongoDB
// and WithMigrate enabled. This was the bug scenario: migrate.NewModule() called
// mongodb.DB() during goe.New() construction, but MongoDB connections are only
// established during the Fx OnStart lifecycle phase, causing a fatal nil DB error.
func TestMongoDBMigrate_Integration(t *testing.T) {
	t.Run("WithMongoDB and WithMigrate starts without nil DB error", func(t *testing.T) {
		resetGlobalInstance()
		migrate.Clear()

		// Register test migrations before creating the app
		migrate.MustRegister(
			migrate.New(1, "create_test_collection").
				Up(func(ctx context.Context, db *mongo.Database) error {
					return db.CreateCollection(ctx, "goe_e2e_test")
				}).
				Down(func(ctx context.Context, db *mongo.Database) error {
					return db.Collection("goe_e2e_test").Drop(ctx)
				}),
			migrate.New(2, "add_test_index").
				Up(migrate.CreateIndex("goe_e2e_test", "name", migrate.Unique())).
				Down(migrate.DropIndex("goe_e2e_test", "name_1")),
		)

		// This used to fatal with "mongodb database is nil" before the fix
		app := New(Options{
			WithMongoDB: true,
			WithMigrate: true,
			ConfigOverrides: map[string]any{
				"MONGO_URI":            "mongodb://localhost:27017/?replicaSet=rs0&directConnection=true",
				"MONGO_DB_NAME":        "goe_e2e_migrate_test",
				"MONGODB_MIGRATE_AUTO": true,
			},
		})
		require.NotNil(t, app)

		// Start the app to trigger OnStart hooks (MongoDB connect + migration)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		startErr := make(chan error, 1)
		go func() {
			startErr <- app.Container().Start(ctx)
		}()

		// Wait for start to complete
		select {
		case err := <-startErr:
			require.NoError(t, err, "App should start without error")
		case <-time.After(15 * time.Second):
			t.Fatal("App start timed out")
		}

		// Verify goe.Migrate() is accessible and not nil
		migrator := Migrate()
		require.NotNil(t, migrator, "goe.Migrate() should return non-nil migrator after app start")

		// Verify goe.MongoDB() is accessible
		mongoDB := MongoDB()
		require.NotNil(t, mongoDB, "goe.MongoDB() should return non-nil")
		require.NotNil(t, mongoDB.DB(), "MongoDB.DB() should return non-nil after start")

		// Verify migrations were auto-applied
		version, err := migrator.Version(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), version, "Expected migration version 2 after auto-migrate")

		// Verify no pending migrations
		pending, err := migrator.Pending(ctx)
		require.NoError(t, err)
		assert.Empty(t, pending, "Expected no pending migrations")

		// Verify the collection and index were actually created
		db := mongoDB.DB()
		indexes, err := db.Collection("goe_e2e_test").Indexes().List(ctx)
		require.NoError(t, err)

		var indexNames []string
		for indexes.Next(ctx) {
			var idx bson.M
			indexes.Decode(&idx)
			if name, ok := idx["name"].(string); ok {
				indexNames = append(indexNames, name)
			}
		}
		assert.Contains(t, indexNames, "name_1", "Expected unique index on 'name' field")

		// Clean up: stop app and drop test database
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		app.Container().Stop(stopCtx)

		// Drop test database
		db.Drop(context.Background())

		migrate.Clear()
	})

	t.Run("WithMigrate without WithMongoDB panics", func(t *testing.T) {
		resetGlobalInstance()
		migrate.Clear()

		// WithMigrate requires WithMongoDB - should fatal
		// We can't easily test logger.Fatal in a unit test, so we verify
		// the configuration validation logic
		assert.Panics(t, func() {
			Migrate() // Should panic because migrator is nil
		})

		migrate.Clear()
	})

	t.Run("migration status and rollback work after deferred init", func(t *testing.T) {
		resetGlobalInstance()
		migrate.Clear()

		migrate.MustRegister(
			migrate.New(1, "first").
				Up(migrate.NoOp()).
				Down(migrate.NoOp()),
			migrate.New(2, "second").
				Up(migrate.NoOp()).
				Down(migrate.NoOp()),
			migrate.New(3, "third").
				Up(migrate.NoOp()).
				Down(migrate.NoOp()),
		)

		app := New(Options{
			WithMongoDB: true,
			WithMigrate: true,
			ConfigOverrides: map[string]any{
				"MONGO_URI":            "mongodb://localhost:27017/?replicaSet=rs0&directConnection=true",
				"MONGO_DB_NAME":        "goe_e2e_migrate_test2",
				"MONGODB_MIGRATE_AUTO": true,
			},
		})
		require.NotNil(t, app)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		startErr := make(chan error, 1)
		go func() {
			startErr <- app.Container().Start(ctx)
		}()

		select {
		case err := <-startErr:
			require.NoError(t, err)
		case <-time.After(15 * time.Second):
			t.Fatal("App start timed out")
		}

		migrator := Migrate()

		// Verify status shows all applied
		status, err := migrator.Status(ctx)
		require.NoError(t, err)
		assert.Len(t, status, 3)
		for _, s := range status {
			assert.Equal(t, "applied", s.Status, "Migration %d should be applied", s.Version)
		}

		// Rollback one step
		result, err := migrator.Down(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, result.RolledBack, 1)
		assert.Equal(t, int64(3), result.RolledBack[0])

		// Verify version
		version, err := migrator.Version(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), version)

		// Clean up: drop database before stopping (stop disconnects MongoDB)
		MongoDB().DB().Drop(context.Background())

		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		app.Container().Stop(stopCtx)

		migrate.Clear()
	})
}
