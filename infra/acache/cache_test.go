package acache

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	var key = "key"
	var _val atomic.Value
	_val.Store("val")

	val := func() string {
		return _val.Load().(string)
	}
	opt := Options{
		RefreshInterval: 50 * time.Millisecond,
		Fetcher: FuncFetcher(func(key string) (any, error) {
			return val(), nil
		}),
	}
	c := newCacheWithTickInterval(opt, 10*time.Millisecond)
	defer c.Close()
	assert.False(t, c.Contains(key))

	got, err := c.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val(), got)
	assert.True(t, c.Contains(key))

	time.Sleep(opt.RefreshInterval / 2)
	_val.Store("newVal")

	got, err = c.Get(key)
	assert.Nil(t, err)
	assert.NotEqual(t, val(), got)
	assert.True(t, c.Contains(key))

	time.Sleep(opt.RefreshInterval + time.Second)
	got, err = c.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val(), got)
}

func TestGetError(t *testing.T) {
	var key, val = "key", "val"
	var first = true
	opt := Options{
		RefreshInterval: 50 * time.Millisecond,
		Fetcher: FuncFetcher(func(key string) (any, error) {
			if first {
				first = false
				return nil, errors.New("error")
			}
			return val, nil
		}),
	}
	c := newCacheWithTickInterval(opt, 10*time.Millisecond)
	defer c.Close()

	got, err := c.Get(key)
	assert.NotNil(t, err)
	assert.Nil(t, got)

	time.Sleep(opt.RefreshInterval / 2)
	_, err2 := c.Get(key)
	assert.Equal(t, err, err2)

	time.Sleep(opt.RefreshInterval + 10*time.Millisecond)
	got, err = c.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, got)
}

func TestGetOrDefault(t *testing.T) {
	var (
		mu                   sync.Mutex
		key, val, defaultVal = "key", "val", "default"
	)
	opt := Options{
		RefreshInterval: 50 * time.Millisecond,
		Fetcher: FuncFetcher(func(key string) (any, error) {
			mu.Lock()
			defer mu.Unlock()
			return val, nil
		}),
	}
	c := newCacheWithTickInterval(opt, 10*time.Millisecond)
	defer c.Close()

	got := c.GetOrDefault(key, defaultVal)
	assert.Equal(t, val, got)

	time.Sleep(opt.RefreshInterval / 2)

	// update val
	mu.Lock()
	val = "newVal"
	mu.Unlock()

	got = c.GetOrDefault(key, defaultVal)
	assert.NotEqual(t, val, got)

	time.Sleep(opt.RefreshInterval)
	got = c.GetOrDefault(key, defaultVal)
	assert.Equal(t, val, got)
}

func TestGetOrDefaultError(t *testing.T) {
	var key, val, defaultVal1, defaultVal2 = "key", "val", "default1", "default2"
	var first = true
	opt := Options{
		RefreshInterval: 50 * time.Millisecond,
		Fetcher: FuncFetcher(func(key string) (any, error) {
			if first {
				first = false
				return nil, errors.New("error")
			}
			return val, nil
		}),
	}
	c := newCacheWithTickInterval(opt, 10*time.Millisecond)
	defer c.Close()

	// First loading, error happens, should get defaultVal1.
	got := c.GetOrDefault(key, defaultVal1)
	assert.Equal(t, defaultVal1, got)

	// The second loading has not been triggered, should get defaultVal2.
	time.Sleep(opt.RefreshInterval / 2)
	got = c.GetOrDefault(key, defaultVal2)
	assert.Equal(t, defaultVal2, got)

	// RefreshInterval has been passed, the second loading has been triggered,
	// we should get "val" from the loader function.
	time.Sleep(opt.RefreshInterval)
	runtime.Gosched()
	got = c.GetOrDefault(key, defaultVal1)
	assert.Equal(t, val, got)
}

func TestSetDefault(t *testing.T) {
	opt := Options{
		RefreshInterval: 50 * time.Millisecond,
		Fetcher: FuncFetcher(func(key string) (any, error) {
			return nil, errors.New("error")
		}),
	}
	c := newCacheWithTickInterval(opt, 10*time.Millisecond)
	defer c.Close()

	got := c.GetOrDefault("key1", "default1")
	assert.Equal(t, "default1", got)

	exist := c.SetDefault("key2", "val2")
	assert.False(t, exist)
	got = c.GetOrDefault("key2", "default2")
	assert.Equal(t, "default2", got)

	// Only the first call of `SetDefault` take effect.
	exist = c.SetDefault("key2", "val3")
	assert.True(t, exist)
	got, err := c.Get("key2")
	assert.Nil(t, err)
	assert.Equal(t, "val2", got)
	got = c.GetOrDefault("key2", "default2")
	assert.Equal(t, "default2", got)
}

