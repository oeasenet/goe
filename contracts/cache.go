package contracts

import "time"

type Cache interface {
	Get(key string) []byte
	GetBind(key string, bindPtr any) error
	Set(key string, value []byte, expire time.Duration) error
	SetBind(key string, bindPtr any, expire time.Duration) error
}
