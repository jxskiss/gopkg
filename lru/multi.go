package lru

import (
	"github.com/jxskiss/gopkg/rthash"
	"reflect"
	"time"
)

func NewMultiCache(buckets, bucketCapacity int) *multiCache {
	mc := &multiCache{
		buckets: uintptr(buckets),
		cache:   make([]*cache, buckets),
	}
	for i := 0; i < buckets; i++ {
		mc.cache[i] = NewCache(bucketCapacity)
	}
	return mc
}

type multiCache struct {
	buckets uintptr
	cache   []*cache
}

func (c *multiCache) Len() (n int) {
	for _, c := range c.cache {
		n += c.Len()
	}
	return
}

func (c *multiCache) Get(key interface{}) (v interface{}, exists, expired bool) {
	h := rthash.Hash(key)
	return c.cache[h%c.buckets].Get(key)
}

func (c *multiCache) GetQuiet(key interface{}) (v interface{}, exists, expired bool) {
	h := rthash.Hash(key)
	return c.cache[h%c.buckets].GetQuiet(key)
}

func (c *multiCache) GetNotStale(key interface{}) (v interface{}, exists bool) {
	h := rthash.Hash(key)
	return c.cache[h%c.buckets].GetNotStale(key)
}

func (c *multiCache) MGetInt64(keys ...int64) map[int64]interface{} {
	grpKeys := c.groupInt64Keys(keys)

	var res map[int64]interface{}
	for h, keys := range grpKeys {
		grp := c.cache[h%c.buckets].MGetInt64(keys...)
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

func (c *multiCache) MGetString(keys ...string) map[string]interface{} {
	grpKeys := c.groupStringKeys(keys)

	var res map[string]interface{}
	for h, keys := range grpKeys {
		grp := c.cache[h%c.buckets].MGetString(keys...)
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

func (c *multiCache) Set(key, value interface{}, ttl time.Duration) {
	h := rthash.Hash(key)
	c.cache[h%c.buckets].Set(key, value, ttl)
}

func (c *multiCache) MSet(kvmap interface{}, ttl time.Duration) {
	m := reflect.ValueOf(kvmap)
	keys := m.MapKeys()

	for _, key := range keys {
		value := m.MapIndex(key)
		c.Set(key.Interface(), value.Interface(), ttl)
	}
}

func (c *multiCache) Del(key interface{}) {
	h := rthash.Hash(key)
	c.cache[h%c.buckets].Del(key)
}

func (c *multiCache) MDelInt64(keys ...int64) {
	grpKeys := c.groupInt64Keys(keys)

	for h, keys := range grpKeys {
		c.cache[h%c.buckets].MDelInt64(keys...)
	}
}

func (c *multiCache) MDelString(keys ...string) {
	grpKeys := c.groupStringKeys(keys)

	for h, keys := range grpKeys {
		c.cache[h%c.buckets].MDelString(keys...)
	}
}

func (c *multiCache) groupInt64Keys(keys []int64) map[uintptr][]int64 {
	grpKeys := make(map[uintptr][]int64)
	for _, key := range keys {
		h := rthash.Int64(key)
		grpKeys[h] = append(grpKeys[h], key)
	}
	return grpKeys
}

func (c *multiCache) groupStringKeys(keys []string) map[uintptr][]string {
	grpKeys := make(map[uintptr][]string)
	for _, key := range keys {
		h := rthash.String(key)
		grpKeys[h] = append(grpKeys[h], key)
	}
	return grpKeys
}
