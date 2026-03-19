package lock

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientPool(t *testing.T) {
	// Create a client (won't connect until first use)
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer func() { _ = client.Close() }()

	pool := NewClientPool(client)

	require.NotNil(t, pool)
	assert.Equal(t, "localhost:6379", pool.Name())
	assert.NotNil(t, pool.Client())
	assert.Equal(t, client, pool.Client())
}

func TestClientPool_Name(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected string
	}{
		{"default address", "localhost:6379", "localhost:6379"},
		{"custom address", "redis.example.com:6380", "redis.example.com:6380"},
		{"ip address", "192.168.1.100:6379", "192.168.1.100:6379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := redis.NewClient(&redis.Options{Addr: tt.addr})
			defer func() { _ = client.Close() }()

			pool := NewClientPool(client)
			assert.Equal(t, tt.expected, pool.Name())
		})
	}
}

func TestClientPool_Close(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	pool := NewClientPool(client)
	err := pool.Close()

	// Close should not error even if not connected
	assert.NoError(t, err)
}

func TestNewClusterPool(t *testing.T) {
	// Create a cluster client (won't connect until first use)
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
	})
	defer func() { _ = client.Close() }()

	pool := NewClusterPool(client)

	require.NotNil(t, pool)
	assert.Equal(t, "cluster", pool.Name())
	assert.NotNil(t, pool.Client())
	assert.Equal(t, client, pool.Client())
}

func TestClusterPool_Close(t *testing.T) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"localhost:7000"},
	})

	pool := NewClusterPool(client)
	err := pool.Close()

	// Close should not error even if not connected
	assert.NoError(t, err)
}

func TestNewFailoverPool(t *testing.T) {
	// Create a failover client (won't connect until first use)
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"localhost:26379"},
	})
	defer func() { _ = client.Close() }()

	pool := NewFailoverPool(client, "mymaster")

	require.NotNil(t, pool)
	assert.Equal(t, "sentinel:mymaster", pool.Name())
	assert.NotNil(t, pool.Client())
	assert.Equal(t, client, pool.Client())
}

func TestFailoverPool_Name(t *testing.T) {
	tests := []struct {
		name       string
		masterName string
		expected   string
	}{
		{"default master", "mymaster", "sentinel:mymaster"},
		{"custom master", "production-master", "sentinel:production-master"},
		{"simple master", "redis", "sentinel:redis"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:    tt.masterName,
				SentinelAddrs: []string{"localhost:26379"},
			})
			defer func() { _ = client.Close() }()

			pool := NewFailoverPool(client, tt.masterName)
			assert.Equal(t, tt.expected, pool.Name())
		})
	}
}

func TestFailoverPool_Close(t *testing.T) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"localhost:26379"},
	})

	pool := NewFailoverPool(client, "mymaster")
	err := pool.Close()

	// Close should not error even if not connected
	assert.NoError(t, err)
}

func TestPoolImplementsInterface(t *testing.T) {
	// Verify all pool types implement the Pool interface
	t.Run("ClientPool implements Pool", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		defer func() { _ = client.Close() }()

		var pool Pool = NewClientPool(client)
		assert.NotNil(t, pool)
	})

	t.Run("ClusterPool implements Pool", func(t *testing.T) {
		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{"localhost:7000"},
		})
		defer func() { _ = client.Close() }()

		var pool Pool = NewClusterPool(client)
		assert.NotNil(t, pool)
	})

	t.Run("FailoverPool implements Pool", func(t *testing.T) {
		client := redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    "mymaster",
			SentinelAddrs: []string{"localhost:26379"},
		})
		defer func() { _ = client.Close() }()

		var pool Pool = NewFailoverPool(client, "mymaster")
		assert.NotNil(t, pool)
	})
}
