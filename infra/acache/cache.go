package acache

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sync/singleflight"

	"github.com/jxskiss/gopkg/v2/collection/heapx"
	"github.com/jxskiss/gopkg/v2/internal/functicker"
)

// ErrFetchTimeout indicates a timeout error when refresh a cached value
// if Options.FetchTimeout is specified.
var ErrFetchTimeout = errors.New("fetch timeout")

// Options configures the behavior of Cache.
type Options struct {

	// Fetcher fetches data from upstream system for a given key.
	// Result value or error will be cached till next refresh execution.
	//
	// The provided Fetcher implementation must return consistently
	// typed values, else it panics when storing a value of different
	// type into the underlying sync/atomic.Value.
	//
	// The returned value from this function should not be changed after
	// retrieved from the Cache, else data race happens since there may be
	// many goroutines access the same value concurrently.
	Fetcher Fetcher

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
	// an error is returned by Fetcher during refresh.
	ErrorCallback func(err error, keys []string)

	// ChangeCallback is an optional callback which will be called when
	// new value is returned by Fetcher during refresh.
	ChangeCallback func(key string, oldData, newData any)

	// DeleteCallback is an optional callback which will be called when
	// a value is deleted from the cache.
	DeleteCallback func(key string, data any)
}

func (p *Options) validate() {
	if p.Fetcher == nil {
		panic("acache: Options.Fetcher must not be nil")
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
	opt     Options
	sfGroup singleflight.Group
	data    sync.Map

	mu           sync.Mutex
	refreshQueue *heapx.PriorityQueue[int64, string]

	ticker       *functicker.Ticker
	preExpireAt  atomic.Int64
	doingExpire  int32
	doingRefresh int32
	closed       int32
}

// NewCache returns a new Cache instance using the given options.
func NewCache(opt Options) *Cache {
	tickInterval := time.Second
	return newCacheWithTickInterval(opt, tickInterval)
}

func newCacheWithTickInterval(opt Options, tickInterval time.Duration) *Cache {
	opt.validate()
	c := &Cache{
		opt:          opt,
		refreshQueue: heapx.NewMinPriorityQueue[int64, string](),
	}
	if opt.ExpireInterval > 0 || opt.RefreshInterval > 0 {
		c.ticker = functicker.New(tickInterval, c.runBackgroundTasks)
	}
	return c
}

// Close closes the Cache.
// It signals the background goroutines to shut down.
//
// It should be called when the Cache is no longer needed,
// or may lead resource leaks.
func (c *Cache) Close() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return
	}
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
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
	nowNano := time.Now().UnixNano()
	ent := allocEntry(value, errDefaultVal, nowNano)
	actual, loaded := c.data.LoadOrStore(key, ent)
	if loaded {
		actual.(*entry).MarkActive()
	} else {
		c.addToRefreshQueue(nowNano, key)
	}
	return loaded
}

// Update sets a value for key into the cache.
// If key is not cached in the cache, it adds the given key value to the cache.
// The param value should not be nil, else it panics.
func (c *Cache) Update(key string, value any) {
	if value == nil {
		panic("acache: value must not be nil")
	}
	nowNano := time.Now().UnixNano()
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		ent.Store(value, nil, nowNano)
	} else {
		ent := allocEntry(value, nil, nowNano)
		c.data.Store(key, ent)
	}
	c.addToRefreshQueue(nowNano, key)
}

