package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"go.oease.dev/goe/core"
	"runtime"
	"sync"
	"time"
)

var sessionStore *session.Store
var sessionMiddleware fiber.Handler
var once sync.Once

func initSessionStore() {
	if sessionStore != nil && sessionMiddleware != nil {
		return
	}
	once.Do(func() {
		store := redis.New(redis.Config{
			Host:     core.UseGoeConfig().Redis.Host,
			Port:     core.UseGoeConfig().Redis.Port,
			Username: core.UseGoeConfig().Redis.Username,
			Password: core.UseGoeConfig().Redis.Password,
			Database: core.RedisDBAuthSession,
			PoolSize: 10 * runtime.GOMAXPROCS(0),
		})

		//create session store
		midw, sStore := session.NewWithStore(session.Config{
			IdleTimeout:     time.Duration(core.UseGoeConfig().Session.Expiration) * time.Second,
			AbsoluteTimeout: 0,
			Storage:         store,
			KeyLookup:       core.UseGoeConfig().Session.KeyLookup,
		})
		sessionStore = sStore
		sessionMiddleware = midw
	})
}

func NewSessionMiddleware() fiber.Handler {
	initSessionStore()
	return sessionMiddleware
}

func GetSessionStore() *session.Store {
	initSessionStore()
	return sessionStore
}

func UseSession(ctx fiber.Ctx) *session.Middleware {
	initSessionStore()
	return session.FromContext(ctx)
}

func IsLoggedIn(ctx fiber.Ctx) bool {
	s := UseSession(ctx)
	if s != nil && s.Session != nil && len(s.Session.ID()) > 0 && len(s.Session.Keys()) > 0 {
		return true
	}
	return false
}
