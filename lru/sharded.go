package lru

import (
	"github.com/jxskiss/gopkg/rthash"
	"reflect"
	"time"
)

var shardingHash = rthash.New()

// NewMultiCache is renamed to NewShardedCache, please use the new name
// instead of this. It will be removed in future.
//
// Deprecated.
func NewMultiCache(buckets, bucketCapacity int) *MultiCache {
	return NewShardedCache(buckets, bucketCapacity)
}

// MultiCache is renamed to ShardedCache, please use the new name instead
// of this. It will be removed in future.
//
// Deprecated.
type MultiCache = ShardedCache

// NewShardedCache returns a hash-sharded lru cache instance which is suitable
// to use for heavy lock contention use-case. It keeps same interface with
// the lru cache instance returned by NewCache function.
// Generally NewCache should be used instead of this unless you are sure that
// you are facing the lock contention problem.
func NewShardedCache(buckets, bucketCapacity int) *ShardedCache {
	buckets = nextPowerOfTwo(buckets)
	mask := uintptr(buckets - 1)
	mc := &ShardedCache{
		buckets: uintptr(buckets),
		mask:    mask,
		cache:   make([]*Cache, buckets),
	}
	for i := 0; i < buckets; i++ {
		mc.cache[i] = NewCache(bucketCapacity)
	}
	return mc
}

func nextPowerOfTwo(x int) int {
	if x == 0 {
		return 1
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16

	return x + 1
}

// ShardedCache is a hash-sharded version of Cache, it minimizes lock
// contention for heavy read workload. Generally Cache should be used
// instead of this unless you are sure that you are facing the lock
// contention problem.
//
// It implements Interface in this package, see Interface for detailed
// api documents.
type ShardedCache struct {
	buckets uintptr
	mask    uintptr
	cache   []*Cache
}

func (c *ShardedCache) Len() (n int) {
	for _, c := range c.cache {
		n += c.Len()
	}
	return
}

func (c *ShardedCache) Has(key interface{}) (exists, expired bool) {
	h := shardingHash.Hash(key)
	return c.cache[h&c.mask].Has(key)
}

func (c *ShardedCache) Get(key interface{}) (v interface{}, exists, expired bool) {
	h := shardingHash.Hash(key)
	return c.cache[h&c.mask].Get(key)
}

func (c *ShardedCache) GetWithTTL(key interface{}) (v interface{}, exists bool, ttl *time.Duration) {
	h := shardingHash.Hash(key)
	return c.cache[h&c.mask].GetWithTTL(key)
}

func (c *ShardedCache) GetQuiet(key interface{}) (v interface{}, exists, expired bool) {
	h := shardingHash.Hash(key)
	return c.cache[h&c.mask].GetQuiet(key)
}

func (c *ShardedCache) GetNotStale(key interface{}) (v interface{}, exists bool) {
	h := shardingHash.Hash(key)
	return c.cache[h&c.mask].GetNotStale(key)
}

func (c *ShardedCache) MGet(keys ...interface{}) map[interface{}]interface{} {
	return c.mget(false, keys...)
}

func (c *ShardedCache) MGetNotStale(keys ...interface{}) map[interface{}]interface{} {
	return c.mget(true, keys...)
}

func (c *ShardedCache) mget(notStale bool, keys ...interface{}) map[interface{}]interface{} {
	grpKeys := c.groupKeys(keys)
	nowNano := time.Now().UnixNano()

	var res map[interface{}]interface{}
	for idx, keys := range grpKeys {
		grp := c.cache[idx].mget(notStale, nowNano, keys...)
		if res == nil {
			res = grp
		} else {
			for k, v := range grp {
				res[k] = v
			}
		}
	}
	return res
}

func (c *ShardedCache) MGetInt(keys ...int) map[int]interface{} {
	return c.mgetInt(false, keys...)
}

func (c *ShardedCache) MGetIntNotStale(keys ...int) map[int]interface{} {
	return c.mgetInt(true, keys...)
}

func (c *ShardedCache) mgetInt(notStale bool, keys ...int) map[int]interface{} {
	grpKeys := c.groupIntKeys(keys)
	nowNano := time.Now().UnixNano()

	var res map[int]interface{}
	for idx, keys := range grpKeys {
		grp := c.cache[idx].mgetInt(notStale, nowNano, keys...)
		if res == nil {
			res = grp
		} else {
			for k, v := range grp {
				res[k] = v
			}
		}
	}
	return res
}

func (c *ShardedCache) MGetInt64(keys ...int64) map[int64]interface{} {
	return c.mgetInt64(false, keys...)
}

func (c *ShardedCache) MGetInt64NotStale(keys ...int64) map[int64]interface{} {
	return c.mgetInt64(true, keys...)
}

func (c *ShardedCache) mgetInt64(notStale bool, keys ...int64) map[int64]interface{} {
	grpKeys := c.groupInt64Keys(keys)
	nowNano := time.Now().UnixNano()

	var res map[int64]interface{}
	for idx, keys := range grpKeys {
		grp := c.cache[idx].mgetInt64(notStale, nowNano, keys...)
		if res == nil {
			res = grp
		} else {
			for k, v := range grp {
				res[k] = v
			}
		}
	}
	return res
}

func (c *ShardedCache) MGetUint64(keys ...uint64) map[uint64]interface{} {
	return c.mgetUint64(false, keys...)
}

func (c *ShardedCache) MGetUint64NotStale(keys ...uint64) map[uint64]interface{} {
	return c.mgetUint64(true, keys...)
}

func (c *ShardedCache) mgetUint64(notStale bool, keys ...uint64) map[uint64]interface{} {
	grpKeys := c.groupUint64Keys(keys)
	nowNano := time.Now().UnixNano()

	var res map[uint64]interface{}
	for idx, keys := range grpKeys {
		grp := c.cache[idx].mgetUint64(notStale, nowNano, keys...)
		if res == nil {
			res = grp
		} else {
			for k, v := range grp {
				res[k] = v
			}
		}
	}
	return res
}

func (c *ShardedCache) MGetString(keys ...string) map[string]interface{} {
	return c.mgetString(false, keys...)
}

func (c *ShardedCache) MGetStringNotStale(keys ...string) map[string]interface{} {
	return c.mgetString(true, keys...)
}

func (c *ShardedCache) mgetString(notStale bool, keys ...string) map[string]interface{} {
	grpKeys := c.groupStringKeys(keys)
	nowNano := time.Now().UnixNano()

	var res map[string]interface{}
	for idx, keys := range grpKeys {
		grp := c.cache[idx].mgetString(notStale, nowNano, keys...)
		if res == nil {
			res = grp
		} else {
			for k, v := range grp {
				res[k] = v
			}
		}
	}
	return res
}

func (c *ShardedCache) Set(key, value interface{}, ttl time.Duration) {
	h := shardingHash.Hash(key)
	c.cache[h&c.mask].Set(key, value, ttl)
}

func (c *ShardedCache) MSet(kvmap interface{}, ttl time.Duration) {
	m := reflect.ValueOf(kvmap)
	keys := m.MapKeys()

	for _, key := range keys {
		value := m.MapIndex(key)
		c.Set(key.Interface(), value.Interface(), ttl)
	}
}

func (c *ShardedCache) Del(key interface{}) {
	h := shardingHash.Hash(key)
	c.cache[h&c.mask].Del(key)
}

func (c *ShardedCache) MDel(keys ...interface{}) {
	grpKeys := c.groupKeys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDel(keys...)
	}
}

