package acache

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sync/singleflight"
)

// ErrFetchTimeout indicates a timeout error when refresh a cached value
// if Options.FetchTimeout is specified.
var ErrFetchTimeout = errors.New("fetch timeout")

// Options configures the behavior of Cache.
type Options struct {

	// FetchFunc is a function which retrieves data from upstream system for
	// the given key. The returned value or error will be cached till next
	// refresh execution.
	// FetchFunc must not be nil, else it panics.
	//
	// The provided function must return consistently typed values,
	// else it panics when storing a value of different type into the
	// underlying sync/atomic.Value.
	//
	// The returned value from this function should not be changed after
	// retrieved from the Cache, else data race happens since there may be
	// many goroutines access the same value concurrently.
	FetchFunc func(key string) (any, error)

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

func (p *Options) validate() {
	if p.FetchFunc == nil {
		panic("acache: Options.FetchFunc must not be nil")
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
	opt   Options
	group singleflight.Group
	data  sync.Map

	refreshTicker callbackTicker
	expireTicker  callbackTicker

	needRefresh  int32
	doingRefresh int32
	doingExpire  int32
	closed       int32
}

// NewCache returns a new Cache instance using the given options.
func NewCache(opt Options) *Cache {
	opt.validate()
	c := &Cache{
		opt: opt,
	}
	if opt.RefreshInterval > 0 {
		c.refreshTicker = newCallbackTicker(opt.RefreshInterval, c.doRefresh)
	}
	if opt.ExpireInterval > 0 {
		c.expireTicker = newCallbackTicker(opt.ExpireInterval, c.doExpire)
	}

	return c
}

// Close closes the Cache.
// It signals the refreshing and expiring goroutines to shut down.
//
// It should be called when the Cache is no longer needed,
// or may lead resource leaks.
func (c *Cache) Close() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return
	}
	if c.refreshTicker != nil {
		c.refreshTicker.Stop()
		c.refreshTicker = nil
	}
	if c.expireTicker != nil {
		c.expireTicker.Stop()
		c.expireTicker = nil
	}
}

// SetDefault sets the default value of a given key if it is new to the cache.
// The param val should not be nil, else it panics.
// The returned bool value indicates whether the key already exists in the cache,
// if it already exists, this is a no-op.
//
// It's useful to warm up the cache.
func (c *Cache) SetDefault(key string, value any) (exists bool) {
	if value == nil {
		panic("acache: value must not be nil")
	}
	ent := allocEntry(value, errDefaultVal)
	_, loaded := c.data.LoadOrStore(key, ent)
	return loaded
}

// Update sets a value for key into the cache.
// If key is not cached in the cache, it adds the given key value to the cache.
// The param val should not be nil, else it panics.
func (c *Cache) Update(key string, value any) {
	if value == nil {
		panic("acache: value must not be nil")
	}
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		ent.Store(value, nil)
	} else {
		ent := allocEntry(value, nil)
		c.data.Store(key, ent)
	}
}

// Contains tells whether the cache contains the specified key.
// It returns false if key is never accessed from the cache,
// true means that a value or an error for key exists in the cache.
func (c *Cache) Contains(key string) bool {
	_, ok := c.data.Load(key)
	return ok
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If it's not cached, a calling to function Options.FetchFunc
// will be fired and the result will be cached.
//
// If error occurs during the first fetching, the error will be cached until
// the subsequent fetching requests triggered by refreshing succeed.
// The cached error will be returned, it does not trigger a calling to
// Options.FetchFunc.
//
// If a default value is set by SetDefault, the default value will be used,
// it does not trigger a calling to Options.FetchFunc.
func (c *Cache) Get(key string) (any, error) {
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		ent.Touch()
		value, err := ent.Load()
		if err == errDefaultVal {
			err = nil
		}
		return value, err
	}

	// Wait the fetch function to get result.
	return c.doFetch(key, nil)
}

// GetOrDefault tries to fetch a value corresponding to the given key from
// the cache. If it's not cached, a calling to function Options.FetchFunc
// will be fired and the result will be cached.
//
// If error occurs during the first fetching, defaultVal will be set into
// the cache and returned.
func (c *Cache) GetOrDefault(key string, defaultVal any) any {
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		ent.Touch()
		value, err := ent.Load()
		if err != nil {
			value = defaultVal
		}
		return value
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
		value, _ := ent.Load()
		if value != nil && c.opt.DeleteCallback != nil {
			c.opt.DeleteCallback(key, value)
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
				value, _ := ent.Load()
				if value != nil {
					c.opt.DeleteCallback(keystr, value)
				}
			}
			c.data.Delete(key)
		}
		return true
	})
}

func (c *Cache) doRefresh(_ time.Time, ok bool) {
	if !ok {
		return
	}

	// The refreshing procedure may run longer than c.RefreshInterval,
	// we don't allow more than one refreshing jobs run simultaneously,
	// but allow only one job to run, in that case, we start another
	// refreshing immediately after the previous complete.
	atomic.StoreInt32(&c.needRefresh, 1)

	if atomic.CompareAndSwapInt32(&c.doingRefresh, 0, 1) {
		defer atomic.StoreInt32(&c.doingRefresh, 0)

		for atomic.CompareAndSwapInt32(&c.needRefresh, 1, 0) {
			if atomic.LoadInt32(&c.closed) > 0 {
				return
			}

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
	}
}

func (c *Cache) doExpire(_ time.Time, ok bool) {
	if !ok || atomic.LoadInt32(&c.closed) > 0 {
		return
	}

	if atomic.CompareAndSwapInt32(&c.doingExpire, 0, 1) {
		defer atomic.StoreInt32(&c.doingExpire, 0)

		hasDeleteCallback := c.opt.DeleteCallback != nil
		c.data.Range(func(key, val any) bool {
			keystr := key.(string)
			ent := val.(*entry)

			// If entry.expire is "active", we will mark it as "inactive" here.
			// Then during the next execution, "inactive" entries will be deleted.
			isActive := atomic.CompareAndSwapInt32(&ent.expire, active, inactive)
			if !isActive {
				if hasDeleteCallback {
					value, _ := ent.Load()
					if value != nil {
						c.opt.DeleteCallback(keystr, value)
					}
				}
				c.data.Delete(key)
			}
			return true
		})
	}
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
