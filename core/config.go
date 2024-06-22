package core

type GoeConfig struct {
	Features    *GoeConfigFeatures
	MongoDB     *GoeConfigMongodb
	Redis       *GoeConfigRedis
	Meilisearch *GoeConfigMeilisearch
	Mailer      *GoeConfigMailer
	Queue       *GoeConfigQueue
}

type GoeConfigFeatures struct {
	MongoDBEnabled       bool `json:"mongo_db_enabled"`
	RedisEnabled         bool `json:"redis_enabled"`
	MeilisearchEnabled   bool `json:"meilisearch_enabled"`
	MSearchDBSyncEnabled bool `json:"m_search_db_sync_enabled"`
	SMTPMailerEnabled    bool `json:"smtp_mailer_enabled"`
	RedisMQEnabled       bool `json:"redis_mq_enabled"`
	CacheEnabled         bool `json:"cache_enabled"`
}

type GoeConfigMongodb struct {
	URI string `json:"uri"`
	DB  string `json:"db"`
}

type GoeConfigRedis struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type GoeConfigMeilisearch struct {
	Endpoint string `json:"endpoint"`
	ApiKey   string `json:"api_key"`
}

type GoeConfigMailer struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Tls       bool   `json:"tls"`
	LocalName string `json:"local_name"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}

type GoeConfigQueue struct {
}
