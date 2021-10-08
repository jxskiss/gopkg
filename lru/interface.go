package lru

import "time"

var _ Interface = (*Cache)(nil)
var _ Interface = (*ShardedCache)(nil)

// Interface is an abstract interface of the LRU cache implemented in this package.
type Interface interface {

	// Len returns the number of cached values.
	Len() int

	// Has checks if a key is in the cache and whether it is expired,
	// without updating its LRU score.
	Has(key interface{}) (exists, expired bool)

	// Get returns the cached value for the given key and updates its LRU score.
	// The returned value may be expired, caller can check the returned value
	// "expired" to check whether the value is expired.
	Get(key interface{}) (v interface{}, exists, expired bool)

	// GetWithTTL returns the cached value for the given key and updates its
	// LRU score. The returned value may be expired, caller can check the
	// returned value "ttl" to check whether the value is expired.
	GetWithTTL(key interface{}) (v interface{}, exists bool, ttl *time.Duration)

	// GetQuiet returns the cached value for the given key, but don't modify its LRU score.
	// The returned value may be expired, caller can check the returned value
	// "expired" to check whether the value is expired.
	GetQuiet(key interface{}) (v interface{}, exists, expired bool)

	// GetNotStale returns the cached value for the given key. The returned value
	// is guaranteed not expired. If unexpired value available, its LRU score
	// will be updated.
	GetNotStale(key interface{}) (v interface{}, exists bool)

	// MGet returns map of cached values for the given interface keys and
	// update their LRU scores. The returned values may be expired.
	// It's a convenient and efficient way to retrieve multiple values.
	MGet(keys ...interface{}) map[interface{}]interface{}

	// MGetNotStale is similar to MGet, but it returns only not stale values.
	MGetNotStale(keys ...interface{}) map[interface{}]interface{}

	// MGetInt returns map of cached values for the given int keys and
	// update their LRU scores. The returned values may be expired.
	// It's a convenient and efficient way to retrieve multiple values for
	// int keys.
	MGetInt(keys ...int) map[int]interface{}

	// MGetIntNotStale is similar to MGetInt, but it returns only not stale values.
	MGetIntNotStale(keys ...int) map[int]interface{}

	// MGetInt64 returns map of cached values for the given int64 keys and
	// update their LRU scores. The returned values may be expired.
	// It's a convenient and efficient way to retrieve multiple values for
	// int64 keys.
	MGetInt64(keys ...int64) map[int64]interface{}

	// MGetInt64NotStale is similar to MGetInt64, but it returns only not stale values.
	MGetInt64NotStale(keys ...int64) map[int64]interface{}

	// MGetUint64 returns map of cached values for the given uint64 keys and
	// update their LRU scores. The returned values may be expired.
	// It's a convenient and efficient way to retrieve multiple values for
	// uint64 keys.
	MGetUint64(keys ...uint64) map[uint64]interface{}

	// MGetUint64NotStale is similar to MGetUint64, but it returns only not stale values.
	MGetUint64NotStale(keys ...uint64) map[uint64]interface{}

	// MGetString returns map of cached values for the given string keys and
	// update their LRU scores. The returned values may be expired.
	// It's a convenient and efficient way to retrieve multiple values for
	// string keys.
	MGetString(keys ...string) map[string]interface{}

	// MGetStringNotStale is similar to MGetString, but it returns only not stale values.
	MGetStringNotStale(keys ...string) map[string]interface{}

	// Set adds an item to the cache overwriting existing one if it exists.
	Set(key, value interface{}, ttl time.Duration)

	// MSet adds multiple items to the cache overwriting existing ones.
	// Unlike calling Set multiple times, it acquires lock only once for
	// multiple key-value pairs.
	MSet(kvmap interface{}, ttl time.Duration)

	// Del removes a key from the cache if it exists.
	Del(key interface{})

	// MDel removes multiple interface keys from the cache if exists.
	// It's a convenient and efficient way to remove multiple keys.
	MDel(keys ...interface{})

	// MDelInt removes multiple int keys from the cache if exists.
	// It's a convenient and efficient way to remove multiple int keys.
	MDelInt(keys ...int)

	// MDelInt64 removes multiple int64 keys from the cache if exists.
	// It's a convenient and efficient way to remove multiple int64 keys.
	MDelInt64(keys ...int64)

	// MDelUint64 removes multiple uint64 keys from the cache if exists.
	// It's a convenient and efficient way to remove multiple uint64 keys.
	MDelUint64(keys ...uint64)

	// MDelString removes multiple string keys from the cache if exists.
	// It's a convenient and efficient way to remove multiple string keys.
	MDelString(keys ...string)
}
