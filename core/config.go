package core

type GoeConfig struct {
	App         *AppConfigs
	Features    *GoeConfigFeatures
	MongoDB     *GoeConfigMongodb
	Redis       *GoeConfigRedis
	Meilisearch *GoeConfigMeilisearch
	Mailer      *GoeConfigMailer
	Queue       *GoeConfigQueue
	Http        *GoeConfigHttp
}

type AppConfigs struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

type GoeConfigFeatures struct {
	MeilisearchEnabled  bool `json:"meilisearch_enabled"`
	SearchDBSyncEnabled bool `json:"search_db_sync_enabled"`
	SMTPMailerEnabled   bool `json:"smtp_mailer_enabled"`
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
	ConcurrentWorkers  int `json:"concurrent_workers"`
	FetchInterval      int `json:"fetch_interval"`
	FetchLimit         int `json:"fetch_limit"`
	MaxConsumeDuration int `json:"max_consume_duration"`
	DefaultRetries     int `json:"default_retries"`
}

type GoeConfigHttp struct {
	Port            string   `json:"port"`
	ServerHeader    string   `json:"server_header"`
	BodyLimit       int      `json:"body_limit"`
	Concurrency     int      `json:"concurrency"`
	ProxyHeader     string   `json:"proxy_header"`
	TrustProxyCheck bool     `json:"trust_proxy_check"`
	TrustProxies    []string `json:"trust_proxies"`
	ReduceMemory    bool     `json:"reduce_memory"`
	IPValidation    bool     `json:"ip_validation"`
}
