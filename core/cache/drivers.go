package cache

import "go.oease.dev/goe/v2/contract"

// builtinDrivers returns the factories every module starts with. Custom
// drivers registered through WithCustomDriver are layered on top and may
// replace a builtin by reusing its name.
func builtinDrivers() map[string]contract.CacheStoreFactory {
	return map[string]contract.CacheStoreFactory{
		"memory": MemoryStoreFactory,
		"redis":  RedisStoreFactory,
		"badger": BadgerStoreFactory,
		"bbolt":  BboltStoreFactory,
	}
}
