package cache

import (
	"fmt"

	"github.com/gofiber/storage/badger/v2"
	"go.oease.dev/goe/v2/contract"
)

// buildBadgerConfig maps CACHE_BADGER_* configuration onto the Fiber Badger
// storage config. Zero values defer to the storage package's defaults
// (./fiber.badger, 10s GC interval).
func buildBadgerConfig(config contract.Config) badger.Config {
	return badger.Config{
		Database:   config.GetString("CACHE_BADGER_DATABASE"),
		Reset:      config.GetBool("CACHE_BADGER_RESET"),
		GCInterval: config.GetDuration("CACHE_BADGER_GC_INTERVAL"),
	}
}

// BadgerStoreFactory creates Fiber Badger store instances — an embedded,
// disk-backed store needing no external service. The storage package panics
// when the database cannot be opened (bad path, held lock); that panic is
// converted into an error so the failure carries the driver's name.
func BadgerStoreFactory(config contract.Config) (store contract.CacheStore, err error) {
	defer func() {
		if r := recover(); r != nil {
			store, err = nil, fmt.Errorf("badger: opening database failed: %v", r)
		}
	}()
	return badger.New(buildBadgerConfig(config)), nil
}
