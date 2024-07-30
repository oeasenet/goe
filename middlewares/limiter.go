package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/storage/redis/v3"
	"github.com/gofiber/utils/v2"
	"github.com/gookit/goutil/strutil"
	"go.oease.dev/goe"
	"go.oease.dev/goe/core"
	"runtime"
	"time"
)

var rateLimiterStorage *redis.Storage

func NewRateLimiter(qpm int) fiber.Handler {
	if rateLimiterStorage == nil {
		//create storage
		redisPort := goe.UseCfg().GetOrDefaultInt("REDIS_PORT", 6379)
		store := redis.New(redis.Config{
			Host:     goe.UseCfg().GetOrDefaultString("REDIS_HOST", "localhost"),
			Port:     redisPort,
			Username: goe.UseCfg().GetOrDefaultString("REDIS_USERNAME", "default"),
			Password: goe.UseCfg().GetOrDefaultString("REDIS_PASSWORD", ""),
			Database: core.RedisDBRateLimiter,
			PoolSize: 10 * runtime.GOMAXPROCS(0),
		})
		rateLimiterStorage = store
	}
	return limiter.New(limiter.Config{
		Next: func(c fiber.Ctx) bool {
			return goe.UseCfg().Get("APP_ENV") != "prod" || c.IP() == "127.0.0.1" || c.IP() == "::1" || c.IP() == "localhost"
		},
		Max:          qpm,
		KeyGenerator: generateRequestKey,
		Expiration:   1 * time.Minute,
		LimitReached: func(ctx fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many requests")
		},
		Storage:           rateLimiterStorage,
		LimiterMiddleware: limiter.SlidingWindow{},
	})
}

func generateRequestKey(ctx fiber.Ctx) string {
	return strutil.Md5(ctx.IP() + utils.ToString(ctx.Request().Header.UserAgent()) + ctx.Path())
}
