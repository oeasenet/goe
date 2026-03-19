package lock

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"go.oease.dev/goe/v2/contract"
)

// URL schemes for auto-detection
const (
	schemeRedis    = "redis"
	schemeRedisTLS = "rediss"
	schemeSentinel = "redis-sentinel"
	schemeCluster  = "redis-cluster"
)

// Manager implements contract.LockManager using Redis as the backend.
// It supports multiple Redis instances for the Redlock algorithm,
// as well as Redis Cluster and Sentinel modes (auto-detected from URL).
type Manager struct {
	pools  []Pool // Redis pools for Redlock algorithm
	config *Config
	logger contract.Logger
	mode   string // Detected mode for logging

	// Stats tracking
	activeLocks   int64
	totalAcquired int64
	totalReleased int64
	totalFailed   int64

	// Mutex management
	mutexes map[string]*Mutex
	mu      sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager creates a new Redis-based lock manager.
// The connection mode is auto-detected from the URL scheme:
//   - redis:// or rediss:// → Single Redis instance
//   - redis-sentinel:// → Redis Sentinel
//   - redis-cluster:// → Redis Cluster
//   - Multiple URLs in RedisURLs → Redlock algorithm
func NewManager(config *Config, logger contract.Logger) (*Manager, error) {
	pools, mode, err := createPools(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis pools: %w", err)
	}

	if len(pools) == 0 {
		return nil, fmt.Errorf("no Redis pools created from configuration")
	}

	// Test connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, pool := range pools {
		if err := pool.Ping(ctx); err != nil {
			// Close already created pools on failure
			for _, p := range pools {
				_ = p.Close()
			}
			return nil, fmt.Errorf("failed to connect to Redis pool %s: %w", pool.Name(), err)
		}
	}

	ctx, cancel = context.WithCancel(context.Background())

	manager := &Manager{
		pools:   pools,
		config:  config,
		logger:  logger,
		mode:    mode,
		mutexes: make(map[string]*Mutex),
		ctx:     ctx,
		cancel:  cancel,
	}

	logger.Info("Lock manager initialized",
		"mode", mode,
		"pools", len(pools),
		"quorum", len(pools)/2+1,
	)

	return manager, nil
}

// createPools creates Redis pools based on configuration with auto-detection.
func createPools(config *Config) ([]Pool, string, error) {
	// Multiple URLs = Redlock mode
	if len(config.RedisURLs) > 0 {
		return createRedlockPools(config)
	}

	// Single URL - detect mode from scheme
	return createPoolFromURL(config.RedisURL, config)
}

// createRedlockPools creates multiple independent pools for Redlock algorithm.
func createRedlockPools(config *Config) ([]Pool, string, error) {
	pools := make([]Pool, 0, len(config.RedisURLs))

	for _, rawURL := range config.RedisURLs {
		pool, mode, err := createPoolFromURL(rawURL, config)
		if err != nil {
			// Close already created pools
			for _, p := range pools {
				_ = p.Close()
			}
			return nil, "", fmt.Errorf("failed to create pool from URL %s: %w", rawURL, err)
		}

		// Redlock requires single Redis instances, not cluster/sentinel
		if mode != "single" {
			for _, p := range pools {
				_ = p.Close()
			}
			for _, p := range pool {
				_ = p.Close()
			}
			return nil, "", fmt.Errorf("redlock mode requires redis:// or rediss:// URLs, got %s", mode)
		}

		pools = append(pools, pool...)
	}

	return pools, "redlock", nil
}

// createPoolFromURL creates pool(s) from a single URL with auto-detection.
// The scheme is detected manually because sentinel and cluster URLs contain commas
// in the host portion, which Go 1.26+'s stricter url.Parse rejects.
func createPoolFromURL(rawURL string, config *Config) ([]Pool, string, error) {
	if rawURL == "" {
		rawURL = "redis://localhost:6379/0"
	}

	// Detect scheme manually to handle sentinel/cluster URLs with commas in host
	scheme, _, ok := strings.Cut(rawURL, "://")
	if !ok {
		return nil, "", fmt.Errorf("invalid Redis URL: missing scheme separator")
	}

	switch scheme {
	case schemeRedis, schemeRedisTLS:
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, "", fmt.Errorf("invalid Redis URL: %w", err)
		}
		return createSinglePool(u, config)
	case schemeSentinel:
		return createSentinelPool(rawURL, config)
	case schemeCluster:
		return createClusterPool(rawURL, config)
	default:
		return nil, "", fmt.Errorf("unsupported URL scheme: %s (supported: redis, rediss, redis-sentinel, redis-cluster)", scheme)
	}
}

