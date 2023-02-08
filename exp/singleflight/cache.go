package singleflight

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// ErrFetchTimeout indicates a timeout error when refresh a cached value
// if CacheOptions.FetchTimeout is specified.
var ErrFetchTimeout = errors.New("fetch timeout")

// CacheOptions .
type CacheOptions struct {

	// FetchTimeout is used to timeout the fetch request if given,
	// default is zero (no timeout).
	//
	// NOTE: properly configured timeout will prevent task which take very long
	// time that don't fail fast, which may further block many requests, and
	// consume huge amount of resources, cause system overload or out of memory.
	FetchTimeout time.Duration

	// RefreshInterval specifies the interval to refresh the cache values,
	// default is zero which means don't refresh the cached values.
	//
	// If there is valid cache value and the subsequential fetch requests
	// failed, the cache value will be kept untouched.
	RefreshInterval time.Duration

	// ExpireInterval optionally enables a go goroutine to expire the cached
	// values, default is zero which means no expiration.
	//
	// Cached values are expired using a mark-then-delete strategy. In each
	// tick of expire duration, an active value will be marked as inactive,
	// if it's not accessed within the next expire interval, it will be
	// deleted from the Cache by the next expire execution.
	//
	// Each access of the cache value will touch it and mark it as active, which
	// prevents it being deleted from the cache.
	ExpireInterval time.Duration

	// Fetch is a function which retrieves data from upstream system for
	// the given key. The returned value or error will be cached till next
	// refresh execution.
	//
	// If RefreshInterval is provided, this option cannot be nil, or it panics.
	//
	// The provided function must return consistently typed values, or it
	// panics when storing a value of different type into the underlying
	// sync/atomic.Value.
	//
	// The returned value from this function should not be changed after
	// retrieved from the Cache, else it may panic since there may be many
	// goroutines access the same value concurrently.
	Fetch func(key string) (any, error)

	// ErrorCallback is an optional callback which will be called in a new
	// goroutine when error is returned by the Fetch function during refresh.
	ErrorCallback func(key string, err error)

	// ChangeCallback is an optional callback which will be called in a new
	// goroutine when new value is returned by the Fetch function during refresh.
	ChangeCallback func(key string, oldData, newData any)

	// DeleteCallback is an optional callback which will be called in a new
	// goroutine when a value is deleted from the cache.
	DeleteCallback func(key string, data any)
}

// Cache is an asynchronous cache which prevents duplicate functions calls
// that is massive or maybe expensive, or some data which rarely change
// and we want to get it fastly.
//
// Zero value of Cache is not ready to use. Use the function NewCache to
// make a new Cache instance. A Cache value shall not be copied after
// initialized.
type Cache struct {
	opt   CacheOptions
	group Group
	data  sync.Map

	exit chan struct{}
}

// NewCache returns a new Cache instance using the given options.
func NewCache(opt CacheOptions) *Cache {
	c := &Cache{
		opt:  opt,
		exit: make(chan struct{}),
	}
	if opt.RefreshInterval > 0 {
		go c.refresh()
	}
	if opt.ExpireInterval > 0 {
		go c.expire()
	}
	return c
}

// Close closes the Cache. It will signal the refresh and expire
// goroutines of the Cache to shut down.
//
// It should be called when the Cache is no longer needed, or may lead
// resource leaks.
func (c *Cache) Close() {
	close(c.exit)
}

// SetDefault sets the default value of a given key if it is new to the cache.
// The param val should not be nil, or it panics.
//
// It's useful to warm up the cache.
func (c *Cache) SetDefault(key string, val any) (exists bool) {
	ent := allocEntry(val, errDefaultVal)
	_, loaded := c.data.LoadOrStore(key, ent)
	return loaded
}

// Get tries to fetch a value corresponding to the given key from the cache
// first. If it's not cached, a calling of function Cache.Fetch will be fired
// and the result will be cached.
//
// If error occurs during the first fetching, the error will be cached until
// the subsequential fetching succeeded.
func (c *Cache) Get(key string) (any, error) {
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		ent.Touch()
		val, err := ent.Load()
		if err == errDefaultVal {
			err = nil
		}
		return val, err
	}

	// wait for the fetcher
	return c.doFetch(key, nil)
}

func (c *Cache) doFetch(key string, defaultVal any) (any, error) {
	if c.opt.FetchTimeout == 0 {
		return c.fetchNoTimeout(key, defaultVal)
	}
	return c.fetchWithTimeout(key, defaultVal)
}

func (c *Cache) fetchNoTimeout(key string, defaultVal any) (any, error) {
	val, err, _ := c.group.Do(key, func() (any, error) {
		val, err := c.opt.Fetch(key)
		if err != nil && defaultVal != nil {
			val, err = defaultVal, errDefaultVal
		}
		ent := allocEntry(val, err)
		c.data.Store(key, ent)
		return val, err
	})
	return val, err
}

