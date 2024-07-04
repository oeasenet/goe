package core

import (
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/modules/cache"
	"go.oease.dev/goe/modules/msearch"
)

type Container struct {
	config      contracts.Config
	logger      contracts.Logger
	mongo       contracts.MongoDB
	meilisearch contracts.Meilisearch
	queue       contracts.Queue
	cache       contracts.Cache
	mailer      contracts.Mailer
	appConfig   *GoeConfig
}

func NewContainer(config contracts.Config, logger contracts.Logger, appConfig *GoeConfig) *Container {
	return &Container{
		config:    config,
		logger:    logger,
		appConfig: appConfig,
	}
}

func (c *Container) InitMongo() {
	// Initialize MongoDB
	mdb, err := NewGoeMongoDB(c.appConfig, c.logger)
	if err != nil {
		c.logger.Panic("Failed to initialize MongoDB: ", err)
		return
	} else {
		c.mongo = mdb
	}
}

func (c *Container) InitMeilisearch() {
	if c.appConfig.Features.MeilisearchEnabled {
		if c.appConfig.Meilisearch.ApiKey == "" {
			c.logger.Panic("meilisearch api key is required")
			return
		}
		if c.appConfig.Meilisearch.Endpoint == "" {
			c.logger.Panic("meilisearch endpoint is required")
			return
		}
		ms := msearch.NewMSearch(c.appConfig.Meilisearch.Endpoint, c.appConfig.Meilisearch.ApiKey, c.logger)
		if ms == nil {
			c.logger.Panic("Failed to initialize Meilisearch")
			return
		}
		c.meilisearch = ms
		if c.appConfig.Features.SearchDBSyncEnabled {
			err := c.mongo.(*GoeMongoDB).SetMeilisearch(ms)
			if err != nil {
				c.logger.Panic("Failed to bind Meilisearch to MongoDB: ", err)
				return
			}
		}
	}
}

func (c *Container) InitCache() {
	// Initialize Cache
	if c.appConfig.Redis.Host != "" && c.appConfig.Redis.Port != 0 && c.appConfig.Redis.Username != "" && c.appConfig.Redis.Password != "" {
		c.cache = cache.NewRedisCache(c.appConfig.Redis.Host, c.appConfig.Redis.Port, c.appConfig.Redis.Username, c.appConfig.Redis.Password, RedisDBCache, c.logger)
		if c.cache == nil {
			c.logger.Panic("Failed to initialize Redis Cache")
		}
	} else {
		c.logger.Panic("Failed to initialize Redis Cache: missing required redis configuration")
		return
	}
}

func (c *Container) InitQueue() {
	// Initialize Queue
	if c.appConfig.Redis.Host != "" && c.appConfig.Redis.Port != 0 && c.appConfig.Redis.Username != "" && c.appConfig.Redis.Password != "" {
		q, err := NewGoeQueue(c.appConfig, c.logger)
		if err != nil {
			c.logger.Panic("Failed to initialize Redis MQ: ", err)
			return
		} else {
			c.queue = q
		}
	} else {
		c.logger.Panic("Failed to initialize Redis MQ: missing required redis configuration")
		return
	}
}

func (c *Container) GetConfig() contracts.Config {
	return c.config
}

func (c *Container) GetMongo() contracts.MongoDB {
	return c.mongo
}

func (c *Container) GetMailer() contracts.Mailer {
	return c.mailer
}

func (c *Container) GetMeilisearch() contracts.Meilisearch {
	return c.meilisearch
}

func (c *Container) GetLogger() contracts.Logger {
	return c.logger
}

func (c *Container) GetQueue() contracts.Queue {
	return c.queue
}

func (c *Container) GetCache() contracts.Cache {
	return c.cache
}
