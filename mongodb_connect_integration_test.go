//go:build integration

package goe_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
)

// Needs the test stack's MongoDB (make test_services_up).
func TestMongoDB_AvailableToInvokers(t *testing.T) {
	var dbReady bool
	app := goe.New(goe.Options{
		WithMongoDB: true,
		ConfigOverrides: map[string]any{
			"MONGO_URI":     "mongodb://localhost:27017/?replicaSet=rs0&directConnection=true",
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
