package cache

import (
	"fmt"
	"sync"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// manager implements the CacheManager interface
type manager struct {
	config  contract.Config
	stores  map[string]contract.Cache
	drivers map[string]contract.CacheStoreFactory
	mu      sync.RWMutex
}

// NewManager creates a new cache manager
func NewManager(config contract.Config) contract.CacheManager {
	return &manager{
		config:  config,
		stores:  make(map[string]contract.Cache),
		drivers: make(map[string]contract.CacheStoreFactory),
	}
}

// Store returns a cache instance by name
func (m *manager) Store(name ...string) contract.Cache {
	storeName := m.getDefaultStore()
	if len(name) > 0 && name[0] != "" {
		storeName = name[0]
	}

	m.mu.RLock()
	if store, exists := m.stores[storeName]; exists {
		m.mu.RUnlock()
		return store
	}
	m.mu.RUnlock()

	// Create the store
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if store, exists := m.stores[storeName]; exists {
		return store
	}

	// Get store configuration
	storeConfig := m.getStoreConfig(storeName)
	if storeConfig == nil {
		panic(fmt.Sprintf("cache store [%s] is not defined", storeName))
	}

	// Get driver factory
	driverName := m.resolveDriverLocked(storeName, storeConfig)
	factory, exists := m.drivers[driverName]
	if !exists {
		panic(fmt.Sprintf("cache driver [%s] is not supported", driverName))
	}

	// Create store
	cacheStore, err := factory(m.config)
	if err != nil {
		panic(fmt.Sprintf("failed to create cache store [%s]: %v", storeName, err))
	}

	// Create cache instance
	prefix := m.config.GetString("CACHE_PREFIX")
	if prefix == "" {
		prefix = m.config.GetString("APP_NAME")
	}
	if storePrefix := storeConfig.Prefix(); storePrefix != "" {
		prefix = storePrefix
	}

	cache := New(cacheStore, prefix)
	m.stores[storeName] = cache

	return cache
}

// Driver returns the default driver name
func (m *manager) Driver() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	defaultStore := m.getDefaultStore()
	return m.resolveDriverLocked(defaultStore, m.getStoreConfig(defaultStore))
}

// resolveDriverLocked resolves the driver for a store: an explicitly
// configured driver wins; otherwise a store named after a registered driver
// uses that driver (so CACHE_STORE=redis or WithStore("redis") means the
// redis driver, not a silent fall-through to memory); anything else defaults
// to memory. Callers must hold m.mu (read or write).
func (m *manager) resolveDriverLocked(storeName string, sc *storeConfig) string {
	if driver := sc.configuredDriver(); driver != "" {
		return driver
	}
	if _, registered := m.drivers[storeName]; registered {
		return storeName
	}
	return "memory"
}

// storeDriverStatus reports how a store's driver resolves, for startup
// validation: the resolved driver name, whether it was explicitly configured
// (rather than derived from the store name or defaulted), and whether it is
// registered.
func (m *manager) storeDriverStatus(storeName string) (driver string, explicit, registered bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sc := m.getStoreConfig(storeName)
	explicit = sc.configuredDriver() != ""
	driver = m.resolveDriverLocked(storeName, sc)
	_, registered = m.drivers[driver]
	return driver, explicit, registered
}

// Extend registers a custom cache driver
func (m *manager) Extend(driver string, factory contract.CacheStoreFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.drivers[driver] = factory
}

// getDefaultStore returns the default store name
func (m *manager) getDefaultStore() string {
	store := m.config.GetString("CACHE_STORE")
	if store == "" {
		store = "default"
	}
	return store
}

// getStoreConfig returns configuration for a specific store
func (m *manager) getStoreConfig(name string) *storeConfig {
	// Create config from environment variables
	return &storeConfig{
		config: m.config,
		name:   name,
	}
}

// storeConfig implements CacheStoreConfig
type storeConfig struct {
	config contract.Config
	name   string
}

// configuredDriver returns the explicitly configured driver for this store —
// the store-specific key first, then the default driver key — or "" when
// neither is set. The manager layers the registered-driver-name fallback on
// top; see resolveDriverLocked.
func (s *storeConfig) configuredDriver() string {
	driver := s.config.GetString(fmt.Sprintf("CACHE_%s_DRIVER", s.name))
	if driver == "" {
		driver = s.config.GetString("CACHE_DRIVER")
	}
	return driver
}

// Driver returns the driver name, satisfying contract.CacheStoreConfig.
func (s *storeConfig) Driver() string {
	if driver := s.configuredDriver(); driver != "" {
		return driver
	}
	return "memory"
}

// Connection returns connection parameters
func (s *storeConfig) Connection() map[string]any {
	params := make(map[string]any)

	// Get all configuration with prefix CACHE_{NAME}_
	prefix := fmt.Sprintf("CACHE_%s_", s.name)
	allConfig := s.config.All()

	for key, value := range allConfig {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			// Remove prefix and convert to lowercase
			paramKey := key[len(prefix):]
			params[paramKey] = value
		}
	}

	// Also check for generic cache connection parameters
	if len(params) == 0 {
		prefix = "CACHE_"
		for key, value := range allConfig {
			if len(key) > len(prefix) && key[:len(prefix)] == prefix &&
				key != "CACHE_STORE" && key != "CACHE_DRIVER" && key != "CACHE_PREFIX" {
				paramKey := key[len(prefix):]
				params[paramKey] = value
			}
		}
	}

	return params
}

// Prefix returns store-specific prefix
func (s *storeConfig) Prefix() string {
	return s.config.GetString(fmt.Sprintf("CACHE_%s_PREFIX", s.name))
}

// TTL returns default TTL
func (s *storeConfig) TTL() time.Duration {
	ttl := s.config.GetDuration(fmt.Sprintf("CACHE_%s_TTL", s.name))
	if ttl == 0 {
		ttl = s.config.GetDuration("CACHE_TTL")
	}
	if ttl == 0 {
		ttl = 2 * time.Hour
	}
	return ttl
}
