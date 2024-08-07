package cache

import (
	"errors"
	"github.com/goccy/go-json"
	"github.com/gofiber/storage/redis/v3"
	"runtime"
	"time"
)

type RedisCache struct {
	store  *redis.Storage
	logger Logger
}

func NewRedisCache(redisHost string, redisPort int, redisUsername string, redisPassword string, redisDB int, logger ...Logger) *RedisCache {
	store := redis.New(redis.Config{
		Host:       redisHost,
		Port:       redisPort,
		Username:   redisUsername,
		Password:   redisPassword,
		Database:   redisDB,
		ClientName: "GOEAppCacheClient",
		Reset:      true,
		PoolSize:   10 * runtime.GOMAXPROCS(0),
	})
	rc := &RedisCache{store: store}
	if len(logger) > 0 && logger[0] != nil {
		rc.logger = logger[0]
	}
	return rc
}

func (r *RedisCache) Get(key string) []byte {
	res, err := r.store.Get(key)
	if err != nil {
		r.logger.Error(err)
		return nil
	}
	return res
}

func (r *RedisCache) GetBind(key string, bindPtr any) error {
	res, err := r.store.Get(key)
	if err != nil {
		r.logger.Error(err)
		return err
	}
	if res == nil {
		bindPtr = nil
		return nil
	}
	return json.Unmarshal(res, bindPtr)
}

func (r *RedisCache) Set(key string, value []byte, expire time.Duration) error {
	return r.store.Set(key, value, expire)
}

func (r *RedisCache) SetBind(key string, bindPtr any, expire time.Duration) error {
	if bindPtr == nil {
		return errors.New("bindPtr is nil")
	}
	b, err := json.Marshal(bindPtr)
	if err != nil {
		r.logger.Error(err)
		return err
	}
	return r.store.Set(key, b, expire)
}

func (r *RedisCache) Delete(key string) error {
	return r.store.Delete(key)
}

func (r *RedisCache) Close() error {
	return r.store.Close()
}
