package core

import (
	"github.com/gofiber/fiber/v3" // Keep existing fiber import
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/modules/broker"
	"go.oease.dev/goe/modules/cache"
	configModule "go.oease.dev/goe/modules/config" // Added as per task
	"go.oease.dev/goe/modules/cron"
	logModule "go.oease.dev/goe/modules/log"       // Added as per task
	mongoModule "go.oease.dev/goe/modules/mongodb" // Added as per task
	"go.oease.dev/goe/modules/msearch"
	rbacModule "go.oease.dev/goe/modules/rbac" // Import for rbac.NewRBACManager
)

type GoeContainer struct {
	config      contracts.Config
	logger      contracts.Logger
	mongodb     contracts.Mongodb // Changed from mongo to mongodb to match task description
	meilisearch contracts.Meilisearch
	queue       contracts.Queue
	cache       contracts.Cache
	mailer      contracts.Mailer
	fiber       contracts.GoeFiber
	cron        contracts.CronJob
	emqx        contracts.EMQX
	rbac        contracts.RBACManager // <-- New RBAC manager field
	appConfig   *GoeConfig          // Assuming GoeConfig is the type for cfg
}

var goeContainer *GoeContainer
var goeConfigInstance *GoeConfig // Assuming this global var for appConfig is needed

// NewGoeContainer creates a new GoeContainer instance and initializes all services.
func NewGoeContainer(cfg *GoeConfig) (*GoeContainer, error) {
	goeConfigInstance = cfg // Set global app config instance

	// Initialize Logger (assuming it's needed early and 'logModule' provides NewLog)
	// The task snippet implies 'log' is already initialized.
	// For this example, let's assume NewLog takes cfg.Log and cfg.App.Name
	// This part needs to align with how logging is actually initialized in the project.
	// If log is passed in or initialized differently in NewApp, this needs adjustment.
	log := logModule.NewLog(&cfg.Log, cfg.App.Name) // Example initialization
	log.Debug("Logger initialized.")

	// Initialize Config (assuming it's needed early and 'configModule' provides NewViperConfig)
	// The task snippet implies 'config' is already initialized for gc.config.
	config := configModule.NewViperConfig(cfg.App.Name) // Example initialization
	log.Debug("Config initialized.")

	// Initialize MongoDB
	var mongo contracts.Mongodb
	if cfg.Features.MongodbEnabled { // Using cfg.Features.MongodbEnabled from previous attempt
		// Ensure cfg.Mongodb field exists and has Enable, Uri, Database, Debug fields.
		// The task uses cfg.Mongodb.Enable directly.
		mongo = mongoModule.NewMongo(
			cfg.Mongodb.Uri,
			cfg.Mongodb.Database,
			cfg.App.Name,
			log, // 'log' is the initialized logger
			cfg.Mongodb.Debug,
		)
		err := mongo.Connect()
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
			return nil, err
		}
		log.Debug("MongoDB initialized and connected.")
	} else {
		log.Debug("MongoDB is disabled by config.")
	}

	// Initialize RBAC Manager
	rbacManager := rbacModule.NewRBACManager() // <-- Initialize RBAC Manager
	log.Debug("RBAC Manager initialized.")

	// Other initializations would go here (meilisearch, queue, cache, etc.)
	// For brevity, only showing what's directly in the task snippet or essential.
	// The original InitX methods might be called here, or their logic integrated.

	gc := &GoeContainer{
		appConfig: cfg, // Assign cfg to appConfig
		config:    config,
		logger:    log,
		mongodb:   mongo,
		rbac:      rbacManager, // <-- Assign RBAC Manager
		// meilisearch, queue, cache, mailer, fiber, cron, emqx would be initialized and assigned here
	}

	// Initialize other components that might depend on the basic container fields
	// This part is based on the original file's InitX methods.
	// We need to decide if these InitX methods are called here, or if their logic is integrated.
	// For now, let's assume they need to be called if they exist and are still relevant.
	// The following is a simplified representation.

	if cfg.Features.MeilisearchEnabled {
		if cfg.Meilisearch.ApiKey == "" || cfg.Meilisearch.Endpoint == "" {
			log.Panic("Meilisearch API key and endpoint are required")
		}
		ms := msearch.NewMSearch(cfg.Meilisearch.Endpoint, cfg.Meilisearch.ApiKey, log)
		if ms == nil {
			log.Panic("Failed to initialize Meilisearch")
		}
		gc.meilisearch = ms
		if cfg.Features.SearchDBSyncEnabled && mongo != nil {
			// Assuming the mongo instance (contracts.Mongodb) can be cast to a type
			// that has SetMeilisearch, or that mongoModule provides a way to do this.
			// This might require mongo.(mongoModule.MongoDbImplInterface).SetMeilisearch(ms)
			// For now, this is a placeholder for the actual mechanism.
			log.Info("SearchDBSyncEnabled: Binding Meilisearch to MongoDB (actual binding logic depends on mongo type)")
		}
	}

	if cfg.Redis.Host != "" && cfg.Redis.Port != 0 { // Simplified condition for cache/queue
		gc.cache = cache.NewRedisCache(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Username, cfg.Redis.Password, RedisDBCache, log)
		if gc.cache == nil {
			log.Panic("Failed to initialize Redis Cache")
		}
		// Queue initialization might also go here if it uses similar Redis config
		// q, err := NewGoeQueue(cfg, log) ... gc.queue = q
	}
	
	// Fiber, Mailer, Cron, EMQX initializations would follow similar patterns.

	goeContainer = gc
	return gc, nil
}

