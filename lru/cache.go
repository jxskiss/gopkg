package lru

import (
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
func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity > maxCapacity {
		panic("invalid too large capacity")
	}
	list := newList(capacity)
	c := &Cache[K, V]{
		list: list,
		m:    make(map[K]uint32, capacity),
		buf:  unsafe.Pointer(newWalBuf()),
	}
	return c
}

// Cache is an in-memory cache using LRU algorithm.
//
// It implements Interface in this package, see Interface for detailed
// api documents.
type Cache[K comparable, V any] struct {
	mu   sync.RWMutex
	list *list
	m    map[K]uint32

	buf unsafe.Pointer // *walbuf
}

func (c *Cache[K, V]) Len() (n int) {
	c.mu.RLock()
	n = len(c.m)
	c.mu.RUnlock()
	return
}

func (c *Cache[K, V]) Has(key K) (exists, expired bool) {
	c.mu.RLock()
	_, elem, exists := c.get(key)
	if exists {
		expired = elem.expires > 0 && elem.expires < time.Now().UnixNano()
	}
	c.mu.RUnlock()
	return
}

func (c *Cache[K, V]) Get(key K) (v V, exists, expired bool) {
	c.mu.RLock()
	idx, elem, exists := c.get(key)
	if exists {
		v = elem.value.(V)
		expired = elem.expires > 0 && elem.expires < time.Now().UnixNano()
		c.promote(idx)
	}
	c.mu.RUnlock()
	return
}

func (c *Cache[K, V]) GetWithTTL(key K) (v V, exists bool, ttl *time.Duration) {
	c.mu.RLock()
	idx, elem, exists := c.get(key)
	if exists {
		v = elem.value.(V)
		if elem.expires > 0 {
			x := time.Duration(elem.expires - time.Now().UnixNano())
			ttl = &x
		}
		c.promote(idx)
	}
	c.mu.RUnlock()
	return
}

func (c *Cache[K, V]) GetQuiet(key K) (v V, exists, expired bool) {
	c.mu.RLock()
	_, elem, exists := c.get(key)
	if exists {
		v = elem.value.(V)
		expired = elem.expires > 0 && elem.expires < time.Now().UnixNano()
	}
	c.mu.RUnlock()
	return
}

func (c *Cache[K, V]) GetNotStale(key K) (v V, exists bool) {
	c.mu.RLock()
	idx, elem, exists := c.get(key)
	if exists {
		expired := elem.expires > 0 && elem.expires < time.Now().UnixNano()
		if !expired {
			v = elem.value.(V)
			c.promote(idx)
		} else {
			exists = false
		}
	}
	c.mu.RUnlock()
	return
}

func (c *Cache[K, V]) get(key K) (idx uint32, elem *element, exists bool) {
	idx, exists = c.m[key]
	if exists {
		elem = c.list.get(idx)
	}
	return
}

func (c *Cache[K, V]) MGet(keys ...K) map[K]V {
	nowNano := time.Now().UnixNano()
	return c.mget(false, nowNano, keys...)
}

func (c *Cache[K, V]) MGetNotStale(keys ...K) map[K]V {
	nowNano := time.Now().UnixNano()
	return c.mget(true, nowNano, keys...)
}

func (c *Cache[K, V]) mget(notStale bool, nowNano int64, keys ...K) map[K]V {
	res := make(map[K]V, len(keys))

	// Split into batches to let the LRU cache to have chance to be updated
	// if length of keys is much larger than walBufSize.
	total := len(keys)
	batch := walBufSize
	for i, j := 0, batch; i < total; i, j = i+batch, j+batch {
		if j > total {
			j = total
		}

		c.mu.RLock()
		for _, key := range keys[i:j] {
			idx, elem, exists := c.get(key)
			if exists {
				if notStale {
					expired := elem.expires > 0 && elem.expires < nowNano
					if expired {
						continue
					}
				}
				res[key] = elem.value.(V)
				c.promote(idx)
			}
		}
		c.mu.RUnlock()
	}
	return res
}

func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) {
	var expires int64
	if ttl > 0 {
		expires = time.Now().UnixNano() + int64(ttl)
	}
	c.mu.Lock()
	c.checkAndFlushBuf()
	c.set(key, value, expires)
	c.mu.Unlock()
}

func (c *Cache[K, V]) MSet(kvmap map[K]V, ttl time.Duration) {
	var expires int64
	if ttl > 0 {
		expires = time.Now().UnixNano() + int64(ttl)
	}

	c.mu.Lock()
	c.checkAndFlushBuf()
	for key, val := range kvmap {
		c.set(key, val, expires)
	}
	c.mu.Unlock()
}

func (c *Cache[K, V]) set(k K, v V, expires int64) {
	idx, exists := c.m[k]
	if exists {
		e := c.list.get(idx)
		e.value = v
		e.expires = expires
		c.list.MoveToFront(e)
	} else {
		e := c.list.Back()
		if e.key != nil {
			delete(c.m, e.key.(K))
		}
		e.key = k
		e.value = v
		e.expires = expires
		c.m[k] = e.index
		c.list.MoveToFront(e)
	}
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	c.del(key)
	c.mu.Unlock()
}

func (c *Cache[K, V]) MDelete(keys ...K) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *Cache[K, V]) del(key K) {
	idx, exists := c.m[key]
	if exists {
		delete(c.m, key)
		elem := c.list.get(idx)
		elem.key = nil
		elem.value = nil
		c.list.MoveToBack(elem)
	}
}

func (c *Cache[K, V]) promote(idx uint32) {
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
			newbuf.p = 0
			walbufpool.Put(newbuf)
			return
		}
	}

	// the oldbuf has been swapped, we take responsibility to flush it
	go func(c *Cache[K, V], buf *walbuf) {
		c.mu.Lock()
		c.flushBuf(buf)
		c.mu.Unlock()
		buf.reset()
		walbufpool.Put(buf)
	}(c, oldbuf)
}

func (c *Cache[K, V]) checkAndFlushBuf() {
	buf := (*walbuf)(c.buf)
	if buf.p > 0 {
		c.flushBuf(buf)
		buf.reset()
	}
}

func (c *Cache[K, V]) flushBuf(buf *walbuf) {
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
}
