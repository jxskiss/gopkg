package lru

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	maxCapacity = 1<<32 - 1
	walBufSize  = 512
)

// NewCache returns a lru cache instance with given capacity, the underlying
// memory will will be immediately allocated. For best performance, the
// memory will be reused and won't be freed for the lifetime of the cache.
//
// Param capacity must be smaller than 2^32, else it will panic.
func NewCache(capacity int) *cache {
	if capacity > maxCapacity {
		panic("invalid too large capacity")
	}
	c := &cache{
		m:     make(map[interface{}]uint32, capacity),
		elems: make([]element, capacity),
		buf:   unsafe.Pointer(&walbuf{}),
		_buf:  unsafe.Pointer(&walbuf{}),
	}
	c.list = newList(c.elems)
	return c
}

type cache struct {
	mu   sync.RWMutex
	list *list
	m    map[interface{}]uint32

	buf  unsafe.Pointer // *walbuf
	_buf unsafe.Pointer // *walbuf

	elems []element
	flush int32
}

// walbuf helps to reduce lock-contention of read requests from the cache.
type walbuf struct {
	b [walBufSize]uint32
	p int32
}

// Len returns the number of cached values.
func (c *cache) Len() (n int) {
	c.mu.RLock()
	n = len(c.m)
	c.mu.RUnlock()
	return
}

// Get returns the cached value for the given key and updates its LRU score.
// The returned value may be expired, caller can check the returned value
// "expired" to check whether the value is expired.
func (c *cache) Get(key interface{}) (v interface{}, exists, expired bool) {
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

// GetQuiet returns the cached value for the given key, but don't modify its LRU score.
// The returned value may be expired, caller can check the returned value
// "expired" to check whether the value is expired.
func (c *cache) GetQuiet(key interface{}) (v interface{}, exists, expired bool) {
	c.mu.RLock()
	_, elem, exists := c.get(key)
	if exists {
		v = elem.value
		expired = elem.expires > 0 && elem.expires < time.Now().UnixNano()
	}
	c.mu.RUnlock()
	return
}

// GetNotStale returns the cached value for the given key. The returned value
// is guaranteed not expired. If unexpired value available, its LRU score
// will be updated.
func (c *cache) GetNotStale(key interface{}) (v interface{}, exists bool) {
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

func (c *cache) get(key interface{}) (idx uint32, elem *element, exists bool) {
	idx, exists = c.m[key]
	if exists {
		elem = &c.elems[idx]
	}
	return
}

// MGetInt64 returns map of cached values for the given int64 keys and
// update their LRU scores. The returned values may be expired.
// It's a convenient and efficient way to retrieve multiple values for
// int64 keys.
func (c *cache) MGetInt64(keys []int64) map[int64]interface{} {
	res := make(map[int64]interface{}, len(keys))
	c.mu.RLock()
	for _, key := range keys {
		idx, exists := c.m[key]
		if exists {
			elem := &c.elems[idx]
			res[key] = elem.value
			c.promote(idx)
		}
	}
	c.mu.RUnlock()
	return res
}

// MGetString returns map of cached values for the given string keys and
// update their LRU scores. The returned values may be expired.
// It's a convenient and efficient way to retrieve multiple values for
// string keys.
func (c *cache) MGetString(keys []string) map[string]interface{} {
	res := make(map[string]interface{}, len(keys))
	c.mu.RLock()
	for _, key := range keys {
		idx, exists := c.m[key]
		if exists {
			elem := &c.elems[idx]
			res[key] = elem.value
			c.promote(idx)
		}
	}
	c.mu.RUnlock()
	return res
}

func (c *cache) promote(idx uint32) {
	buf := (*walbuf)(atomic.LoadPointer(&c.buf))
	i := atomic.AddInt32(&buf.p, 1)
	if i <= walBufSize {
		buf.b[i-1] = idx
		return
	}

	// buffer is full, swap buffer
	oldbuf := buf
	newbuf := (*walbuf)(atomic.SwapPointer(&c._buf, nil))
	if newbuf != nil {
		atomic.StorePointer(&c.buf, unsafe.Pointer(newbuf))
	} else {
		newbuf = (*walbuf)(atomic.LoadPointer(&c.buf))
		if newbuf == oldbuf {
			newbuf = nil
		}
	}
	// in case of too high concurrency, discard current promotion
	if newbuf != nil {
		i = atomic.AddInt32(&newbuf.p, 1)
		if i <= walBufSize {
			newbuf.b[i-1] = idx
		}
	}

	// flush the full buffer
	if atomic.CompareAndSwapInt32(&c.flush, 0, 1) {
		go func(c *cache, buf *walbuf) {
			c.mu.Lock()
			c.flushBuf(buf)
			c._buf = unsafe.Pointer(buf)
			c.flush = 0
			c.mu.Unlock()
		}(c, oldbuf)
	}
}

// Set adds an item to the cache overwriting existing one if it exists.
func (c *cache) Set(key, value interface{}, ttl time.Duration) {
	var expires int64
	if ttl > 0 {
		expires = time.Now().UnixNano() + int64(ttl)
	}
	c.mu.Lock()
	c.checkAndFlushBuf()
	c.set(key, value, expires)
	c.mu.Unlock()
}

// MSet adds multiple items to the cache overwriting existing ones.
// Unlike calling Set multiple times, it acquires lock only once for
// multiple key-value pairs.
func (c *cache) MSet(kvmap interface{}, ttl time.Duration) {
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

func (c *cache) set(k, v interface{}, expires int64) {
	idx, exists := c.m[k]
	if exists {
		e := &c.elems[idx]
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

// Del removes a key from the cache if it exists.
func (c *cache) Del(key interface{}) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	c.del(key)
	c.mu.Unlock()
}

// MDelInt64 removes multiple int64 keys from the cache if exists.
// It's a convenient and efficient way to remove multiple int64 keys.
func (c *cache) MDelInt64(keys []int64) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

// MDelString removes multiple string keys from the cache if exists.
// It's a convenient and efficient way to remove multiple string keys.
func (c *cache) MDelString(keys []string) {
	c.mu.Lock()
	c.checkAndFlushBuf()
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *cache) del(key interface{}) {
	idx, exists := c.m[key]
	if exists {
		delete(c.m, key)
		elem := &c.elems[idx]
		elem.key = nil
		elem.value = nil
		c.list.MoveToBack(elem)
	}
}

func (c *cache) checkAndFlushBuf() {
	buf := (*walbuf)(c.buf)
	if buf.p > 0 {
		c.flushBuf(buf)
	}
}

func (c *cache) flushBuf(buf *walbuf) {
	l1 := buf.p
	if l1 > walBufSize {
		l1 = walBufSize
	}

	// remove duplicate idx in-place
	m := make(map[uint32]struct{}, l1/4)
	b, p := buf.b, l1-1
	for i := l1 - 1; i >= 0; i-- {
		idx := buf.b[i]
		if _, ok := m[idx]; !ok {
			m[idx] = struct{}{}
			b[p] = idx
			p--
		}
	}

	// promote elements by their access order
	for i := p + 1; i < l1; i++ {
		idx := b[i]
		elem := &c.elems[idx]
		c.list.MoveToFront(elem)
	}

	buf.p = 0
}
