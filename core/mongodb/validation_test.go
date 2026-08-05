package mongodb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig_ValidConfiguration(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("MONGO_URI", "mongodb://localhost:27017")
	cfg.Set("MONGO_DB_NAME", "testdb")

	dbm := NewDBModule(cfg, &nopLogger{})
	assert.NoError(t, dbm.ValidateConfig())
}

func TestValidateConfig_MissingRequiredFields(t *testing.T) {
	t.Run("missing everything", func(t *testing.T) {
		dbm := NewDBModule(newMapConfig(), &nopLogger{})
		err := dbm.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MONGO_URI")
		assert.Contains(t, err.Error(), "MONGO_DB_NAME")
	})

	t.Run("missing db name", func(t *testing.T) {
		cfg := newMapConfig()
		cfg.Set("MONGO_URI", "mongodb://localhost:27017")
		dbm := NewDBModule(cfg, &nopLogger{})
		err := dbm.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MONGO_DB_NAME")
	})
}

func TestValidateConfig_OptionalPoolSettings(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("MONGO_URI", "mongodb://localhost:27017")
	cfg.Set("MONGO_DB_NAME", "testdb")
	cfg.Set("MONGO_MIN_POOL_SIZE", 5)
	cfg.Set("MONGO_MAX_POOL_SIZE", 50)
	cfg.Set("MONGO_MAX_CONN_IDLE_TIME", time.Minute)

	dbm := NewDBModule(cfg, &nopLogger{})
	assert.NoError(t, dbm.ValidateConfig())
}

func TestValidateConfig_InvalidPoolSettings(t *testing.T) {
	cfg := newMapConfig()
	cfg.Set("MONGO_URI", "mongodb://localhost:27017")
	cfg.Set("MONGO_DB_NAME", "testdb")
	cfg.Set("MONGO_MIN_POOL_SIZE", -1)

	dbm := NewDBModule(cfg, &nopLogger{})
	assert.Error(t, dbm.ValidateConfig())
}
