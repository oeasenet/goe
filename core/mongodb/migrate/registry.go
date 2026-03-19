package migrate

import (
	"fmt"
	"slices"
	"sort"
	"sync"
)

// registry holds all registered migrations
var registry = &migrationRegistry{
	migrations: make(map[int64]*Migration),
}

// migrationRegistry manages the collection of registered migrations
type migrationRegistry struct {
	mu         sync.RWMutex
	migrations map[int64]*Migration
}

// Register registers one or more migrations to the global registry.
// This function is typically called in init() functions.
//
// Example:
//
//	func init() {
//	    migrate.Register(
//	        migrate.New(1, "create_users_indexes").
//	            Up(createUsersIndexes).
//	            Down(dropUsersIndexes),
//	        migrate.New(2, "add_email_unique").
//	            Up(migrate.CreateIndex("users", "email", migrate.Unique())).
//	            Down(migrate.DropIndex("users", "email_1")),
//	    )
//	}
func Register(migrations ...*Migration) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	for _, m := range migrations {
		if m == nil {
			continue
		}

		if m.Version <= 0 {
			return fmt.Errorf("%w: version must be positive, got %d", ErrInvalidVersion, m.Version)
		}

		if existing, ok := registry.migrations[m.Version]; ok {
			return fmt.Errorf("%w: version %d already registered as '%s', cannot register '%s'",
				ErrDuplicateVersion, m.Version, existing.Name, m.Name)
		}

		registry.migrations[m.Version] = m
	}

	return nil
}

// MustRegister registers migrations and panics on error.
// This is a convenience function for use in init() functions.
//
// Example:
//
//	func init() {
//	    migrate.MustRegister(
//	        migrate.New(1, "create_users_indexes").
//	            Up(createUsersIndexes).
//	            Down(dropUsersIndexes),
//	    )
//	}
func MustRegister(migrations ...*Migration) {
	if err := Register(migrations...); err != nil {
		panic(fmt.Sprintf("migrate: failed to register migrations: %v", err))
	}
}

// GetMigrations returns all registered migrations sorted by version
func GetMigrations() []*Migration {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	migrations := make([]*Migration, 0, len(registry.migrations))
	for _, m := range registry.migrations {
		migrations = append(migrations, m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations
}

// GetMigration returns a specific migration by version
func GetMigration(version int64) (*Migration, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	m, ok := registry.migrations[version]
	if !ok {
		return nil, fmt.Errorf("%w: version %d", ErrMigrationNotFound, version)
	}

	return m, nil
}

// GetMigrationsInRange returns migrations within the specified version range (inclusive)
func GetMigrationsInRange(fromVersion, toVersion int64) []*Migration {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	var migrations []*Migration
	for version, m := range registry.migrations {
		if version >= fromVersion && version <= toVersion {
			migrations = append(migrations, m)
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations
}

// GetMigrationsAfter returns all migrations with version greater than the specified version
func GetMigrationsAfter(version int64) []*Migration {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	var migrations []*Migration
	for v, m := range registry.migrations {
		if v > version {
			migrations = append(migrations, m)
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations
}

// GetLatestVersion returns the highest registered migration version
func GetLatestVersion() int64 {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	var latest int64
	for version := range registry.migrations {
		if version > latest {
			latest = version
		}
	}

	return latest
}

// Count returns the number of registered migrations
func Count() int {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	return len(registry.migrations)
}

// Clear removes all registered migrations.
// This is primarily useful for testing.
func Clear() {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.migrations = make(map[int64]*Migration)
}

// HasMigrations returns true if there are any registered migrations
func HasMigrations() bool {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	return len(registry.migrations) > 0
}

// Versions returns all registered migration versions sorted in ascending order
func Versions() []int64 {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	versions := make([]int64, 0, len(registry.migrations))
	for version := range registry.migrations {
		versions = append(versions, version)
	}

	slices.Sort(versions)

	return versions
}
