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

// CacheOptions configures the behavior of Cache.
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
	// If there is valid cache value and the subsequent fetch requests
	// failed, the existing cache value will be kept untouched.
	RefreshInterval time.Duration

	// ExpireInterval optionally enables purging unused cached values,
	// default is zero which means no expiration.
	//
	// Note this is mainly used to purge unused data to prevent the cache
	// growing endlessly, the timing is inaccurate. Also note that
	// it may delete unused default values set by SetDefault.
	//
	// Cached values are deleted using a mark-then-delete strategy.
	// In each tick of expire interval, an active value will be marked as inactive,
	// if it's not accessed within the next expire interval, it will be
	// deleted from the Cache by the next expire execution.
	//
	// Each access of the cache value will touch it and mark it as active, which
	// prevents it being deleted from the cache.
	ExpireInterval time.Duration

	// FetchFunc is a function which retrieves data from upstream system for
	// the given key. The returned value or error will be cached till next
	// refresh execution.
	// FetchFunc must not be nil, or it panics.
	//
	// The provided function must return consistently typed values, or it
	// panics when storing a value of different type into the underlying
	// sync/atomic.Value.
	//
	// The returned value from this function should not be changed after
	// retrieved from the Cache, else data race happens since there may be
	// many goroutines access the same value concurrently.
	FetchFunc func(key string) (any, error)

	// ErrorCallback is an optional callback which will be called when
	// an error is returned by the FetchFunc during refresh.
	ErrorCallback func(key string, err error)

	// ChangeCallback is an optional callback which will be called when
	// new value is returned by the FetchFunc during refresh.
	ChangeCallback func(key string, oldData, newData any)

	// DeleteCallback is an optional callback which will be called when
	// a value is deleted from the cache.
	DeleteCallback func(key string, data any)
}

func (p *CacheOptions) validate() {
	if p.FetchFunc == nil {
		panic("CacheOptions.FetchFunc must not be nil")
	}
}

// Cache is an asynchronous cache which prevents duplicate functions calls
// that is massive or maybe expensive, or some data which rarely change,
// and we want to get it quickly.
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
	opt.validate()
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

// Close closes the Cache.
// It signals the refreshing and expiring goroutines to shut down.
//
// It should be called when the Cache is no longer needed,
// or may lead resource leaks.
func (c *Cache) Close() {
	close(c.exit)
}

// SetDefault sets the default value of a given key if it is new to the cache.
// The param val should not be nil, or it panics.
//
// It's useful to warm up the cache.
func (c *Cache) SetDefault(key string, val any) (exists bool) {
	if val == nil {
		panic("default value must not be nil")
	}
	ent := allocEntry(val, errDefaultVal)
	_, loaded := c.data.LoadOrStore(key, ent)
	return loaded
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If it's not cached, a calling of function FetchFunc will be fired
// and the result will be cached.
//
// If error occurs during the first fetching, the error will be cached until
// the subsequent fetching succeeded.
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

	// Wait the fetch function to get result.
	return c.doFetch(key, nil)
}

// GetOrDefault tries to fetch a value corresponding to the given key from
// the cache. If it's not cached, a calling of function FetchFunc will be
// fired and the result will be cached.
//
// If error occurs during the first fetching, defaultVal will be set into
// the cache and returned.
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

	// Fetch result from upstream or use the default
	val, _ = c.doFetch(key, defaultVal)
	return val
}

func (c *Cache) doFetch(key string, defaultVal any) (any, error) {
	if c.opt.FetchTimeout == 0 {
		return c.fetchNoTimeout(key, defaultVal)
	}
	return c.fetchWithTimeout(key, defaultVal)
}

func (c *Cache) fetchNoTimeout(key string, defaultVal any) (any, error) {
	val, err, _ := c.group.Do(key, func() (any, error) {
		val, err := c.opt.FetchFunc(key)
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
		val, err := c.opt.FetchFunc(key)
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

// Delete deletes the entry of key from the cache if it exists.
func (c *Cache) Delete(key string) {
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		val, _ := ent.Load()
		if val != nil && c.opt.DeleteCallback != nil {
			c.opt.DeleteCallback(key, val)
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
					c.opt.DeleteCallback(keystr, val)
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

		newVal, err := c.opt.FetchFunc(keystr)
		if err != nil {
			if hasErrorCallback {
				c.opt.ErrorCallback(keystr, err)
			}
			_, oldErr := ent.Load()
			if oldErr != nil {
				ent.SetError(err)
			}
			return true
		}

		// Save the new value from upstream.
		if hasChangeCallback {
			oldVal, _ := ent.Load()
			c.opt.ChangeCallback(keystr, oldVal, newVal)
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
					c.opt.DeleteCallback(keystr, val)
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
		e.SetValue(val)
	} else if val != nil {
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
