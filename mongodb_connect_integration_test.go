//go:build integration

package goe_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
)

// Needs a MongoDB: the test stack's (make test_services_up), or the one
// GOE_TEST_MONGO_URI names.
func TestMongoDB_AvailableToInvokers(t *testing.T) {
	uri := os.Getenv("GOE_TEST_MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017/?replicaSet=rs0&directConnection=true"
	}
	var dbReady bool
	app := goe.New(goe.Options{
		WithMongoDB: true,
		ConfigOverrides: map[string]any{
			"MONGO_URI":     uri,
			"MONGO_DB_NAME": "goe_connect_integration",
		},
		Invokers: []any{
			func(m contract.MongoDB) { dbReady = m.DB() != nil },
		},
	})
	assert.True(t, dbReady, "an invoker can use the database before the app starts")

	ctx := context.Background()
	require.NoError(t, app.Start(ctx))
	require.NoError(t, app.Stop(ctx))
}
