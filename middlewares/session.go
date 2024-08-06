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
var once sync.Once

func initSessionStore() {
	if sessionStore != nil {
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
		sessionStoreConfig := session.ConfigDefault
		sessionStoreConfig.Expiration = time.Duration(core.UseGoeConfig().Session.Expiration) * time.Second
		sessionStoreConfig.Storage = store
		sessionStoreConfig.KeyLookup = core.UseGoeConfig().Session.KeyLookup
		sessionStore = session.New(sessionStoreConfig)
	})
}

func UseSession(ctx fiber.Ctx) *session.Session {
	initSessionStore()
	s, err := sessionStore.Get(ctx)
	if err != nil {
		return nil
	}
	return s
}

func IsLoggedIn(ctx fiber.Ctx) bool {
	s := UseSession(ctx)
	if s != nil && len(s.ID()) > 0 && len(s.Keys()) > 0 {
		return true
	}
	return false
}