// createSinglePool creates a single Redis client pool.
// URL format: redis://[user:pass@]host[:port][/db]
func createSinglePool(u *url.URL, config *Config) ([]Pool, string, error) {
	opts := &redis.Options{
		Addr: u.Host,
	}

	// Default port if not specified
	if !strings.Contains(u.Host, ":") {
		opts.Addr = u.Host + ":6379"
	}

	// TLS for rediss://
	if u.Scheme == schemeRedisTLS {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: config.TLSInsecureSkipVerify,
		}
	}

	// Extract credentials
	if u.User != nil {
		opts.Username = u.User.Username()
		if password, ok := u.User.Password(); ok {
			opts.Password = password
		}
	}

	// Extract database number
	if u.Path != "" && u.Path != "/" {
		dbStr := strings.TrimPrefix(u.Path, "/")
		if db, err := strconv.Atoi(dbStr); err == nil {
			opts.DB = db
		}
	}

	// Apply pool size
	if config.PoolSize > 0 {
		opts.PoolSize = config.PoolSize
	}

	client := redis.NewClient(opts)
	return []Pool{NewClientPool(client)}, "single", nil
}

// createSentinelPool creates a Redis Sentinel failover pool.
// URL format: redis-sentinel://[user:pass@]master@sentinel1:port,sentinel2:port[/db]
// Parses the URL manually because sentinel URLs contain commas in the host
// portion, which Go 1.26+'s stricter url.Parse rejects.
func createSentinelPool(rawURL string, config *Config) ([]Pool, string, error) {
	// Remove scheme prefix
	rest := strings.TrimPrefix(rawURL, schemeSentinel+"://")

	// Extract path (database number) from the end
	var dbPath string
	if slashIdx := strings.LastIndex(rest, "/"); slashIdx != -1 {
		dbPath = rest[slashIdx+1:]
		rest = rest[:slashIdx]
	}

	// Extract userinfo if present.
	// Sentinel URLs may have: user:pass@master@hosts or just master@hosts.
	// If there are 2+ @ signs, the first segment (containing ":") is userinfo.
	var username, password string
	if atCount := strings.Count(rest, "@"); atCount >= 2 {
		firstAt := strings.Index(rest, "@")
		userinfo := rest[:firstAt]
		rest = rest[firstAt+1:]
		if colonIdx := strings.Index(userinfo, ":"); colonIdx != -1 {
			username = userinfo[:colonIdx]
			password = userinfo[colonIdx+1:]
		} else {
			username = userinfo
		}
	}

	// Parse master name from remaining: master@sentinel1:port,sentinel2:port
	hostPart := rest
	masterName := ""

	if atIdx := strings.Index(hostPart, "@"); atIdx != -1 {
		masterName = hostPart[:atIdx]
		hostPart = hostPart[atIdx+1:]
	}

	if masterName == "" {
		return nil, "", fmt.Errorf("sentinel URL must include master name: redis-sentinel://master@sentinel:port")
	}

	// Parse sentinel addresses
	sentinelAddrs := strings.Split(hostPart, ",")
	if len(sentinelAddrs) == 0 {
		return nil, "", fmt.Errorf("no sentinel addresses provided")
	}

	// Add default port if missing
	for i, addr := range sentinelAddrs {
		if !strings.Contains(addr, ":") {
			sentinelAddrs[i] = addr + ":26379"
		}
	}

	opts := &redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Username:      username,
		Password:      password,
	}

	// Extract database number
	if dbPath != "" {
		if db, err := strconv.Atoi(dbPath); err == nil {
			opts.DB = db
		}
	}

	// Apply pool size
	if config.PoolSize > 0 {
		opts.PoolSize = config.PoolSize
	}

	// TLS if insecure skip is configured (implies TLS wanted)
	if config.TLSInsecureSkipVerify {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client := redis.NewFailoverClient(opts)
	return []Pool{NewFailoverPool(client, masterName)}, "sentinel", nil
}

