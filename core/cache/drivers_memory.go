package cache

import (
	"time"

	"github.com/gofiber/storage/memory/v2"
	"go.oease.dev/goe/v2/contract"
)

// MemoryStoreFactory creates Fiber memory store instances
func MemoryStoreFactory(config contract.Config) (contract.CacheStore, error) {
	// Get GC interval from config
	gcInterval := config.GetDuration("CACHE_MEMORY_GC_INTERVAL")
	if gcInterval == 0 {
		gcInterval = 10 * time.Second
	}

	return memory.New(memory.Config{
		GCInterval: gcInterval,
	}), nil
}
