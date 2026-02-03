package migrate

import (
	"context"
	"crypto/sha256"
	"fmt"
	"reflect"
	"runtime"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MigrationFunc is the function signature for migration operations
type MigrationFunc func(ctx context.Context, db *mongo.Database) error

// Migration represents a single database migration
type Migration struct {
	// Version is the unique version number for this migration
	Version int64

	// Name is a descriptive name for the migration
	Name string

	// up is the function to run when migrating up
	up MigrationFunc

	// down is the function to run when migrating down
	down MigrationFunc

	// checksum is computed from the function pointers for verification
	checksum string
}

// New creates a new migration with the given version and name.
// Use the fluent builder methods Up() and Down() to set the migration functions.
//
// Example:
//
//	migrate.New(1, "create_users_indexes").
//	    Up(func(ctx context.Context, db *mongo.Database) error {
//	        // create indexes
//	        return nil
//	    }).
//	    Down(func(ctx context.Context, db *mongo.Database) error {
//	        // drop indexes
//	        return nil
//	    })
func New(version int64, name string) *Migration {
	return &Migration{
		Version: version,
		Name:    name,
	}
}

// Up sets the up migration function and returns the migration for chaining
func (m *Migration) Up(fn MigrationFunc) *Migration {
	m.up = fn
	m.updateChecksum()
	return m
}

// Down sets the down migration function and returns the migration for chaining
func (m *Migration) Down(fn MigrationFunc) *Migration {
	m.down = fn
	m.updateChecksum()
	return m
}

// HasUp returns true if the migration has an up function defined
func (m *Migration) HasUp() bool {
	return m.up != nil
}

// HasDown returns true if the migration has a down function defined
func (m *Migration) HasDown() bool {
	return m.down != nil
}

// RunUp executes the up migration
func (m *Migration) RunUp(ctx context.Context, db *mongo.Database) error {
	if m.up == nil {
		return nil
	}
	return m.up(ctx, db)
}

// RunDown executes the down migration
func (m *Migration) RunDown(ctx context.Context, db *mongo.Database) error {
	if m.down == nil {
		return nil
	}
	return m.down(ctx, db)
}

// Checksum returns the checksum of the migration functions
func (m *Migration) Checksum() string {
	return m.checksum
}

// updateChecksum computes and updates the checksum based on function pointers
func (m *Migration) updateChecksum() {
	h := sha256.New()

	// Write version and name
	h.Write([]byte(fmt.Sprintf("%d:%s:", m.Version, m.Name)))

	// Add function pointer info to checksum
	if m.up != nil {
		h.Write([]byte(getFuncName(m.up)))
	}
	if m.down != nil {
		h.Write([]byte(getFuncName(m.down)))
	}

	m.checksum = fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// getFuncName returns the name of a function for checksum purposes
func getFuncName(fn any) string {
	if fn == nil {
		return ""
	}
	ptr := reflect.ValueOf(fn).Pointer()
	f := runtime.FuncForPC(ptr)
	if f == nil {
		return fmt.Sprintf("unknown_%d", ptr)
	}
	return f.Name()
}

// String returns a string representation of the migration
func (m *Migration) String() string {
	hasUp := "no"
	if m.HasUp() {
		hasUp = "yes"
	}
	hasDown := "no"
	if m.HasDown() {
		hasDown = "yes"
	}
	return fmt.Sprintf("Migration{version=%d, name=%s, up=%s, down=%s}",
		m.Version, m.Name, hasUp, hasDown)
}