func TestUpdate(t *testing.T) {
	opt := Options{
		Fetcher: FuncFetcher(func(key string) (any, error) {
			if key == "testError" {
				return nil, errors.New("test error")
			}
			return "val", nil
		}),
	}
	c := newCacheWithTickInterval(opt, 10*time.Millisecond)
	defer c.Close()

	got1, err1 := c.Get("key1")
	assert.Nil(t, err1)
	assert.Equal(t, "val", got1)

	got2 := c.GetOrDefault("key2", "defaultVal")
	assert.Equal(t, "val", got2)

	c.SetDefault("key3", "defaultVal")
	got3, err3 := c.Get("key3")
	assert.Nil(t, err3)
	assert.Equal(t, "defaultVal", got3)

	got4 := c.GetOrDefault("testError", "defaultVal")
	assert.Equal(t, "defaultVal", got4)

	c.Update("key1", "updateVal")
	c.Update("key2", "updateVal")
	c.Update("key3", "updateVal")
	c.Update("testError", "updateVal")
	c.Update("key4", "updateVal")

	_get := func(k string) any {
		ret, _ := c.Get(k)
		return ret
	}
	assert.Equal(t, "updateVal", _get("key1"))
	assert.Equal(t, "updateVal", _get("key2"))
	assert.Equal(t, "updateVal", _get("key3"))
	assert.Equal(t, "updateVal", _get("testError"))
	assert.Equal(t, "updateVal", _get("key4"))
}

func TestDeleteFunc(t *testing.T) {
	opt := Options{
		RefreshInterval: 50 * time.Millisecond,
		Fetcher: FuncFetcher(func(key string) (any, error) {
			return nil, errors.New("error")
		}),
	}
	c := newCacheWithTickInterval(opt, 10*time.Millisecond)
	defer c.Close()

	c.SetDefault("key", "val")
	got := c.GetOrDefault("key", "default")
	assert.Equal(t, "default", got)

	c.DeleteFunc(func(string) bool { return true })

	got = c.GetOrDefault("key", "default")
	assert.Equal(t, "default", got)
}

func TestClose(t *testing.T) {
	// Timer on Windows platform is inaccurate, which cause this test fails
	// randomly, skip it.
	if runtime.GOOS == "windows" {
		t.Skip("Skip acache.TestClose on the Windows platform")
		return
	}

	var sleep = 200 * time.Millisecond
	var count int64
	opt := Options{
		RefreshInterval: sleep - 10*time.Millisecond,
		Fetcher: FuncFetcher(func(key string) (any, error) {
			x := atomic.AddInt64(&count, 1)
			return int(x), nil
		}),
	}
	c := newCacheWithTickInterval(opt, 10*time.Millisecond)
	defer c.Close()

	got := c.GetOrDefault("key", 10)
	assert.Equal(t, 1, got)

	time.Sleep(sleep)
	got = c.GetOrDefault("key", 10)
	assert.True(t, got == 1 || got == 2)

	time.Sleep(sleep)
	got = c.GetOrDefault("key", 10)
	assert.True(t, got == 2 || got == 3)

	c.Close()

	time.Sleep(5 * sleep)
	got = c.GetOrDefault("key", 10)
	assert.True(t, got == 3 || got == 4)
}

func TestExpire(t *testing.T) {
	// trigger is used to mark whether fetch is called
	trigger := false
	opt := Options{
		ExpireInterval:  3 * time.Minute,
		RefreshInterval: time.Minute,
		Fetcher: FuncFetcher(func(key string) (any, error) {
			trigger = true
			return "", nil
		}),
	}
	c := NewCache(opt)
	defer c.Close()

	// GetOrDefault cannot trigger fetch after SetDefault
	c.SetDefault("default", "")
	c.SetDefault("alive", "")
	c.GetOrDefault("alive", "")
	assert.False(t, trigger)

	c.Get("expire")
	assert.True(t, trigger)

	// first expire will mark entries as inactive
	c.doExpire(true)

	trigger = false
	c.Get("alive")
	assert.False(t, trigger)

	// second expire, both default & expire will be removed
	c.doExpire(true)

	// make sure refresh does not affect expire
	c.doRefresh()

	trigger = false
	c.Get("alive")
	assert.False(t, trigger)

	trigger = false
	c.Get("default")
	assert.True(t, trigger)

	trigger = false
	c.Get("expire")
	assert.True(t, trigger)
}
