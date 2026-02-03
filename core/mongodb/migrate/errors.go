package migrate

import (
	"errors"
	"fmt"
)

// Sentinel errors for migration operations.
// Use errors.Is() to check for these specific error types.
var (
	// ErrMigrationNotFound indicates that a requested migration version was not found
	ErrMigrationNotFound = errors.New("migration not found")

	// ErrDuplicateVersion indicates that a migration with the same version already exists
	ErrDuplicateVersion = errors.New("duplicate migration version")

	// ErrDirtyState indicates that the database is in a dirty state from a failed migration
	ErrDirtyState = errors.New("database in dirty state from failed migration")

	// ErrLockAcquisitionFailed indicates that the migration lock could not be acquired
	ErrLockAcquisitionFailed = errors.New("failed to acquire migration lock")

	// ErrLockLost indicates that the migration lock was lost during execution
	ErrLockLost = errors.New("migration lock was lost")

	// ErrChecksumMismatch indicates that a migration's checksum does not match the recorded value
	ErrChecksumMismatch = errors.New("migration checksum mismatch")

	// ErrMigrationTimeout indicates that a migration exceeded its timeout
	ErrMigrationTimeout = errors.New("migration timed out")

	// ErrNoMigrations indicates that no migrations are registered
	ErrNoMigrations = errors.New("no migrations registered")

	// ErrAlreadyApplied indicates that a migration has already been applied
	ErrAlreadyApplied = errors.New("migration already applied")

	// ErrInvalidVersion indicates that an invalid version number was provided
	ErrInvalidVersion = errors.New("invalid migration version")

	// ErrTransactionNotSupported indicates that transactions are not available
	ErrTransactionNotSupported = errors.New("transactions not supported (requires replica set)")
)

// Direction represents the migration direction
type Direction string

const (
	// DirectionUp represents a forward migration
	DirectionUp Direction = "up"

	// DirectionDown represents a rollback migration
	DirectionDown Direction = "down"
)

// MigrationError wraps an error with migration context
type MigrationError struct {
	Version   int64
	Name      string
	Direction Direction
	Cause     error
}

// Error implements the error interface
func (e *MigrationError) Error() string {
	return fmt.Sprintf("migration %d (%s) %s failed: %v", e.Version, e.Name, e.Direction, e.Cause)
}

// Unwrap returns the underlying error
func (e *MigrationError) Unwrap() error {
	return e.Cause
}

// NewMigrationError creates a new MigrationError
func NewMigrationError(version int64, name string, direction Direction, cause error) *MigrationError {
	return &MigrationError{
		Version:   version,
		Name:      name,
		Direction: direction,
		Cause:     cause,
	}
}

// LockError wraps a lock-related error with context
type LockError struct {
	Operation string
	Holder    string
	Cause     error
}

// Error implements the error interface
func (e *LockError) Error() string {
	if e.Holder != "" {
		return fmt.Sprintf("lock %s failed: held by %s: %v", e.Operation, e.Holder, e.Cause)
	}
	return fmt.Sprintf("lock %s failed: %v", e.Operation, e.Cause)
}

// Unwrap returns the underlying error
func (e *LockError) Unwrap() error {
	return e.Cause
}

// NewLockError creates a new LockError
func NewLockError(operation, holder string, cause error) *LockError {
	return &LockError{
		Operation: operation,
		Holder:    holder,
		Cause:     cause,
	}
}

// StateError wraps a state-related error with context
type StateError struct {
	Operation string
	Version   int64
	Cause     error
}

// Error implements the error interface
func (e *StateError) Error() string {
	if e.Version > 0 {
		return fmt.Sprintf("state %s for version %d failed: %v", e.Operation, e.Version, e.Cause)
	}
	return fmt.Sprintf("state %s failed: %v", e.Operation, e.Cause)
}

// Unwrap returns the underlying error
func (e *StateError) Unwrap() error {
	return e.Cause
}

// NewStateError creates a new StateError
func NewStateError(operation string, version int64, cause error) *StateError {
	return &StateError{
		Operation: operation,
		Version:   version,
		Cause:     cause,
	}
}

// ChecksumError provides details about a checksum mismatch
type ChecksumError struct {
	Version  int64
	Name     string
	Expected string
	Actual   string
}

// Error implements the error interface
func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum mismatch for migration %d (%s): expected %s, got %s",
		e.Version, e.Name, e.Expected, e.Actual)
}

// Is allows ChecksumError to match ErrChecksumMismatch with errors.Is
func (e *ChecksumError) Is(target error) bool {
	return target == ErrChecksumMismatch
}

// NewChecksumError creates a new ChecksumError
func NewChecksumError(version int64, name, expected, actual string) *ChecksumError {
	return &ChecksumError{
		Version:  version,
		Name:     name,
		Expected: expected,
		Actual:   actual,
	}
}

// DirtyStateError provides details about a dirty database state
type DirtyStateError struct {
	Version      int64
	Name         string
	ErrorMessage string
	AppliedAt    string
}

// Error implements the error interface
func (e *DirtyStateError) Error() string {
	return fmt.Sprintf("database in dirty state: migration %d (%s) failed at %s: %s",
		e.Version, e.Name, e.AppliedAt, e.ErrorMessage)
}

// Is allows DirtyStateError to match ErrDirtyState with errors.Is
func (e *DirtyStateError) Is(target error) bool {
	return target == ErrDirtyState
}

// NewDirtyStateError creates a new DirtyStateError
func NewDirtyStateError(version int64, name, errMsg, appliedAt string) *DirtyStateError {
	return &DirtyStateError{
		Version:      version,
		Name:         name,
		ErrorMessage: errMsg,
		AppliedAt:    appliedAt,
	}
}
