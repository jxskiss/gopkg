package lru

import (
	"time"
)

var _ Interface = (*Cache)(nil)
var _ Interface = (*ShardedCache)(nil)

type Interface interface {
	Len() int
	Has(key interface{}) (exists, expired bool)
	Get(key interface{}) (v interface{}, exists, expired bool)
	GetQuiet(key interface{}) (v interface{}, exists, expired bool)
	GetNotStale(key interface{}) (v interface{}, exists bool)
	MGet(keys ...interface{}) map[interface{}]interface{}
	MGetNotStale(keys ...interface{}) map[interface{}]interface{}
	MGetInt(keys ...int) map[int]interface{}
	MGetIntNotStale(keys ...int) map[int]interface{}
	MGetInt64(keys ...int64) map[int64]interface{}
	MGetInt64NotStale(keys ...int64) map[int64]interface{}
	MGetUint64(keys ...uint64) map[uint64]interface{}
	MGetUint64NotStale(keys ...uint64) map[uint64]interface{}
	MGetString(keys ...string) map[string]interface{}
	MGetStringNotStale(keys ...string) map[string]interface{}

	Set(key, value interface{}, ttl time.Duration)
	MSet(kvmap interface{}, ttl time.Duration)

	Del(key interface{})
	MDel(keys ...interface{})
	MDelInt(keys ...int)
	MDelInt64(keys ...int64)
	MDelUint64(keys ...uint64)
	MDelString(keys ...string)
}
