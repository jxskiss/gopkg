package lru

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxCapacity = 1<<32 - 1
	walBufSize  = 512
)

func NewCache(capacity int) *cache {
	if capacity > maxCapacity {
		panic("invalid too large capacity")
	}
	c := &cache{
		m:     make(map[interface{}]uint32, capacity),
		elems: make([]element, capacity),
		buf:   &walbuf{},
	}
	c.list = newList(c.elems)
	return c
}

type cache struct {
	mu   sync.RWMutex
	list *list
	buf  *walbuf
	m    map[interface{}]uint32

	elems []element
	flush int32
}

type walbuf struct {
	b [walBufSize]uint32
	p int32
}

func (c *cache) Len() (n int) {
	c.mu.RLock()
	n = len(c.m)
	c.mu.RUnlock()
	return
}

func (c *cache) Get(key interface{}) (v interface{}, exists, expired bool) {
	c.mu.RLock()
	idx, elem, exists := c.get(key)
	if exists {
		v = elem.value
		expired = isExpired(elem.expires)
		c.promote(idx)
	}
	c.mu.RUnlock()
	return
}

func (c *cache) GetQuiet(key interface{}) (v interface{}, exists, expired bool) {
	c.mu.RLock()
	_, elem, exists := c.get(key)
	if exists {
		v = elem.value
		expired = isExpired(elem.expires)
	}
	c.mu.RUnlock()
	return
}

func (c *cache) GetNotStale(key interface{}) (v interface{}, exists bool) {
	c.mu.RLock()
	idx, elem, exists := c.get(key)
	if exists {
		if !isExpired(elem.expires) {
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

func (c *cache) MGetInt64(keys ...int64) map[int64]interface{} {
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

func (c *cache) MGetString(keys ...string) map[string]interface{} {
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
	buf := c.buf
	if i := atomic.AddInt32(&buf.p, 1); i > walBufSize {
		// wal buffer is full, discard current promotion and trigger flush
		if atomic.CompareAndSwapInt32(&c.flush, 0, 1) {
			go func() {
				c.mu.Lock()
				if oldbuf := c.buf; oldbuf.p > 0 {
					c.buf = &walbuf{}
					c.flushBuf(oldbuf)
				}
				c.mu.Unlock()
				atomic.StoreInt32(&c.flush, 0)
			}()
		}
	} else {
		buf.b[i-1] = idx
	}
}

func (c *cache) Set(key, value interface{}, ttl time.Duration) {
	expires := expires(ttl)
	c.mu.Lock()
	if c.buf.p > 0 {
		c.flushBuf(c.buf)
	}
	c.set(key, value, expires)
	c.mu.Unlock()
}

func (c *cache) MSet(kvmap interface{}, ttl time.Duration) {
	expires := expires(ttl)
	m := reflect.ValueOf(kvmap)
	keys := m.MapKeys()

	c.mu.Lock()
	if c.buf.p > 0 {
		c.flushBuf(c.buf)
	}
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

func (c *cache) Del(key interface{}) {
	c.mu.Lock()
	if c.buf.p > 0 {
		c.flushBuf(c.buf)
	}
	c.del(key)
	c.mu.Unlock()
}

func (c *cache) MDelInt64(keys ...int64) {
	c.mu.Lock()
	if c.buf.p > 0 {
		c.flushBuf(c.buf)
	}
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *cache) MDelString(keys ...string) {
	c.mu.Lock()
	if c.buf.p > 0 {
		c.flushBuf(c.buf)
	}
	for _, key := range keys {
		c.del(key)
	}
	c.mu.Unlock()
}

func (c *cache) del(key interface{}) {
	idx, exists := c.m[key]
	if exists {
		delete(c.m, key)
		c.list.MoveToBack(&c.elems[idx])
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
