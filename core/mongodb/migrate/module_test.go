package migrate

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// nilMongoDB implements contract.MongoDB with no connection, standing in for
// a MongoDB module whose OnStart has not (or could not have) run.
type nilMongoDB struct{}

func (n *nilMongoDB) Client() *mongo.Client        { return nil }
func (n *nilMongoDB) DB() *mongo.Database          { return nil }
func (n *nilMongoDB) Col(string) *mongo.Collection { return nil }

// testDatabase returns a *mongo.Database without any I/O: mongo.Connect does
// not dial, so this is safe for construction-only tests.
func testDatabase(t *testing.T) *mongo.Database {
	t.Helper()
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://127.0.0.1:1"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	return client.Database("construction_test")
}

func TestNewModule_ResolvesOptionsOverEnvironment(t *testing.T) {
	cfg := newMockConfig()
	cfg.Set("MONGODB_MIGRATE_COLLECTION", "env_migrations")

	m, err := NewModule(cfg, &nopLogger{}, &nilMongoDB{},
		WithCollection("code_migrations"),
		WithTimeout(2*time.Minute),
	)
	require.NoError(t, err)

	assert.Equal(t, "mongodb_migrate", m.Name())
	assert.Equal(t, "code_migrations", m.Config().Collection)
	assert.Equal(t, 2*time.Minute, m.Config().Timeout)
	// The migrator does not exist until OnStart provides the database.
	assert.Nil(t, m.Migrator())
}

func TestModule_OnStartFailsWithoutDatabase(t *testing.T) {
	// The MongoDB module now fails startup when unreachable, so a nil DB here
	// means mis-ordered lifecycles; the migration module must say so rather
	// than panic.
	m, err := NewModule(newMockConfig(), &nopLogger{}, &nilMongoDB{})
	require.NoError(t, err)

	err = m.OnStart(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mongodb database is nil")
}

func TestModule_OnStopWithoutMigratorIsClean(t *testing.T) {
	m, err := NewModule(newMockConfig(), &nopLogger{}, &nilMongoDB{})
	require.NoError(t, err)
	assert.NoError(t, m.OnStop(context.Background()))
}

func TestNewModuleWithDB_ConstructsMigrator(t *testing.T) {
	db := testDatabase(t)

	m, err := NewModuleWithDB(newMockConfig(), &nopLogger{}, db)
	require.NoError(t, err)

	require.NotNil(t, m.Migrator())
	assert.Same(t, m.Migrator(), m.Provide())
	assert.Same(t, m.Migrator(), m.ProvideMigrator())
}

func TestNewModuleWithMigrator_WrapsExisting(t *testing.T) {
	db := testDatabase(t)
	migrator := NewMigrator(db, WithLogger(&nopLogger{}))

	m := NewModuleWithMigrator(migrator, &nopLogger{})
	assert.Same(t, migrator, m.Migrator())
	assert.Same(t, migrator.config, m.Config())
}
