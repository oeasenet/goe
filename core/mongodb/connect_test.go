package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModule_ConnectFailsFastWhenUnreachable(t *testing.T) {
	env := newMapConfig()
	env.Set("MONGO_URI", "mongodb://127.0.0.1:1/?serverSelectionTimeoutMS=200&directConnection=true")
	env.Set("MONGO_DB_NAME", "goe_connect_test")
	m := NewDBModule(env, &nopLogger{}, WithPingTimeout(300*time.Millisecond))

	start := time.Now()
	err := m.Connect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ping failed")
	assert.Less(t, time.Since(start), 3*time.Second)
	assert.Nil(t, m.DB())
}

func TestModule_ConnectReportsOptionErrors(t *testing.T) {
	m := NewDBModule(newMapConfig(), &nopLogger{}, WithMaxPoolSize(-1))
	assert.Error(t, m.Connect(context.Background()), "an invalid option fails before connecting")
}
