package lock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Pool represents a Redis connection pool that can execute lock operations.
// This interface allows supporting different Redis backends (single, sentinel, cluster).
type Pool interface {
	// SetNX sets a key-value pair with expiry only if the key doesn't exist.
	// Returns true if the key was set, false if it already exists.
	SetNX(ctx context.Context, key, value string, expiry time.Duration) (bool, error)

	// Eval executes a Lua script.
	Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)

	// EvalSha executes a Lua script by SHA.
	EvalSha(ctx context.Context, sha string, keys []string, args ...any) (any, error)

	// ScriptLoad loads a Lua script into Redis.
	ScriptLoad(ctx context.Context, script string) (string, error)

	// Ping checks the connection.
	Ping(ctx context.Context) error

	// Close closes the pool.
	Close() error

	// Name returns a name for this pool (for logging/debugging).
	Name() string
}

// ClientPool wraps a standard Redis client.
type ClientPool struct {
	client *redis.Client
	name   string
}

// NewClientPool creates a pool from a standard Redis client.
func NewClientPool(client *redis.Client) *ClientPool {
	return &ClientPool{
		client: client,
		name:   client.Options().Addr,
	}
}

func (p *ClientPool) SetNX(ctx context.Context, key, value string, expiry time.Duration) (bool, error) {
	result := p.client.SetArgs(ctx, key, value, redis.SetArgs{Mode: "NX", TTL: expiry})
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *ClientPool) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return p.client.Eval(ctx, script, keys, args...).Result()
}

func (p *ClientPool) EvalSha(ctx context.Context, sha string, keys []string, args ...any) (any, error) {
	return p.client.EvalSha(ctx, sha, keys, args...).Result()
}

func (p *ClientPool) ScriptLoad(ctx context.Context, script string) (string, error) {
	return p.client.ScriptLoad(ctx, script).Result()
}

func (p *ClientPool) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

func (p *ClientPool) Close() error {
	return p.client.Close()
}

func (p *ClientPool) Name() string {
	return p.name
}

// Client returns the underlying Redis client.
func (p *ClientPool) Client() *redis.Client {
	return p.client
}

// ClusterPool wraps a Redis cluster client.
type ClusterPool struct {
	client *redis.ClusterClient
	name   string
}

// NewClusterPool creates a pool from a Redis cluster client.
func NewClusterPool(client *redis.ClusterClient) *ClusterPool {
	return &ClusterPool{
		client: client,
		name:   "cluster",
	}
}

func (p *ClusterPool) SetNX(ctx context.Context, key, value string, expiry time.Duration) (bool, error) {
	result := p.client.SetArgs(ctx, key, value, redis.SetArgs{Mode: "NX", TTL: expiry})
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *ClusterPool) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return p.client.Eval(ctx, script, keys, args...).Result()
}

func (p *ClusterPool) EvalSha(ctx context.Context, sha string, keys []string, args ...any) (any, error) {
	return p.client.EvalSha(ctx, sha, keys, args...).Result()
}

func (p *ClusterPool) ScriptLoad(ctx context.Context, script string) (string, error) {
	// For cluster, we need to load on all nodes - this is a simplification
	return p.client.ScriptLoad(ctx, script).Result()
}

func (p *ClusterPool) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

func (p *ClusterPool) Close() error {
	return p.client.Close()
}

func (p *ClusterPool) Name() string {
	return p.name
}

// Client returns the underlying Redis cluster client.
func (p *ClusterPool) Client() *redis.ClusterClient {
	return p.client
}

// FailoverPool wraps a Redis failover (sentinel) client.
type FailoverPool struct {
	client *redis.Client
	name   string
}

// NewFailoverPool creates a pool from a Redis failover client.
func NewFailoverPool(client *redis.Client, masterName string) *FailoverPool {
	return &FailoverPool{
		client: client,
		name:   "sentinel:" + masterName,
	}
}

func (p *FailoverPool) SetNX(ctx context.Context, key, value string, expiry time.Duration) (bool, error) {
	result := p.client.SetArgs(ctx, key, value, redis.SetArgs{Mode: "NX", TTL: expiry})
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *FailoverPool) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return p.client.Eval(ctx, script, keys, args...).Result()
}

func (p *FailoverPool) EvalSha(ctx context.Context, sha string, keys []string, args ...any) (any, error) {
	return p.client.EvalSha(ctx, sha, keys, args...).Result()
}

func (p *FailoverPool) ScriptLoad(ctx context.Context, script string) (string, error) {
	return p.client.ScriptLoad(ctx, script).Result()
}

func (p *FailoverPool) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

func (p *FailoverPool) Close() error {
	return p.client.Close()
}

func (p *FailoverPool) Name() string {
	return p.name
}

// Client returns the underlying Redis client.
func (p *FailoverPool) Client() *redis.Client {
	return p.client
}
