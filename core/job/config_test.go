package job

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_RedactedTarget(t *testing.T) {
	t.Run("url form reduces to host:port without credentials", func(t *testing.T) {
		c := &Config{RedisURL: "redis://default:s3cret@redis.internal:6380/2"}
		got := c.RedactedTarget()
		assert.Equal(t, "redis.internal:6380", got)
		assert.NotContains(t, got, "s3cret")
	})

	t.Run("url wins over hosts, matching the manager's priority", func(t *testing.T) {
		c := &Config{
			RedisURL:   "redis://redis.internal:6380",
			RedisHosts: []string{"other:6379"},
		}
		assert.Equal(t, "redis.internal:6380", c.RedactedTarget())
	})

	t.Run("invalid url never echoes the raw value", func(t *testing.T) {
		c := &Config{RedisURL: "redis://user:secretpw@host:notaport"}
		got := c.RedactedTarget()
		assert.Equal(t, "<invalid JOB_REDIS_URL>", got)
		assert.NotContains(t, got, "secretpw")
	})

	t.Run("hosts form joins the list", func(t *testing.T) {
		c := &Config{RedisHosts: []string{"a:6379", "b:6379"}}
		assert.Equal(t, "a:6379,b:6379", c.RedactedTarget())
	})

	t.Run("legacy addr is the last resort", func(t *testing.T) {
		c := &Config{RedisAddr: "legacy:6379"}
		assert.Equal(t, "legacy:6379", c.RedactedTarget())
	})
}
