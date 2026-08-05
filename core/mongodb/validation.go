package mongodb

import (
	"errors"
	"fmt"

	"go.oease.dev/goe/v2/core/internal/configvalidator"
)

// ValidateConfig validates the MongoDB module configuration
func (dbm *DatabaseModule) ValidateConfig() error {
	// Option failures are reported first and abort startup. The module fell
	// back to the environment-only configuration when this happened, so
	// returning here guarantees a half-configured module is never served.
	if len(dbm.optErrs) > 0 {
		return errors.Join(dbm.optErrs...)
	}

	v := configvalidator.NewConfigValidator(dbm.config, "mongodb")

	v.Require("MONGO_URI", "MongoDB connection URI")
	v.Require("MONGO_DB_NAME", "MongoDB database name")

	// Optional connection pool settings
	v.Optional("MONGO_MIN_POOL_SIZE", "MongoDB minimum pool size", configvalidator.ValidatePositiveInt)
	v.Optional("MONGO_MAX_POOL_SIZE", "MongoDB maximum pool size", configvalidator.ValidatePositiveInt)
	v.Optional("MONGO_MAX_CONN_IDLE_TIME", "MongoDB max connection idle time", func(value any) error {
		duration := dbm.config.GetDuration("MONGO_MAX_CONN_IDLE_TIME")
		if duration < 0 {
			return fmt.Errorf("MAX_CONN_IDLE_TIME must be positive")
		}
		return nil
	})

	return v.Validate()
}