func (c *Cache) fetchWithTimeout(key string, defaultVal any) (any, error) {
	timeout := time.NewTimer(c.opt.FetchTimeout)
	ch := c.group.DoChan(key, func() (any, error) {
		val, err := c.opt.Fetch(key)
		if err != nil && defaultVal != nil {
			val, err = defaultVal, errDefaultVal
		}
		ent := allocEntry(val, err)
		c.data.Store(key, ent)
		return val, err
	})
	select {
	case <-timeout.C:
		return nil, ErrFetchTimeout
	case result := <-ch:
		timeout.Stop()
		return result.Val, result.Err
	}
}

// GetOrDefault tries to fetch a value corresponding to the given key from
// the cache first. If it's not cached, a calling of function Cache.Fetch
// will be fired and the result will be cached.
//
// If the fetching fails, the default value will be set into the cache
// and returned.
func (c *Cache) GetOrDefault(key string, defaultVal any) any {
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		val, err := ent.Load()
		if err != nil {
			val = defaultVal
		}
		ent.Touch()
		return val
	}

	// fetch the value from upstream or use the default
	val, _ = c.doFetch(key, defaultVal)
	return val
}

// Delete deletes the entry of key from the cache if it is cached.
func (c *Cache) Delete(key string) {
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		val, _ := ent.Load()
		if val != nil && c.opt.DeleteCallback != nil {
			go c.opt.DeleteCallback(key, val)
		}
		c.data.Delete(key)
	}
}

// DeleteFunc iterates the cache and deletes entries that the key matches
// the given function.
func (c *Cache) DeleteFunc(match func(key string) bool) {
	hasDeleteCallback := c.opt.DeleteCallback != nil
	c.data.Range(func(key, val any) bool {
		keystr := key.(string)
		if match(keystr) {
			if hasDeleteCallback {
				ent := val.(*entry)
				val, _ := ent.Load()
				if val != nil {
					go c.opt.DeleteCallback(keystr, val)
				}
			}
			c.data.Delete(key)
		}
		return true
	})
}

func (c *Cache) refresh() {
	ticker := time.NewTicker(c.opt.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.doRefresh()
		case <-c.exit:
			return
		}
	}
}

func (c *Cache) doRefresh() {
	hasErrorCallback := c.opt.ErrorCallback != nil
	hasChangeCallback := c.opt.ErrorCallback != nil
	c.data.Range(func(key, val any) bool {
		keystr := key.(string)
		ent := val.(*entry)

		newVal, err := c.opt.Fetch(keystr)
		if err != nil {
			if hasErrorCallback {
				go c.opt.ErrorCallback(keystr, err)
			}
			_, oldErr := ent.Load()
			if oldErr != nil {
				ent.SetError(err)
			}
			return true
		}

		// save the fresh value
		if hasChangeCallback {
			oldVal, _ := ent.Load()
			go c.opt.ChangeCallback(keystr, oldVal, newVal)
		}
		ent.Store(newVal, nil)
		return true
	})
}

func (c *Cache) expire() {
	ticker := time.NewTicker(c.opt.ExpireInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.doExpire()
		case <-c.exit:
			return
		}
	}
}

func (c *Cache) doExpire() {
	hasDeleteCallback := c.opt.DeleteCallback != nil
	c.data.Range(func(key, val any) bool {
		keystr := key.(string)
		ent := val.(*entry)

		// If entry.expire is "active", we will mark it as "inactive" here.
		// Then during the next execution, "inactive" entries will be deleted.
		isActive := atomic.CompareAndSwapInt32(&ent.expire, active, inactive)
		if !isActive {
			if hasDeleteCallback {
				val, _ := ent.Load()
				if val != nil {
					go c.opt.DeleteCallback(keystr, val)
				}
			}
			c.data.Delete(key)
		}
		return true
	})
}

func allocEntry(val any, err error) *entry {
	ent := &entry{}
	ent.Store(val, err)
	return ent
}

type entry struct {
	val    atomic.Value
	errp   unsafe.Pointer // *error
	expire int32
}

func (e *entry) Load() (any, error) {
	val := e.val.Load()
	errp := atomic.LoadPointer(&e.errp)
	if errp == nil {
		return val, nil
	}
	return val, *(*error)(errp)
}

func (e *entry) Store(val any, err error) {
	if err != nil {
		atomic.StorePointer(&e.errp, unsafe.Pointer(&err))
		if val != nil {
			e.val.Store(val)
		}
	} else {
		e.val.Store(val)
		atomic.StorePointer(&e.errp, nil)
	}
}

func (e *entry) SetValue(val any) {
	if val != nil {
		e.val.Store(val)
	}
}

func (e *entry) SetError(err error) {
	atomic.StorePointer(&e.errp, unsafe.Pointer(&err))
}

func (e *entry) Touch() {
	atomic.StoreInt32(&e.expire, active)
}

const (
	active   = 0
	inactive = 1
)

var errDefaultVal = tombError(1)

type tombError int

func (e tombError) Error() string {
	return fmt.Sprintf("tombError(%d)", e)
}
