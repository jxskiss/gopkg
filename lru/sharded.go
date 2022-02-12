package lru

import (
	"reflect"
	"time"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal"
	"github.com/jxskiss/gopkg/v2/rthash"
)

var shardingHash = rthash.New()

func buildHashFunction[K comparable]() func(x K) uintptr {
	var _k K
	typ := reflect.TypeOf((interface{})(_k))
	if typ == nil {
		return func(x K) uintptr {
			return shardingHash.Interface(x)
		}
	}
	switch typ.Kind() {
	case reflect.String:
		return func(x K) uintptr {
			return shardingHash.String(*(*string)(unsafe.Pointer(&x)))
		}
	case reflect.Int8:
		return func(x K) uintptr {
			return shardingHash.Int8(*(*int8)(unsafe.Pointer(&x)))
		}
	case reflect.Uint8:
		return func(x K) uintptr {
			return shardingHash.Uint8(*(*uint8)(unsafe.Pointer(&x)))
		}
	case reflect.Int16:
		return func(x K) uintptr {
			return shardingHash.Int16(*(*int16)(unsafe.Pointer(&x)))
		}
	case reflect.Uint16:
		return func(x K) uintptr {
			return shardingHash.Uint16(*(*uint16)(unsafe.Pointer(&x)))
		}
	case reflect.Int32:
		return func(x K) uintptr {
			return shardingHash.Int32(*(*int32)(unsafe.Pointer(&x)))
		}
	case reflect.Uint32:
		return func(x K) uintptr {
			return shardingHash.Uint32(*(*uint32)(unsafe.Pointer(&x)))
		}
	case reflect.Int64:
		return func(x K) uintptr {
			return shardingHash.Int64(*(*int64)(unsafe.Pointer(&x)))
		}
	case reflect.Uint64:
		return func(x K) uintptr {
			return shardingHash.Uint64(*(*uint64)(unsafe.Pointer(&x)))
		}
	case reflect.Int:
		return func(x K) uintptr {
			return shardingHash.Int(*(*int)(unsafe.Pointer(&x)))
		}
	case reflect.Uint:
		return func(x K) uintptr {
			return shardingHash.Uint(*(*uint)(unsafe.Pointer(&x)))
		}
	case reflect.Uintptr:
		return func(x K) uintptr {
			return shardingHash.Uintptr(*(*uintptr)(unsafe.Pointer(&x)))
		}
	case reflect.Float32:
		return func(x K) uintptr {
			return shardingHash.Float32(*(*float32)(unsafe.Pointer(&x)))
		}
	case reflect.Float64:
		return func(x K) uintptr {
			return shardingHash.Float64(*(*float64)(unsafe.Pointer(&x)))
		}
	case reflect.Complex64:
		return func(x K) uintptr {
			return shardingHash.Complex64(*(*complex64)(unsafe.Pointer(&x)))
		}
	case reflect.Complex128:
		return func(x K) uintptr {
			return shardingHash.Complex128(*(*complex128)(unsafe.Pointer(&x)))
		}
	}
	return func(x K) uintptr {
		return shardingHash.Interface(x)
	}
}

// NewShardedCache returns a hash-sharded lru cache instance which is suitable
// to use for heavy lock contention use-case. It keeps same interface with
// the lru cache instance returned by NewCache function.
// Generally NewCache should be used instead of this unless you are sure that
// you are facing the lock contention problem.
func NewShardedCache[K comparable, V any](buckets, bucketCapacity int) *ShardedCache[K, V] {
	buckets = int(internal.NextPowerOfTwo(uint(buckets)))
	mask := uintptr(buckets - 1)
	mc := &ShardedCache[K, V]{
		buckets: uintptr(buckets),
		mask:    mask,
		cache:   make([]*Cache[K, V], buckets),
	}
	for i := 0; i < buckets; i++ {
		mc.cache[i] = NewCache[K, V](bucketCapacity)
	}
	mc.hashFunc = buildHashFunction[K]()
	return mc
}

// ShardedCache is a hash-sharded version of Cache, it minimizes lock
// contention for heavy read workload. Generally Cache should be used
// instead of this unless you are sure that you are facing the lock
// contention problem.
//
// It implements Interface in this package, see Interface for detailed
// api documents.
type ShardedCache[K comparable, V any] struct {
	buckets uintptr
	mask    uintptr
	cache   []*Cache[K, V]

	hashFunc func(K) uintptr
}

func (c *ShardedCache[K, V]) Len() (n int) {
	for _, c := range c.cache {
		n += c.Len()
	}
	return
}

func (c *ShardedCache[K, V]) Has(key K) (exists, expired bool) {
	h := c.hashFunc(key)
	return c.cache[h&c.mask].Has(key)
}

func (c *ShardedCache[K, V]) Get(key K) (v V, exists, expired bool) {
	h := c.hashFunc(key)
	return c.cache[h&c.mask].Get(key)
}

func (c *ShardedCache[K, V]) GetWithTTL(key K) (v V, exists bool, ttl *time.Duration) {
	h := c.hashFunc(key)
	return c.cache[h&c.mask].GetWithTTL(key)
}

func (c *ShardedCache[K, V]) GetQuiet(key K) (v V, exists, expired bool) {
	h := c.hashFunc(key)
	return c.cache[h&c.mask].GetQuiet(key)
}

func (c *ShardedCache[K, V]) GetNotStale(key K) (v V, exists bool) {
	h := c.hashFunc(key)
	return c.cache[h&c.mask].GetNotStale(key)
}

func (c *ShardedCache[K, V]) MGet(keys ...K) map[K]V {
	return c.mget(false, keys...)
}

func (c *ShardedCache[K, V]) MGetNotStale(keys ...K) map[K]V {
	return c.mget(true, keys...)
}

func (c *ShardedCache[K, V]) mget(notStale bool, keys ...K) map[K]V {
	grpKeys := c.groupKeys(keys)
	nowNano := time.Now().UnixNano()

	var res map[K]V
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

func (c *ShardedCache[K, V]) Set(key K, value V, ttl time.Duration) {
	h := c.hashFunc(key)
	c.cache[h&c.mask].Set(key, value, ttl)
}

func (c *ShardedCache[K, V]) MSet(kvmap map[K]V, ttl time.Duration) {
	for key, val := range kvmap {
		c.Set(key, val, ttl)
	}
}

func (c *ShardedCache[K, V]) Delete(key K) {
	h := c.hashFunc(key)
	c.cache[h&c.mask].Delete(key)
}

func (c *ShardedCache[K, V]) MDelete(keys ...K) {
	grpKeys := c.groupKeys(keys)

	for idx, keys := range grpKeys {
		c.cache[idx].MDelete(keys...)
	}
}

func (c *ShardedCache[K, V]) groupKeys(keys []K) map[uintptr][]K {
	grpKeys := make(map[uintptr][]K)
	for _, key := range keys {
		idx := c.hashFunc(key) & c.mask
		grpKeys[idx] = append(grpKeys[idx], key)
	}
	return grpKeys
}
