package lru

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const maxCapacity = 1<<32 - 1

// NewCache returns a lru cache instance with given capacity, the underlying
// memory will be immediately allocated. For best performance, the memory
// will be reused and won't be freed for the lifetime of the cache.
//
// Param capacity must be smaller than 2^32, else it will panic.
func NewCache(capacity int) *Cache {
	if capacity > maxCapacity {
		panic("invalid too large capacity")
	}
	list := newList(capacity)
	c := &Cache{
		list: list,
		m:    make(map[interface{}]uint32, capacity),
		buf:  unsafe.Pointer(newWalBuf()),
	}
	return c
}

// Cache is a in-memory cache using LRU algorithm.
//
// It implements Interface in this package, see Interface for detailed
// api documents.
type Cache struct {
	mu   sync.RWMutex
	list *list
	m    map[interface{}]uint32

	buf unsafe.Pointer // *walbuf
}

func (c *Cache) Len() (n int) {
	c.mu.RLock()
	n = len(c.m)
	c.mu.RUnlock()
	return
}

func (c *Cache) Has(key interface{}) (exists, expired bool) {
	c.mu.RLock()
	_, elem, exists := c.get(key)
	if exists {
		expired = elem.expires > 0 && elem.expires < time.Now().UnixNano()
	}
	c.mu.RUnlock()
	return
}

func (c *Cache) Get(key interface{}) (v interface{}, exists, expired bool) {
	c.mu.RLock()
	idx, elem, exists := c.get(key)
	if exists {
		v = elem.value
		expired = elem.expires > 0 && elem.expires < time.Now().UnixNano()
		c.promote(idx)
	}
	c.mu.RUnlock()
	return
}

func (c *Cache) GetWithTTL(key interface{}) (v interface{}, exists bool, ttl *time.Duration) {
	c.mu.RLock()
	idx, elem, exists := c.get(key)
	if exists {
		v = elem.value
		if elem.expires > 0 {
			x := time.Duration(elem.expires - time.Now().UnixNano())
			ttl = &x
		}
		c.promote(idx)
	}
	c.mu.RUnlock()
	return
}

func (c *Cache) GetQuiet(key interface{}) (v interface{}, exists, expired bool) {
	c.mu.RLock()
	_, elem, exists := c.get(key)
	if exists {
		v = elem.value
		expired = elem.expires > 0 && elem.expires < time.Now().UnixNano()
	}
	c.mu.RUnlock()
	return
}

func (c *Cache) GetNotStale(key interface{}) (v interface{}, exists bool) {
	c.mu.RLock()
	idx, elem, exists := c.get(key)
	if exists {
		expired := elem.expires > 0 && elem.expires < time.Now().UnixNano()
		if !expired {
			v = elem.value
			c.promote(idx)
		} else {
			exists = false
		}
	}
	c.mu.RUnlock()
	return
}

func (c *Cache) get(key interface{}) (idx uint32, elem *element, exists bool) {
	idx, exists = c.m[key]
	if exists {
		elem = c.list.get(idx)
	}
	return
}

func (c *Cache) MGet(keys ...interface{}) map[interface{}]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mget(false, nowNano, keys...)
}

func (c *Cache) MGetNotStale(keys ...interface{}) map[interface{}]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mget(true, nowNano, keys...)
}

func (c *Cache) mget(notStale bool, nowNano int64, keys ...interface{}) map[interface{}]interface{} {
	res := make(map[interface{}]interface{}, len(keys))
	c.mu.RLock()
	for _, key := range keys {
		idx, elem, exists := c.get(key)
		if exists {
			if notStale {
				expired := elem.expires > 0 && elem.expires < nowNano
				if expired {
					continue
				}
			}
			res[key] = elem.value
			c.promote(idx)
		}
	}
	c.mu.RUnlock()
	return res
}

func (c *Cache) MGetInt(keys ...int) map[int]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mgetInt(false, nowNano, keys...)
}

func (c *Cache) MGetIntNotStale(keys ...int) map[int]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mgetInt(true, nowNano, keys...)
}

func (c *Cache) mgetInt(notStale bool, nowNano int64, keys ...int) map[int]interface{} {
	res := make(map[int]interface{}, len(keys))
	c.mu.RLock()
	for _, key := range keys {
		idx, elem, exists := c.get(key)
		if exists {
			if notStale {
				expired := elem.expires > 0 && elem.expires < nowNano
				if expired {
					continue
				}
			}
			res[key] = elem.value
			c.promote(idx)
		}
	}
	c.mu.RUnlock()
	return res
}

func (c *Cache) MGetInt64(keys ...int64) map[int64]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mgetInt64(false, nowNano, keys...)
}

func (c *Cache) MGetInt64NotStale(keys ...int64) map[int64]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mgetInt64(true, nowNano, keys...)
}

func (c *Cache) mgetInt64(notStale bool, nowNano int64, keys ...int64) map[int64]interface{} {
	res := make(map[int64]interface{}, len(keys))
	c.mu.RLock()
	for _, key := range keys {
		idx, elem, exists := c.get(key)
		if exists {
			if notStale {
				expired := elem.expires > 0 && elem.expires < nowNano
				if expired {
					continue
				}
			}
			res[key] = elem.value
			c.promote(idx)
		}
	}
	c.mu.RUnlock()
	return res
}

