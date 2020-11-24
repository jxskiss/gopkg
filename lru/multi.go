package lru

import (
	"github.com/jxskiss/gopkg/fasthash"
	"reflect"
	"time"
)

// NewMultiCache returns a hash-shared lru cache instance which is suitable
// to use for heavy lock contention use-case. It keeps same interface with
// the lru cache instance returned by NewCache function.
// Generally NewCache should be used instead of this unless you are sure that
// you are facing the lock contention problem.
func NewMultiCache(buckets, bucketCapacity int) *MultiCache {
	mc := &MultiCache{
		buckets: uintptr(buckets),
		cache:   make([]*Cache, buckets),
	}
	for i := 0; i < buckets; i++ {
		mc.cache[i] = NewCache(bucketCapacity)
	}
	return mc
}

type MultiCache struct {
	buckets uintptr
	cache   []*Cache
}

func (c *MultiCache) Len() (n int) {
	for _, c := range c.cache {
		n += c.Len()
	}
	return
}

func (c *MultiCache) Has(key interface{}) (exists, expired bool) {
	h := fasthash.Hash(key)
	return c.cache[h%c.buckets].Has(key)
}

func (c *MultiCache) Get(key interface{}) (v interface{}, exists, expired bool) {
	h := fasthash.Hash(key)
	return c.cache[h%c.buckets].Get(key)
}

func (c *MultiCache) GetQuiet(key interface{}) (v interface{}, exists, expired bool) {
	h := fasthash.Hash(key)
	return c.cache[h%c.buckets].GetQuiet(key)
}

func (c *MultiCache) GetNotStale(key interface{}) (v interface{}, exists bool) {
	h := fasthash.Hash(key)
	return c.cache[h%c.buckets].GetNotStale(key)
}

func (c *MultiCache) MGet(keys ...interface{}) map[interface{}]interface{} {
	return c.mget(false, keys...)
}

func (c *MultiCache) MGetNotStale(keys ...interface{}) map[interface{}]interface{} {
	return c.mget(true, keys...)
}

func (c *MultiCache) mget(notStale bool, keys ...interface{}) map[interface{}]interface{} {
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

func (c *MultiCache) MGetInt(keys ...int) map[int]interface{} {
	return c.mgetInt(false, keys...)
}

func (c *MultiCache) MGetIntNotStale(keys ...int) map[int]interface{} {
	return c.mgetInt(true, keys...)
}

func (c *MultiCache) mgetInt(notStale bool, keys ...int) map[int]interface{} {
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

func (c *MultiCache) MGetInt64(keys ...int64) map[int64]interface{} {
	return c.mgetInt64(false, keys...)
}

func (c *MultiCache) MGetInt64NotStale(keys ...int64) map[int64]interface{} {
	return c.mgetInt64(true, keys...)
}

func (c *MultiCache) mgetInt64(notStale bool, keys ...int64) map[int64]interface{} {
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

func (c *MultiCache) MGetUint64(keys ...uint64) map[uint64]interface{} {
	return c.mgetUint64(false, keys...)
}

func (c *MultiCache) MGetUint64NotStale(keys ...uint64) map[uint64]interface{} {
	return c.mgetUint64(true, keys...)
}

func (c *MultiCache) mgetUint64(notStale bool, keys ...uint64) map[uint64]interface{} {
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

func (c *MultiCache) MGetString(keys ...string) map[string]interface{} {
	return c.mgetString(false, keys...)
}

func (c *MultiCache) MGetStringNotStale(keys ...string) map[string]interface{} {
	return c.mgetString(true, keys...)
}

func (c *MultiCache) mgetString(notStale bool, keys ...string) map[string]interface{} {
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

func (c *MultiCache) Set(key, value interface{}, ttl time.Duration) {
	h := fasthash.Hash(key)
	c.cache[h%c.buckets].Set(key, value, ttl)
}

func (c *MultiCache) MSet(kvmap interface{}, ttl time.Duration) {
	m := reflect.ValueOf(kvmap)
	keys := m.MapKeys()

	for _, key := range keys {
		value := m.MapIndex(key)
		c.Set(key.Interface(), value.Interface(), ttl)
	}
}

func (c *MultiCache) Del(key interface{}) {
	h := fasthash.Hash(key)
	c.cache[h%c.buckets].Del(key)
}

func (c *MultiCache) MDel(keys ...interface{}) {
	grpKeys := c.groupKeys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDel(keys...)
	}
}

func (c *MultiCache) MDelInt(keys ...int) {
	grpKeys := c.groupIntKeys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelInt(keys...)
	}
}

func (c *MultiCache) MDelInt64(keys ...int64) {
	grpKeys := c.groupInt64Keys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelInt64(keys...)
	}
}

func (c *MultiCache) MDelUint64(keys ...uint64) {
	grpKeys := c.groupUint64Keys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelUint64(keys...)
	}
}

func (c *MultiCache) MDelString(keys ...string) {
	grpKeys := c.groupStringKeys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelString(keys...)
	}
}

func (c *MultiCache) groupKeys(keys []interface{}) map[uintptr][]interface{} {
	grpKeys := make(map[uintptr][]interface{})
	for _, key := range keys {
		idx := fasthash.Interface(key) % c.buckets
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}

func (c *MultiCache) groupIntKeys(keys []int) map[uintptr][]int {
	grpKeys := make(map[uintptr][]int)
	for _, key := range keys {
		idx := fasthash.Int(key) % c.buckets
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}

func (c *MultiCache) groupInt64Keys(keys []int64) map[uintptr][]int64 {
	grpKeys := make(map[uintptr][]int64)
	for _, key := range keys {
		idx := fasthash.Int64(key) % c.buckets
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}

func (c *MultiCache) groupUint64Keys(keys []uint64) map[uintptr][]uint64 {
	grpKeys := make(map[uintptr][]uint64)
	for _, key := range keys {
		idx := fasthash.Uint64(key) % c.buckets
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}

func (c *MultiCache) groupStringKeys(keys []string) map[uintptr][]string {
	grpKeys := make(map[uintptr][]string)
	for _, key := range keys {
		idx := fasthash.String(key) % c.buckets
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}