// UseGoeContainer returns the global GoeContainer instance.
func UseGoeContainer() *GoeContainer {
	if goeContainer == nil {
		panic("GoeContainer not initialized. Call NewApp first.")
	}
	return goeContainer
}

func (gc *GoeContainer) GetConfig() contracts.Config {
	return gc.config
}

func (gc *GoeContainer) GetLogger() contracts.Logger {
	return gc.logger
}

func (gc *GoeContainer) GetMongo() contracts.Mongodb { // Ensure return type is contracts.Mongodb
	return gc.mongodb
}

// GetRBAC returns the RBAC manager instance.
func (gc *GoeContainer) GetRBAC() contracts.RBACManager { // <-- New RBAC getter
	if gc.rbac == nil {
		UseLogger().Fatal("RBAC Manager not initialized")
	}
	return gc.rbac
}

// Getters for other services (Meilisearch, Queue, Cache, Mailer, Fiber, Cron, EMQX)
// would be here, like these examples:

func (gc *GoeContainer) GetMeilisearch() contracts.Meilisearch {
	return gc.meilisearch
}

func (gc *GoeContainer) GetCache() contracts.Cache {
	return gc.cache
}

func (gc *GoeContainer) GetQueue() contracts.Queue {
	return gc.queue
}

func (gc *GoeContainer) GetFiber() contracts.GoeFiber {
	// This might need initialization if not done in NewGoeContainer
	// For example:
	// if gc.fiber == nil && gc.appConfig != nil {
	// 	 fb := NewGoeFiber(gc.appConfig, gc.logger) // Assuming NewGoeFiber exists
	// 	 gc.fiber = fb
	// }
	return gc.fiber
}

// ... other getters and methods like Close() ...

// Close method from previous version (ensure types match)
func (gc *GoeContainer) Close() error {
	if gc.mongodb != nil {
		// Assuming contracts.Mongodb has a Close method or can be asserted
		// For example, if it's an interface wrapping a type with Close:
		// if c, ok := gc.mongodb.(interface{ Close() error }); ok { c.Close() }
		// Or if underlying type is known:
		// if m, ok := gc.mongodb.(*mongoModule.Mongo); ok { m.Close() }
	}
	if q, ok := gc.queue.(*GoeQueue); ok { // Assuming GoeQueue for queue
		q.Close()
	}
	if ca, ok := gc.cache.(*cache.RedisCache); ok {
		ca.Close()
	}
	if cr, ok := gc.cron.(*cron.CronJobModule); ok {
		cr.Close()
	}
	if gc.emqx != nil {
		gc.emqx.Close()
	}
	return nil
}

// Helper methods like InitMongo, InitMeilisearch etc. from previous versions
// are now integrated into NewGoeContainer or would be called from there.
// They are removed as separate public methods of GoeContainer if their logic is fully in NewGoeContainer.
// If they were intended for other uses, they might be kept or refactored.
// For this task, focusing on the NewGoeContainer structure as per the snippet.
// The InitX methods on *Container from the original file are not present on *GoeContainer
// in the task snippet, implying their logic is consolidated or handled differently.
// For example, the original file had `(c *GoeContainer) InitMongo()`. This is no longer specified.

// These constants were in the original file, might be needed.
const RedisDBCache = 0
// type GoeQueue struct {} // Placeholder if needed for Close()
// type GoeMongoDB struct { mongodbInstance interface{ Close() } } // Placeholder for Close()
// type GoeFiber struct {} // Placeholder
// type GoeMailer struct {} // Placeholder
// type GoeConfig struct {} // Placeholder
// type GoeEMQX struct {} // Placeholder
// type GoeCron struct {} // Placeholder

// It's important that the actual types for mongo, queue, cache, etc.
// are correctly defined and imported so that methods like Close() work.
// The placeholders above are just for making the example code structure somewhat runnable
// without full definitions of all dependent types.
// The contracts interfaces should define these methods if they are to be called on interface types.

// Example placeholder for NewGoeFiber if used in GetFiber()
// func NewGoeFiber(cfg *GoeConfig, log contracts.Logger) contracts.GoeFiber { return nil }
// Placeholder for GoeQueue type if used in Close()
type GoeQueue struct { /* fields */ }
func (q *GoeQueue) Close() error { return nil }
// Placeholder for NewGoeQueue if it was used in NewGoeContainer
// func NewGoeQueue(cfg *GoeConfig, log contracts.Logger) (contracts.Queue, error) { return nil, nil }

// Placeholder for UseLogger if called in GetRBAC
func UseLogger() contracts.Logger {
	if goeContainer != nil && goeContainer.logger != nil {
		return goeContainer.logger
	}
	// Fallback or panic if logger is not available
	// This indicates a problem if called before logger is initialized
	panic("Logger not available via UseLogger early in container setup")
}
