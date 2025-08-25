package mongodb

import (
	"fmt"
	"strings"

	"go.oease.dev/goe/v2/core/internal/configvalidator"
)

// ValidateConfig validates the MongoDB module configuration
func (dbm *DatabaseModule) ValidateConfig() error {
	v := configvalidator.NewConfigValidator(dbm.config, "mongodb")

	// Get default connection name
	defaultConnectionName := dbm.config.GetString("MONGO_DB_CONNECTION")
	if defaultConnectionName == "" {
		defaultConnectionName = "default"
	}

	// Validate default connection
	configPrefix := "MONGO_DB_"
	if defaultConnectionName != "default" {
		configPrefix = "MONGO_DB_" + defaultConnectionName + "_"
	}

	v.Require(configPrefix+"URI", "MongoDB connection URI")
	v.Require(configPrefix+"DB_NAME", "MongoDB database name")

	// Optional connection pool settings
	v.Optional(configPrefix+"MIN_POOL_SIZE", "MongoDB minimum pool size", configvalidator.ValidatePositiveInt)
	v.Optional(configPrefix+"MAX_POOL_SIZE", "MongoDB maximum pool size", configvalidator.ValidatePositiveInt)
	v.Optional(configPrefix+"MAX_CONN_IDLE_TIME", "MongoDB max connection idle time", func(value any) error {
		duration := dbm.config.GetDuration(configPrefix + "MAX_CONN_IDLE_TIME")
		if duration < 0 {
			return fmt.Errorf("MAX_CONN_IDLE_TIME must be positive")
		}
		return nil
	})

	// Validate additional connections if configured
	connectionsList := dbm.config.GetString("MONGO_DB_CONNECTIONS")
	if connectionsList != "" {
		// Parse connection names
		connectionNames := parseConnectionNames(connectionsList)
		for _, connName := range connectionNames {
			if connName == defaultConnectionName {
				continue // Skip default connection (already validated)
			}
			if connName == "" {
				continue // Skip empty names
			}

			// Validate each additional connection
			additionalPrefix := "MONGO_DB_" + connName + "_"
			v.Require(additionalPrefix+"URI", "MongoDB connection URI for "+connName)
			v.Require(additionalPrefix+"DB_NAME", "MongoDB database name for "+connName)

			// Optional settings for additional connections
			v.Optional(additionalPrefix+"MIN_POOL_SIZE", "MongoDB minimum pool size for "+connName, configvalidator.ValidatePositiveInt)
			v.Optional(additionalPrefix+"MAX_POOL_SIZE", "MongoDB maximum pool size for "+connName, configvalidator.ValidatePositiveInt)
			v.Optional(additionalPrefix+"MAX_CONN_IDLE_TIME", "MongoDB max connection idle time for "+connName, func(value any) error {
				duration := dbm.config.GetDuration(additionalPrefix + "MAX_CONN_IDLE_TIME")
				if duration < 0 {
					return fmt.Errorf("MAX_CONN_IDLE_TIME must be positive")
				}
				return nil
			})
		}
	}

	return v.Validate()
}

// parseConnectionNames parses a comma-separated list of connection names
func parseConnectionNames(connectionsList string) []string {
	connectionNames := strings.Split(connectionsList, ",")
	result := make([]string, 0)
	for _, connName := range connectionNames {
		connName = strings.TrimSpace(connName)
		if connName != "" {
			result = append(result, strings.ToUpper(connName))
		}
	}
	return result
}
