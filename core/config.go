package core

var goeConfigInstance *GoeConfig

func UseGoeConfig() *GoeConfig {
	return goeConfigInstance
}

type GoeConfig struct {
	App         *AppConfigs
	Features    *GoeConfigFeatures
	MongoDB     *GoeConfigMongodb
	Redis       *GoeConfigRedis
	Meilisearch *GoeConfigMeilisearch
	Mailer      *GoeConfigMailer
	Queue       *GoeConfigQueue
	Http        *GoeConfigHttp
	S3          *GoeConfigS3
	OIDC        *GoeOIDCConfig
}

type AppConfigs struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

type GoeConfigFeatures struct {
	MongoDBEnabled      bool `json:"mongodb_enabled"`
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

type GoeConfigS3 struct {
	Endpoint     string `json:"endpoint"`
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	Bucket       string `json:"bucket"`
	Region       string `json:"region"`
	BucketLookup string `json:"bucket_lookup"`
	UseSSL       bool   `json:"use_ssl"`
	Token        string `json:"token"`
}

type GoeOIDCConfig struct {
	AppId     string   `json:"app_id"`
	AppSecret string   `json:"app_secret"`
	AppScopes []string `json:"app_scopes"`
	Issuer    string   `json:"issuer"`
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
