package lru

import "time"

var _ Interface[int, string] = (*Cache[int, string])(nil)
var _ Interface[string, string] = (*ShardedCache[string, string])(nil)

// Interface is a generic abstract interface of the LRU cache implemented
// in this package.
type Interface[K comparable, V any] interface {

	// Len returns the number of cached values.
	Len() int

	// Has checks if a key is in the cache and whether it is expired,
	// without updating its LRU score.
	Has(key K) (exists, expired bool)

	// Get returns the cached value for the given key and updates its LRU score.
	// The returned value may be expired, caller can check the returned value
	// "expired" to check whether the value is expired.
	Get(key K) (v V, exists, expired bool)

	// GetWithTTL returns the cached value for the given key and updates its
	// LRU score. The returned value may be expired, caller can check the
	// returned value "ttl" to check whether the value is expired.
	GetWithTTL(key K) (v V, exists bool, ttl *time.Duration)

	// GetQuiet returns the cached value for the given key, but don't modify its LRU score.
	// The returned value may be expired, caller can check the returned value
	// "expired" to check whether the value is expired.
	GetQuiet(key K) (v V, exists, expired bool)

	// GetNotStale returns the cached value for the given key. The returned value
	// is guaranteed not expired. If unexpired value available, its LRU score
	// will be updated.
	GetNotStale(key K) (v V, exists bool)

	// MGet returns map of cached values for the given interface keys and
	// update their LRU scores. The returned values may be expired.
	// It's a convenient and efficient way to retrieve multiple values.
	MGet(keys ...K) map[K]V

	// MGetNotStale is similar to MGet, but it returns only not stale values.
	MGetNotStale(keys ...K) map[K]V

	// Set adds an item to the cache overwriting existing one if it exists.
	Set(key K, value V, ttl time.Duration)

	// MSet adds multiple items to the cache overwriting existing ones.
	// Unlike calling Set multiple times, it acquires lock only once for
	// multiple key-value pairs.
	MSet(kvmap map[K]V, ttl time.Duration)

	// Delete removes a key from the cache if it exists.
	Delete(key K)

	// MDelete removes multiple interface keys from the cache if exists.
	// It's a convenient and efficient way to remove multiple keys.
	MDelete(keys ...K)
}
