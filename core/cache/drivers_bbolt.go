package cache

import (
	"fmt"

	"github.com/gofiber/storage/bbolt/v2"
	"go.oease.dev/goe/v2/contract"
)

// buildBboltConfig maps CACHE_BBOLT_* configuration onto the Fiber bbolt
// storage config. Zero values defer to the storage package's defaults
// (fiber.db, bucket fiber_storage, 60s file-lock timeout).
//
// ReadOnly is deliberately not exposed: a read-only cache cannot honor the
// Cache contract's writes.
func buildBboltConfig(config contract.Config) bbolt.Config {
	return bbolt.Config{
		Database: config.GetString("CACHE_BBOLT_DATABASE"),
		Bucket:   config.GetString("CACHE_BBOLT_BUCKET"),
		Timeout:  config.GetDuration("CACHE_BBOLT_TIMEOUT"),
		Reset:    config.GetBool("CACHE_BBOLT_RESET"),
	}
}

// BboltStoreFactory creates Fiber bbolt store instances — a single-file
// embedded store needing no external service. The storage package panics when
// the database cannot be opened (bad path, held file lock); that panic is
// converted into an error so the failure carries the driver's name.
func BboltStoreFactory(config contract.Config) (store contract.CacheStore, err error) {
	defer func() {
		if r := recover(); r != nil {
			store, err = nil, fmt.Errorf("bbolt: opening database failed: %v", r)
		}
	}()
	return bbolt.New(buildBboltConfig(config)), nil
}