func (c *Cache) MGetUint64(keys ...uint64) map[uint64]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mgetUint64(false, nowNano, keys...)
}

func (c *Cache) MGetUint64NotStale(keys ...uint64) map[uint64]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mgetUint64(true, nowNano, keys...)
}

func (c *Cache) mgetUint64(notStale bool, nowNano int64, keys ...uint64) map[uint64]interface{} {
	res := make(map[uint64]interface{}, len(keys))
	c.mu.RLock()
	for _, key := range keys {
		idx, elem, exists := c.get(key)
		if exists {
			if notStale {
				expired := elem.expires > 0 && elem.expires < nowNano
				if expired {
					continue
				}
			}
			res[key] = elem.value
			c.promote(idx)
		}
	}
	c.mu.RUnlock()
	return res
}

func (c *Cache) MGetString(keys ...string) map[string]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mgetString(false, nowNano, keys...)
}

func (c *Cache) MGetStringNotStale(keys ...string) map[string]interface{} {
	nowNano := time.Now().UnixNano()
	return c.mgetString(true, nowNano, keys...)
}

func (c *Cache) mgetString(notStale bool, nowNano int64, keys ...string) map[string]interface{} {
	res := make(map[string]interface{}, len(keys))
	c.mu.RLock()
	for _, key := range keys {
		idx, elem, exists := c.get(key)
		if exists {
			if notStale {
				expired := elem.expires > 0 && elem.expires < nowNano
				if expired {
					continue
				}
			}
			res[key] = elem.value
			c.promote(idx)
		}
	}
	c.mu.RUnlock()
	return res
}

func (c *Cache) promote(idx uint32) {
	buf := (*walbuf)(atomic.LoadPointer(&c.buf))
	i := atomic.AddInt32(&buf.p, 1)
	if i <= walBufSize {
		buf.b[i-1] = idx
		return
	}

	// buffer is full, swap buffer
	oldbuf := buf

	// create new buffer, and reserve the first element to use for
	// this promotion request
	newbuf := newWalBuf()
	newbuf.p = 1
	for {
		swapped := atomic.CompareAndSwapPointer(&c.buf, unsafe.Pointer(oldbuf), unsafe.Pointer(newbuf))
		if swapped {
			newbuf.b[0] = idx
			break
		}

		// try again
		oldbuf = (*walbuf)(atomic.LoadPointer(&c.buf))
		i = atomic.AddInt32(&oldbuf.p, 1)
		if i <= walBufSize {
			oldbuf.b[i-1] = idx
			walbufpool.Put(newbuf)
			return
		}
	}

	// the oldbuf has been swapped, we take responsibility to flush it
	go func(c *Cache, buf *walbuf) {
		c.mu.Lock()
		c.flushBuf(buf)
		c.mu.Unlock()
		walbufpool.Put(buf)
	}(c, oldbuf)
}

func (c *Cache) Set(key, value interface{}, ttl time.Duration) {
	var expires int64
	if ttl > 0 {
		expires = time.Now().UnixNano() + int64(ttl)
	}
	c.mu.Lock()
	c.checkAndFlushBuf()
	c.set(key, value, expires)
	c.mu.Unlock()
}

func (c *Cache) MSet(kvmap interface{}, ttl time.Duration) {
	var expires int64
	if ttl > 0 {
		expires = time.Now().UnixNano() + int64(ttl)
	}
	m := reflect.ValueOf(kvmap)
	keys := m.MapKeys()

	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		value := m.MapIndex(key)
		c.set(key.Interface(), value.Interface(), expires)
	}
	c.mu.Unlock()
}

func (c *Cache) set(k, v interface{}, expires int64) {
	idx, exists := c.m[k]
	if exists {
		e := c.list.get(idx)
		e.value = v
		e.expires = expires
		c.list.MoveToFront(e)
	} else {
		e := c.list.Back()
		if e.key != nil {
			delete(c.m, e.key)
		}
		e.key = k
		e.value = v
		e.expires = expires
		c.m[k] = e.index
		c.list.MoveToFront(e)
	}
}

func (c *Cache) Del(key interface{}) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	c.del(key)
	c.mu.Unlock()
}

func (c *Cache) MDel(keys ...interface{}) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *Cache) MDelInt(keys ...int) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *Cache) MDelInt64(keys ...int64) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *Cache) MDelUint64(keys ...uint64) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *Cache) MDelString(keys ...string) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *Cache) del(key interface{}) {
	idx, exists := c.m[key]
	if exists {
		delete(c.m, key)
		elem := c.list.get(idx)
		elem.key = nil
		elem.value = nil
		c.list.MoveToBack(elem)
	}
}

func (c *Cache) checkAndFlushBuf() {
	buf := (*walbuf)(c.buf)
	if buf.p > 0 {
		c.flushBuf(buf)
	}
}

func (c *Cache) flushBuf(buf *walbuf) {
	if buf.p == 0 {
		return
	}

	// remove duplicate elements
	b := buf.deduplicate()

	// promote elements by their access order
	for _, idx := range b {
		elem := c.list.get(idx)
		c.list.MoveToFront(elem)
	}

	buf.p = 0
}
