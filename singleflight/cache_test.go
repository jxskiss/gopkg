package singleflight

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	var key, val = "key", "val"
	opt := CacheOptions{
		RefreshInterval: 50 * time.Millisecond,
		Fetch: func(key string) (interface{}, error) {
			return val, nil
		},
	}
	c := NewCache(opt)

	got, err := c.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, got)

	time.Sleep(opt.RefreshInterval / 2)
	val = "newVal"
	got, err = c.Get(key)
	assert.Nil(t, err)
	assert.NotEqual(t, val, got)

	time.Sleep(opt.RefreshInterval)
	got, err = c.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, got)
}

func TestGetError(t *testing.T) {
	var key, val = "key", "val"
	var first = true
	opt := CacheOptions{
		RefreshInterval: 50 * time.Millisecond,
		Fetch: func(key string) (interface{}, error) {
			if first {
				first = false
				return nil, errors.New("error")
			}
			return val, nil
		},
	}
	c := NewCache(opt)

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
	var key, val, defaultVal = "key", "val", "default"
	opt := CacheOptions{
		RefreshInterval: 50 * time.Millisecond,
		Fetch: func(key string) (interface{}, error) {
			return val, nil
		},
	}
	c := NewCache(opt)

	got := c.GetOrDefault(key, defaultVal)
	assert.Equal(t, val, got)

	time.Sleep(opt.RefreshInterval / 2)
	val = "newVal"
	got = c.GetOrDefault(key, defaultVal)
	assert.NotEqual(t, val, got)

	time.Sleep(opt.RefreshInterval)
	got = c.GetOrDefault(key, defaultVal)
	assert.Equal(t, val, got)
}

func TestGetOrDefaultError(t *testing.T) {
	var key, val, defaultVal = "key", "val", "default"
	var first = true
	opt := CacheOptions{
		RefreshInterval: 50 * time.Millisecond,
		Fetch: func(key string) (interface{}, error) {
			if first {
				first = false
				return nil, errors.New("error")
			}
			return val, nil
		},
	}
	c := NewCache(opt)

	got := c.GetOrDefault(key, defaultVal)
	assert.Equal(t, defaultVal, got)

	time.Sleep(opt.RefreshInterval / 2)
	got = c.GetOrDefault(key, val)
	assert.NotEqual(t, val, got)
	assert.Equal(t, defaultVal, got)

	time.Sleep(opt.RefreshInterval)
	got = c.GetOrDefault(key, defaultVal)
	assert.Equal(t, val, got)
}

func TestSetDefault(t *testing.T) {
	opt := CacheOptions{
		RefreshInterval: 50 * time.Millisecond,
		Fetch: func(key string) (interface{}, error) {
			return nil, errors.New("error")
		},
	}
	c := NewCache(opt)

	got := c.GetOrDefault("key1", "default1")
	assert.Equal(t, "default1", got)

	exist := c.SetDefault("key2", "val2")
	assert.False(t, exist)
	got = c.GetOrDefault("key2", "default2")
	assert.Equal(t, "val2", got)

	exist = c.SetDefault("key2", "val3")
	assert.True(t, exist)
	got = c.GetOrDefault("key2", "default2")
	assert.Equal(t, "val2", got)
}

func TestDeleteFunc(t *testing.T) {
	opt := CacheOptions{
		RefreshInterval: 50 * time.Millisecond,
		Fetch: func(key string) (interface{}, error) {
			return nil, errors.New("error")
		},
	}
	c := NewCache(opt)

	c.SetDefault("key", "val")
	got := c.GetOrDefault("key", "default")
	assert.Equal(t, "val", got)

	c.DeleteFunc(func(string) bool { return true })

	got = c.GetOrDefault("key", "default")
	assert.Equal(t, "default", got)
}

func TestClose(t *testing.T) {
	var sleep = 50 * time.Millisecond
	var count int
	opt := CacheOptions{
		RefreshInterval: sleep - 10*time.Millisecond,
		Fetch: func(key string) (interface{}, error) {
			count++
			return count, nil
		},
	}
	c := NewCache(opt)

	got := c.GetOrDefault("key", 10)
	assert.Equal(t, 1, got)

	time.Sleep(sleep)
	got = c.GetOrDefault("key", 10)
	assert.Equal(t, 2, got)

	time.Sleep(sleep)
	got = c.GetOrDefault("key", 10)
	assert.Equal(t, 3, got)

	c.Close()

	time.Sleep(sleep)
	got = c.GetOrDefault("key", 10)
	assert.Equal(t, 3, got)
}

func TestExpire(t *testing.T) {
	// trigger is used to mark whether fetch is called
	trigger := false
	opt := CacheOptions{
		ExpireInterval:  3 * time.Minute,
		RefreshInterval: time.Minute,
		Fetch: func(key string) (interface{}, error) {
			trigger = true
			return "", nil
		},
	}
	c := NewCache(opt)

	// GetOrDefault cannot trigger fetch after SetDefault
	c.SetDefault("default", "")
	c.SetDefault("alive", "")
	c.GetOrDefault("alive", "")
	assert.False(t, trigger)

	c.Get("expire")
	assert.True(t, trigger)

	// first expire will mark entries as inactive
	c.doExpire()

	trigger = false
	c.Get("alive")
	assert.False(t, trigger)

	// second expire, both default & expire will be removed
	c.doExpire()

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
