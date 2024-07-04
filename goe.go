package goe

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/core"
	"go.oease.dev/goe/modules/config"
	"go.oease.dev/goe/modules/log"
)

type App struct {
	configs   *core.GoeConfig
	container *core.Container
}

var appInstance *App

func NewApp() error {
	configModule := config.New("./configs")
	appEnv := configModule.GetOrDefaultString("APP_ENV", "dev")
	var logModule *log.Log
	if appEnv == "dev" {
		logModule = log.New(log.LevelDev)
	} else {
		logModule = log.New(log.LevelProd)
	}
	app := &App{}
	err := app.applyEnvConfig(configModule)
	if err != nil {
		return err
	}
	app.container = core.NewContainer(configModule, logModule, app.configs)
	appInstance = app

	// Initialize MongoDB
	appInstance.container.InitMongo()

	// Initialize Meilisearch
	if appInstance.configs.Features.MeilisearchEnabled {
		appInstance.container.InitMeilisearch()
	}

	// Init Queue
	appInstance.container.InitQueue()

	// Init Cache
	appInstance.container.InitCache()

	// Init Mailer
	if appInstance.configs.Features.SMTPMailerEnabled {
		appInstance.container.InitMailer()
	}

	// Init Fiber
	appInstance.container.InitFiber()

	return nil
}

// applyEnvConfig applies environment configuration to the App instance.
// It populates the configs field with values from the configModule parameter.
// It returns an error if there is an issue applying the configuration.
func (app *App) applyEnvConfig(configModule *config.Config) error {
	app.configs = &core.GoeConfig{
		App: &core.AppConfigs{
			Name:    configModule.GetOrDefaultString("APP_NAME", "GoeApp"),
			Version: configModule.GetOrDefaultString("APP_VERSION", "v1.0.0"),
			Env:     configModule.GetOrDefaultString("APP_ENV", "dev"),
		},
		Features: &core.GoeConfigFeatures{
			MeilisearchEnabled:  configModule.GetOrDefaultBool("MEILISEARCH_ENABLED", false),
			SearchDBSyncEnabled: configModule.GetOrDefaultBool("MEILISEARCH_DB_SYNC", false),
			SMTPMailerEnabled:   configModule.GetOrDefaultBool("SMTP_MAILER_ENABLED", false),
		},
		MongoDB: &core.GoeConfigMongodb{
			URI: configModule.GetOrDefaultString("MONGODB_URI", ""),
			DB:  configModule.GetOrDefaultString("MONGODB_DB", ""),
		},
		Redis: &core.GoeConfigRedis{
			Host:     configModule.GetOrDefaultString("REDIS_HOST", ""),
			Port:     configModule.GetOrDefaultInt("REDIS_PORT", 0),
			Username: configModule.GetOrDefaultString("REDIS_USERNAME", ""),
			Password: configModule.GetOrDefaultString("REDIS_PASSWORD", ""),
		},
		Meilisearch: &core.GoeConfigMeilisearch{
			Endpoint: configModule.GetOrDefaultString("MEILISEARCH_ENDPOINT", ""),
			ApiKey:   configModule.GetOrDefaultString("MEILISEARCH_API_KEY", ""),
		},
		Mailer: &core.GoeConfigMailer{
			Host:      configModule.GetOrDefaultString("SMTP_HOST", ""),
			Port:      configModule.GetOrDefaultInt("SMTP_PORT", 0),
			Username:  configModule.GetOrDefaultString("SMTP_USERNAME", ""),
			Password:  configModule.GetOrDefaultString("SMTP_PASSWORD", ""),
			Tls:       configModule.GetOrDefaultBool("SMTP_TLS", false),
			LocalName: configModule.GetOrDefaultString("SMTP_LOCAL_NAME", ""),
			FromEmail: configModule.GetOrDefaultString("SMTP_FROM_EMAIL", ""),
			FromName:  configModule.GetOrDefaultString("SMTP_FROM_NAME", ""),
		},
		Queue: &core.GoeConfigQueue{
			ConcurrentWorkers:  configModule.GetOrDefaultInt("QUEUE_CONCURRENCY", 1),
			FetchInterval:      configModule.GetOrDefaultInt("QUEUE_FETCH_INTERVAL", 1),
			FetchLimit:         configModule.GetOrDefaultInt("QUEUE_FETCH_LIMIT", 0),
			MaxConsumeDuration: configModule.GetOrDefaultInt("QUEUE_MAX_CONSUME_DURATION", 5),
			DefaultRetries:     configModule.GetOrDefaultInt("QUEUE_DEFAULT_RETRIES", 3),
		},
		Http: &core.GoeConfigHttp{
			Port:            configModule.GetOrDefaultString("HTTP_PORT", "3000"),
			ServerHeader:    configModule.GetOrDefaultString("HTTP_SERVER_HEADER", "GoeAppServer/v1"),
			BodyLimit:       configModule.GetOrDefaultInt("HTTP_BODY_LIMIT", fiber.DefaultBodyLimit),
			Concurrency:     configModule.GetOrDefaultInt("HTTP_CONCURRENCY", fiber.DefaultConcurrency),
			ProxyHeader:     configModule.GetOrDefaultString("HTTP_PROXY_HEADER", ""),
			TrustProxyCheck: configModule.GetOrDefaultBool("HTTP_TRUSTED_PROXY_CHECK", false),
			TrustProxies:    configModule.GetStringSlice("HTTP_TRUSTED_PROXIES"),
			IPValidation:    configModule.GetOrDefaultBool("HTTP_IP_VALIDATION", false),
			ReduceMemory:    configModule.GetOrDefaultBool("HTTP_REDUCE_MEMORY", false),
		},
	}
	return nil
}

func UseDB() contracts.MongoDB {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return nil
	}
	return appInstance.container.GetMongo()
}

func UseLog() contracts.Logger {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return nil
	}
	return appInstance.container.GetLogger()
}

func UseCfg() contracts.Config {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return nil
	}
	return appInstance.container.GetConfig()
}

func UseMQ() contracts.Queue {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return nil
	}
	return appInstance.container.GetQueue()
}

func UseCache() contracts.Cache {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return nil
	}
	return appInstance.container.GetCache()
}

func UseSearch() contracts.Meilisearch {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return nil
	}
	return appInstance.container.GetMeilisearch()
}

func UseMailer() contracts.Mailer {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return nil
	}
	return appInstance.container.GetMailer()
}

func UseFiber() contracts.GoeFiber {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return nil
	}
	return appInstance.container.GetFiber()
}

func Close() {
	if appInstance == nil {
		panic("must initialize App first, by calling NewApp() method")
		return
	}
	appInstance.container.Close()
}