func (c *ShardedCache) MDelInt(keys ...int) {
	grpKeys := c.groupIntKeys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelInt(keys...)
	}
}

func (c *ShardedCache) MDelInt64(keys ...int64) {
	grpKeys := c.groupInt64Keys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelInt64(keys...)
	}
}

func (c *ShardedCache) MDelUint64(keys ...uint64) {
	grpKeys := c.groupUint64Keys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelUint64(keys...)
	}
}

func (c *ShardedCache) MDelString(keys ...string) {
	grpKeys := c.groupStringKeys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelString(keys...)
	}
}

func (c *ShardedCache) groupKeys(keys []interface{}) map[uintptr][]interface{} {
	grpKeys := make(map[uintptr][]interface{})
	for _, key := range keys {
		idx := shardingHash.Interface(key) & c.mask
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}

func (c *ShardedCache) groupIntKeys(keys []int) map[uintptr][]int {
	grpKeys := make(map[uintptr][]int)
	for _, key := range keys {
		idx := shardingHash.Int(key) & c.mask
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}

func (c *ShardedCache) groupInt64Keys(keys []int64) map[uintptr][]int64 {
	grpKeys := make(map[uintptr][]int64)
	for _, key := range keys {
		idx := shardingHash.Int64(key) & c.mask
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}

func (c *ShardedCache) groupUint64Keys(keys []uint64) map[uintptr][]uint64 {
	grpKeys := make(map[uintptr][]uint64)
	for _, key := range keys {
		idx := shardingHash.Uint64(key) & c.mask
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}

func (c *ShardedCache) groupStringKeys(keys []string) map[uintptr][]string {
	grpKeys := make(map[uintptr][]string)
	for _, key := range keys {
		idx := shardingHash.String(key) & c.mask
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}