// Contains tells whether the cache contains the specified key.
// It returns false if key is never accessed from the cache,
// true means that a value or an error for key exists in the cache.
func (c *Cache) Contains(key string) bool {
	_, ok := c.data.Load(key)
	return ok
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If it's not cached, a calling to Fetcher.Fetch will be fired
// and the result will be cached.
//
// If error occurs during the first fetching, the error will be cached until
// the subsequent fetching requests triggered by refreshing succeed.
// The cached error will be returned, it does not trigger a calling to
// Options.Fetcher again.
//
// If a default value is set by SetDefault, the default value will be used,
// it does not trigger a calling to Options.Fetcher.
func (c *Cache) Get(key string) (any, error) {
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		ent.MarkActive()
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
// the cache. If it's not cached, a calling to Options.Fetcher
// will be fired and the result will be cached.
//
// If error occurs during the first fetching, defaultVal will be set into
// the cache and returned, the default value will also be used for
// further calling of Get and GetOrDefault.
func (c *Cache) GetOrDefault(key string, defaultVal any) any {
	val, ok := c.data.Load(key)
	if ok {
		ent := val.(*entry)
		ent.MarkActive()
		value, err := ent.Load()
		if err != nil {
			value = defaultVal
		}
		return value
	}

	// Fetch result from upstream or use the default value
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
	val, err, _ := c.sfGroup.Do(key, func() (any, error) {
		val, err := c.opt.Fetcher.Fetch(key)
		if err != nil && defaultVal != nil {
			val, err = defaultVal, errDefaultVal
		}
		nowNano := time.Now().UnixNano()
		ent := allocEntry(val, err, nowNano)
		c.data.Store(key, ent)
		c.addToRefreshQueue(nowNano, key)
		return val, err
	})
	return val, err
}

func (c *Cache) fetchWithTimeout(key string, defaultVal any) (any, error) {
	timeout := time.NewTimer(c.opt.FetchTimeout)
	ch := c.sfGroup.DoChan(key, func() (any, error) {
		val, err := c.opt.Fetcher.Fetch(key)
		if err != nil && defaultVal != nil {
			val, err = defaultVal, errDefaultVal
		}
		nowNano := time.Now().UnixNano()
		ent := allocEntry(val, err, nowNano)
		c.data.Store(key, ent)
		c.addToRefreshQueue(nowNano, key)
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

func (c *Cache) addToRefreshQueue(updateAtNano int64, key string) {
	c.mu.Lock()
	c.refreshQueue.Push(updateAtNano, key)
	c.mu.Unlock()
}

func (c *Cache) runBackgroundTasks() {
	if atomic.LoadInt32(&c.closed) > 0 {
		return
	}
	if c.opt.ExpireInterval > 0 {
		c.doExpire(false)
	}
	if c.opt.RefreshInterval > 0 {
		if atomic.CompareAndSwapInt32(&c.doingRefresh, 0, 1) {
			c.doRefresh()
			atomic.StoreInt32(&c.doingRefresh, 0)
		}
	}
}

func (c *Cache) doExpire(force bool) {
	nowUnix := time.Now().Unix()
	preExpireAt := c.preExpireAt.Load()

	// "force" helps to do unittest.
	if !force {
		if preExpireAt == 0 {
			c.preExpireAt.Store(nowUnix)
			return
		}
		if time.Duration(nowUnix-preExpireAt)*time.Second < c.opt.ExpireInterval {
			return
		}
	}

	hasDeleteCallback := c.opt.DeleteCallback != nil
	if atomic.CompareAndSwapInt32(&c.doingExpire, 0, 1) {
		defer atomic.StoreInt32(&c.doingExpire, 0)
		c.preExpireAt.Store(nowUnix)
		c.data.Range(func(key, val any) bool {
			keystr := key.(string)
			ent := val.(*entry)

			// If entry.expire is "active", we mark it as "inactive" here.
			// Then during the next execution, "inactive" entries will be deleted.
			isActive := atomic.CompareAndSwapInt32(&ent.expire, active, inactive)
			if !isActive {
				if hasDeleteCallback {
					value, _ := ent.Load()
					if value != nil {
						c.opt.DeleteCallback(keystr, val)
					}
				}
				c.data.Delete(key)
			}
			return true
		})
	}
}

func (c *Cache) needRefresh(nowNano, updateAtNano int64) bool {
	return time.Duration(nowNano-updateAtNano) >= c.opt.RefreshInterval
}

func (c *Cache) checkEntryNeedRefresh(nowNano, updateAtNano int64, key string) (ent *entry, refresh bool) {
	val, _ := c.data.Load(key)
	if val == nil {
		// The data has already been deleted.
		return nil, false
	}
	ent = val.(*entry)
	if ent.GetUpdateAt() != updateAtNano {
		// The data has already been changed.
		return ent, false
	}
	if !c.needRefresh(nowNano, updateAtNano) {
		return ent, false
	}
	return ent, true
}

func (c *Cache) doRefresh() {
	if _, ok := c.opt.Fetcher.(BatchFetcher); ok {
		c.doBatchRefresh()
		return
	}

	hasErrorCallback := c.opt.ErrorCallback != nil
	hasChangeCallback := c.opt.ChangeCallback != nil
	for {
		nowNano := time.Now().UnixNano()
		needRefresh := false

		c.mu.Lock()
		updateAt, key, ok := c.refreshQueue.Peek()
		if ok && c.needRefresh(nowNano, updateAt) {
			c.refreshQueue.Pop()
			needRefresh = true
		}
		c.mu.Unlock()
		if !needRefresh {
			break
		}

		var ent *entry
		ent, needRefresh = c.checkEntryNeedRefresh(nowNano, updateAt, key)
		if !needRefresh {
			continue
		}
		newVal, err := c.opt.Fetcher.Fetch(key)
		if err != nil {
			if hasErrorCallback {
				c.opt.ErrorCallback(err, []string{key})
			}
			_, oldErr := ent.Load()
			if oldErr != nil {
				ent.SetError(err)
			}
		} else {
			// Save the new value from upstream.
			if hasChangeCallback {
				oldVal, _ := ent.Load()
				c.opt.ChangeCallback(key, oldVal, newVal)
			}
			ent.Store(newVal, nil, nowNano)
		}
		c.addToRefreshQueue(nowNano, key)
	}
}

type refreshQueueItem struct {
	key   string
	tNano int64
}

func (c *Cache) doBatchRefresh() {
	fetcher := c.opt.Fetcher.(BatchFetcher)
	batchSize := fetcher.BatchSize()
	expiredItems := make([]refreshQueueItem, 0, batchSize)
	keys := make([]string, 0, batchSize)
	for {
		nowNano := time.Now().UnixNano()
		expiredItems = expiredItems[:0]
		keys = keys[:0]

		c.mu.Lock()
		for len(expiredItems) < batchSize {
			updateAt, key, ok := c.refreshQueue.Peek()
			if !ok || !c.needRefresh(nowNano, updateAt) {
				break
			}
			c.refreshQueue.Pop()
			expiredItems = append(expiredItems, refreshQueueItem{key, updateAt})
		}
		c.mu.Unlock()

		for _, item := range expiredItems {
			_, needRefresh := c.checkEntryNeedRefresh(nowNano, item.tNano, item.key)
			if !needRefresh {
				continue
			}
			keys = append(keys, item.key)
		}
		if len(keys) > 0 {
			c.batchRefreshKeys(keys)
		}

		// No more expired data to refresh.
		if len(expiredItems) < batchSize {
			break
		}
	}
}

func (c *Cache) batchRefreshKeys(keys []string) {
	nowNano := time.Now().UnixNano()
	fetcher := c.opt.Fetcher.(BatchFetcher)
	newValMap, err := fetcher.BatchFetch(keys)
	if err != nil {
		hasErrorCallback := c.opt.ErrorCallback != nil
		if hasErrorCallback {
			c.opt.ErrorCallback(err, keys)
		}
		for _, key := range keys {
			val, _ := c.data.Load(key)
			if val == nil {
				continue
			}
			ent := val.(*entry)
			_, oldErr := ent.Load()
			if oldErr != nil {
				ent.SetError(err)
			}
		}
		return
	}
	hasChangeCallback := c.opt.ChangeCallback != nil
	for key, newVal := range newValMap {
		entVal, _ := c.data.Load(key)
		if entVal == nil {
			continue
		}
		ent := entVal.(*entry)
		if hasChangeCallback {
			oldVal, _ := ent.Load()
			c.opt.ChangeCallback(key, oldVal, newVal)
		}
		ent.Store(newVal, nil, nowNano)
		c.addToRefreshQueue(nowNano, key)
	}
}

func allocEntry(val any, err error, updateAtNano int64) *entry {
	ent := &entry{}
	ent.Store(val, err, updateAtNano)
	return ent
}

type entry struct {
	val      atomic.Value
	errp     unsafe.Pointer // *error
	updateAt int64
	expire   int32
}

func (e *entry) Load() (any, error) {
	val := e.val.Load()
	errp := atomic.LoadPointer(&e.errp)
	if errp == nil {
		return val, nil
	}
	return val, *(*error)(errp)
}

func (e *entry) Store(val any, err error, updateAtNano int64) {
	if err != nil {
		atomic.StorePointer(&e.errp, unsafe.Pointer(&err))
		if val != nil {
			e.val.Store(val)
		}
	} else if val != nil {
		e.val.Store(val)
		atomic.StorePointer(&e.errp, nil)
	}
	e.SetUpdateAt(updateAtNano)
}

func (e *entry) SetError(err error) {
	atomic.StorePointer(&e.errp, unsafe.Pointer(&err))
}

func (e *entry) GetUpdateAt() int64 {
	return atomic.LoadInt64(&e.updateAt)
}

func (e *entry) SetUpdateAt(updateAtNano int64) {
	atomic.StoreInt64(&e.updateAt, updateAtNano)
}

func (e *entry) MarkActive() {
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
