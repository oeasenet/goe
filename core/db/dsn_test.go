package db

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildDSN_PostgresValuesReachTheDriverUnchanged parses the DSN with the
// pgx parser gorm's postgres driver uses: whatever characters a setting holds,
// the driver must see exactly the configured value.
func TestBuildDSN_PostgresValuesReachTheDriverUnchanged(t *testing.T) {
	tests := []struct {
		name, user, password string
	}{
		{"plain", "svc", "plain"},
		{"backslash", "svc", `back\slash`},
		{"single quote", "svc", `it's`},
		{"space", "svc", "with space"},
		{"trailing backslash", "svc", `trailing\`},
		{"equals sign", "svc", "key=value"},
		{"username with space", "svc user", "plain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &MockConfig{}
			for key, value := range map[string]string{
				"DB_DSN":      "",
				"DB_HOST":     "db.internal",
				"DB_PORT":     "6432",
				"DB_DATABASE": "app",
				"DB_USERNAME": tt.user,
				"DB_PASSWORD": tt.password,
				"DB_CHARSET":  "",
				"DB_TIMEZONE": "America/New_York",
				"DB_SSLMODE":  "disable",
			} {
				cfg.On("GetString", key).Return(value)
			}
			dbm := &DatabaseModule{config: cfg}

			dsn, err := dbm.buildDSN("default", "postgres", "DB_")
			require.NoError(t, err)

			parsed, err := pgconn.ParseConfig(dsn)
			require.NoError(t, err)
			assert.Equal(t, tt.password, parsed.Password)
			assert.Equal(t, tt.user, parsed.User)
			assert.Equal(t, "app", parsed.Database)
			assert.Equal(t, "db.internal", parsed.Host)
			assert.Equal(t, uint16(6432), parsed.Port)
			assert.Equal(t, "America/New_York", parsed.RuntimeParams["TimeZone"])
		})
	}
}