// createClusterPool creates a Redis Cluster pool.
// URL format: redis-cluster://[user:pass@]node1:port,node2:port
// Parses the URL manually because cluster URLs contain commas in the host
// portion, which Go 1.26+'s stricter url.Parse rejects.
func createClusterPool(rawURL string, config *Config) ([]Pool, string, error) {
	// Remove scheme prefix
	rest := strings.TrimPrefix(rawURL, schemeCluster+"://")

	// Extract path (not typically used for cluster, but strip it)
	if slashIdx := strings.LastIndex(rest, "/"); slashIdx != -1 {
		rest = rest[:slashIdx]
	}

	// Extract userinfo if present (user:pass@hosts)
	var username, password string
	if atIdx := strings.Index(rest, "@"); atIdx != -1 {
		userinfo := rest[:atIdx]
		rest = rest[atIdx+1:]
		if colonIdx := strings.Index(userinfo, ":"); colonIdx != -1 {
			username = userinfo[:colonIdx]
			password = userinfo[colonIdx+1:]
		} else {
			username = userinfo
		}
	}

	// Parse cluster node addresses
	addrs := strings.Split(rest, ",")
	if len(addrs) == 0 {
		return nil, "", fmt.Errorf("no cluster addresses provided")
	}

	// Add default port if missing
	for i, addr := range addrs {
		if !strings.Contains(addr, ":") {
			addrs[i] = addr + ":6379"
		}
	}

	opts := &redis.ClusterOptions{
		Addrs:    addrs,
		Username: username,
		Password: password,
	}

	// Apply pool size
	if config.PoolSize > 0 {
		opts.PoolSize = config.PoolSize
	}

	// TLS if insecure skip is configured (implies TLS wanted)
	if config.TLSInsecureSkipVerify {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client := redis.NewClusterClient(opts)
	return []Pool{NewClusterPool(client)}, "cluster", nil
}

// NewMutex creates a new distributed mutex with the given name.
// The mutex uses the Redlock algorithm when multiple pools are available.
func (m *Manager) NewMutex(name string, opts ...contract.MutexOption) contract.Mutex {
	// Apply key prefix
	fullName := m.config.DefaultKeyPrefix + name

	// Merge default options with provided options
	defaultOpts := []contract.MutexOption{
		contract.WithExpiry(m.config.DefaultExpiry),
		contract.WithTries(m.config.DefaultTries),
		contract.WithRetryDelay(m.config.DefaultRetryDelay),
		contract.WithDriftFactor(m.config.DefaultDriftFactor),
	}

	// User options override defaults
	allOpts := append(defaultOpts, opts...)

	// Create mutex with all pools
	mutex := NewMutex(fullName, m.pools, m.logger, allOpts...)

	// Store reference for tracking
	m.mu.Lock()
	m.mutexes[fullName] = mutex
	m.mu.Unlock()

	// Wrap with stats tracking
	return &trackedMutex{
		Mutex:   mutex,
		manager: m,
	}
}

// Health checks the health of the Redis backend(s).
// For Redlock mode, returns success if quorum of pools is healthy.
func (m *Manager) Health(ctx context.Context) error {
	healthyCount := 0
	var lastErr error

	for _, pool := range m.pools {
		if err := pool.Ping(ctx); err != nil {
			lastErr = err
			m.logger.Warn("Redis pool unhealthy",
				"pool", pool.Name(),
				"error", err,
			)
		} else {
			healthyCount++
		}
	}

	// Require quorum for health
	quorum := len(m.pools)/2 + 1
	if healthyCount < quorum {
		return fmt.Errorf("insufficient healthy pools: %d/%d (quorum: %d), last error: %w",
			healthyCount, len(m.pools), quorum, lastErr)
	}

	return nil
}

// Stats returns statistics about the lock manager.
func (m *Manager) Stats(ctx context.Context) (*contract.LockStats, error) {
	// Measure latency against first pool
	start := time.Now()
	if len(m.pools) > 0 {
		if err := m.pools[0].Ping(ctx); err != nil {
			return nil, err
		}
	}
	latency := time.Since(start)

	return &contract.LockStats{
		ActiveLocks:    atomic.LoadInt64(&m.activeLocks),
		TotalAcquired:  atomic.LoadInt64(&m.totalAcquired),
		TotalReleased:  atomic.LoadInt64(&m.totalReleased),
		TotalFailed:    atomic.LoadInt64(&m.totalFailed),
		BackendLatency: latency,
	}, nil
}

// Close closes the lock manager and releases all resources.
func (m *Manager) Close(ctx context.Context) error {
	m.cancel()

	var lastErr error
	for _, pool := range m.pools {
		if err := pool.Close(); err != nil {
			lastErr = err
			m.logger.Error("Error closing Redis pool",
				"pool", pool.Name(),
				"error", err,
			)
		}
	}

	return lastErr
}

// Pools returns the underlying Redis pools (for testing/debugging).
func (m *Manager) Pools() []Pool {
	return m.pools
}

// Mode returns the detected connection mode.
func (m *Manager) Mode() string {
	return m.mode
}

// recordAcquire records a successful lock acquisition.
func (m *Manager) recordAcquire() {
	atomic.AddInt64(&m.activeLocks, 1)
	atomic.AddInt64(&m.totalAcquired, 1)
}

// recordRelease records a lock release.
func (m *Manager) recordRelease() {
	atomic.AddInt64(&m.activeLocks, -1)
	atomic.AddInt64(&m.totalReleased, 1)
}

// recordFailed records a failed lock attempt.
func (m *Manager) recordFailed() {
	atomic.AddInt64(&m.totalFailed, 1)
}

// trackedMutex wraps a Mutex to track statistics.
type trackedMutex struct {
	*Mutex
	manager *Manager
}

// Lock acquires the lock with stats tracking.
func (t *trackedMutex) Lock(ctx context.Context) error {
	err := t.Mutex.Lock(ctx)
	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			t.manager.recordFailed()
		}
		return err
	}
	t.manager.recordAcquire()
	return nil
}

// TryLock attempts to acquire the lock with stats tracking.
func (t *trackedMutex) TryLock(ctx context.Context) (bool, error) {
	acquired, err := t.Mutex.TryLock(ctx)
	if err != nil {
		t.manager.recordFailed()
		return false, err
	}
	if acquired {
		t.manager.recordAcquire()
	}
	return acquired, nil
}

// Unlock releases the lock with stats tracking.
func (t *trackedMutex) Unlock(ctx context.Context) error {
	err := t.Mutex.Unlock(ctx)
	if err != nil {
		return err
	}
	t.manager.recordRelease()
	return nil
}
