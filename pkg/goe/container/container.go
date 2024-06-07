package container

// Container is the interface for putting all global accessible objects together
type Container interface {
	UseFiber()
	UseMongoDB()
	UseRedis()
	UseLogger()
	UseMq()
	UseCache()
	UseSearch()
	UseMailer()
	UseEventBus()
}
