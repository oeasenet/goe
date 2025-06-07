package types

import "go.oease.dev/goe/v2/contract"

// Version represents the current version of the Goe framework
const Version = "1.0.0"

// Environment represents the application environment
type Environment string

const (
	// EnvLocal represents local environment
	EnvLocal Environment = "local"

	// EnvDev represents development environment
	EnvDev Environment = "dev"

	// EnvTest represents test environment
	EnvTest Environment = "test"

	// EnvStaging represents staging environment
	EnvStaging Environment = "staging"

	// EnvProd represents production environment
	EnvProd Environment = "prod"
)

// String returns the string representation of the environment
func (e Environment) String() string {
	return string(e)
}

// IsDevelopment returns true if the environment is development
func (e Environment) IsDevelopment() bool {
	return e == EnvLocal || e == EnvDev
}

// IsProduction returns true if the environment is production
func (e Environment) IsProduction() bool {
	return e == EnvProd
}

// IsTesting returns true if the environment is testing
func (e Environment) IsTesting() bool {
	return e == EnvTest
}

type field struct {
	key   string
	value any
}

func (f *field) Key() string {
	return f.key
}

func (f *field) Value() any {
	return f.value
}

// NewField creates a new field for structured logging
func NewField(key string, value any) contract.Field {
	return &field{key: key, value: value}
}
