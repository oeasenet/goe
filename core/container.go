package core

//
//import (
//	"github.com/samber/do/v2"
//	"oease.dev/pkg/goe/config"
//	"oease.dev/pkg/goe/logging"
//	"oease.dev/pkg/goe/modules/mongo"
//	"os"
//)
//
//// Container is the interface for putting all global accessible objects together
//type Container interface {
//	UseConfig() config.Config
//	UseFiber()
//	UseMongoDB()
//	UseRedis()
//	UseLogger() logging.Logger
//	UseMq()
//	UseCache()
//	UseSearch()
//	UseMailer()
//	UseEventBus()
//
//	GetFrameworkVersion() string
//}
//
//type container struct {
//	goeConfig *config.GoeConfig
//	logger    logging.Logger
//	config    config.Config
//	injector  *do.RootScope
//}
//
//func NewContainer(cfg *config.GoeConfig) Container {
//	c := &container{}
//	c.goeConfig = cfg
//	c.init()
//	return c
//}
//
//// init initializes the container, take all the necessary steps to make all modules ready.
//func (c *container) init() {
//	// init logger
//	log := logging.New()
//
//	// init config
//	envFolder := os.Getenv("ENV_HOME")
//	if envFolder == "" {
//		envFolder = "."
//	}
//	envCfg := config.New(envFolder, log)
//	// check log level, if different create new log instance
//	if envCfg.GetAppEnv() == "prod" {
//		log.Close()
//		log = logging.New(logging.LevelProd)
//	}
//
//	// base core modules init completed
//
//	// init dependency injection core
//	c.injector = do.New()
//	goePackages := do.Package(
//		do.Lazy(func(injector do.Injector) (*mongo.GoeMongo, error) { return mongo.NewGoeMongo(injector) }),
//	)
//	c.injector.Scope("goe", goePackages)
//}
