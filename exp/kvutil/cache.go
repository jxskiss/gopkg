package kvutil

import (
	"context"
	"encoding"
	"reflect"
	"time"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/easy"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/internal/rtype"
	"github.com/jxskiss/gopkg/v2/lru"
)

// DefaultBatchSize is the default batch size for batch operations.
const DefaultBatchSize = 100

// DefaultLRUExpiration is the default expiration time for data in LRU cache.
const DefaultLRUExpiration = time.Second

// Model is the interface implemented by types that can be cached by Cache.
type Model interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// KVPair represents a key value pair to work with Cache.
type KVPair struct {
	K string
	V []byte
}

// Storage is the interface which provides storage for Cache.
// Uses may use any key-value storage to implement this.
type Storage interface {
	MGet(ctx context.Context, keys ...string) ([][]byte, error)
	MSet(ctx context.Context, kvPairs []KVPair, expiration time.Duration) error
	MDelete(ctx context.Context, keys ...string) error
}

// CacheConfig is used to configure a Cache instance.
type CacheConfig[K comparable, V Model] struct {

	// Storage must return a Storage implementation which will be used
	// as the underlying key-value storage.
	Storage func(ctx context.Context) Storage

	// IdFunc returns the primary key of a Model object.
	IdFunc func(V) K

	// KeyFunc specifies the key function to use with the storage.
	KeyFunc Key

	// MGetBatchSize optionally specifies the batch size for one MGet
	// call to storage. The default is 100.
	MGetBatchSize int

	// MSetBatchSize optionally specifies the batch size for one MSet
	// call to storage. The default is 100.
	MSetBatchSize int

	// MGetBatchSize optionally specifies the batch size for one MDelete
	// call to storage. The default is 100.
	MDeleteBatchSize int

	// LRUCache optionally enables LRU cache, which may help to improve
	// the performance for high concurrency use-case.
	LRUCache *lru.Cache[K, V]

	// LRUExpiration specifies the expiration time for data in LRU cache.
	// The default is one second.
	LRUExpiration time.Duration
}

func (p *CacheConfig[_, V]) buildNewElemFunc() func() V {
	var x V
	typ := rtype.RTypeOf(x)
	if typ.Kind() == reflect.Ptr {
		valTyp := typ.Elem()
		return func() V {
			elem := linkname.Reflect_unsafe_New(unsafe.Pointer(valTyp))
			return typ.PackInterface(elem).(V)
		}
	}
	return func() V {
		elem := new(V)
		return *elem
	}
}

// NewCache returns a new Cache instance.
func NewCache[K comparable, V Model](config *CacheConfig[K, V]) *Cache[K, V] {
	if config.MGetBatchSize <= 0 {
		config.MGetBatchSize = DefaultBatchSize
	}
	if config.MSetBatchSize <= 0 {
		config.MSetBatchSize = DefaultBatchSize
	}
	if config.MDeleteBatchSize <= 0 {
		config.MDeleteBatchSize = DefaultBatchSize
	}
	if config.LRUExpiration <= 0 {
		config.LRUExpiration = DefaultLRUExpiration
	}
	newElemFn := config.buildNewElemFunc()
	return &Cache[K, V]{
		config:      config,
		newElemFunc: newElemFn,
	}
}

// Cache encapsulates frequently used batching cache operations,
// such as MGet, MSet and MDelete.
//
// A Cache must not be copied after initialized.
type Cache[K comparable, V Model] struct {
	config *CacheConfig[K, V]

	newElemFunc func() V
}

// MGetSlice queries cache and returns the cached values as a slice
// of type []V.
func (p *Cache[K, V]) MGetSlice(ctx context.Context, pks []K) ([]V, error) {
	if len(pks) == 0 {
		return nil, nil
	}

	// pk 去重
	pks = easy.Unique(pks, false)

	out := make([]V, 0, len(pks))
	err := p.mget(ctx, pks, func(pk K, elem V) {
		out = append(out, elem)
	})
	return out, err
}

// MGetMap queries cache and returns the cached values as a map
// of type map[K]V.
func (p *Cache[K, V]) MGetMap(ctx context.Context, pks []K) (map[K]V, error) {
	if len(pks) == 0 {
		return nil, nil
	}

	// pk 去重
	pks = easy.Unique(pks, false)

	out := make(map[K]V, len(pks))
	err := p.mget(ctx, pks, func(pk K, elem V) {
		out[pk] = elem
	})
	return out, err
}

func (p *Cache[K, V]) mget(ctx context.Context, pks []K, f func(pk K, elem V)) error {
	lruMissingKeys := make([]string, 0, len(pks))
	for _, pk := range pks {
		if p.config.LRUCache != nil {
			val, exists := p.config.LRUCache.GetNotStale(pk)
			if exists {
				f(pk, val)
				continue
			}
		}
		key := p.config.KeyFunc(pk)
		lruMissingKeys = append(lruMissingKeys, key)
	}

	stor := p.config.Storage(ctx)

	var err error
	var batchValues [][]byte
	var lruMissingKeyValues map[K]V
	if p.config.LRUCache != nil {
		lruMissingKeyValues = make(map[K]V, len(lruMissingKeys))
	}
	var batchKeys = easy.Split(lruMissingKeys, p.config.MGetBatchSize)
	for _, bat := range batchKeys {
		batchValues, err = stor.MGet(ctx, bat...)
		if err != nil {
			return err
		}
		for _, val := range batchValues {
			if len(val) == 0 {
				continue
			}
			elem := p.newElemFunc()
			err = elem.UnmarshalBinary(val)
			if err != nil {
				return err
			}
			pk := p.config.IdFunc(elem)
			if lruMissingKeyValues != nil {
				lruMissingKeyValues[pk] = elem
			}
			f(pk, elem)
		}
	}
	if len(lruMissingKeyValues) > 0 {
		p.config.LRUCache.MSet(lruMissingKeyValues, p.config.LRUExpiration)
	}

	return nil
}

// MSetSlice writes the given models to Cache.
func (p *Cache[K, V]) MSetSlice(ctx context.Context, models []V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}

	var kvPairs = make([]KVPair, 0, len(models))
	var kvMap = make(map[K]V, len(models))
	for _, elem := range models {
		buf, err := elem.MarshalBinary()
		if err != nil {
			return err
		}
		pk := p.config.IdFunc(elem)
		key := p.config.KeyFunc(pk)
		kvPairs = append(kvPairs, KVPair{key, buf})
		kvMap[pk] = elem
	}
	return p.mset(ctx, kvPairs, kvMap, expiration)
}

// MSetMap writes the given models to Cache.
func (p *Cache[K, V]) MSetMap(ctx context.Context, models map[K]V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}

	var kvPairs = make([]KVPair, 0, len(models))
	for pk, elem := range models {
		buf, err := elem.MarshalBinary()
		if err != nil {
			return err
		}
		key := p.config.KeyFunc(pk)
		kvPairs = append(kvPairs, KVPair{key, buf})
	}
	return p.mset(ctx, kvPairs, models, expiration)
}

func (p *Cache[K, V]) mset(ctx context.Context, kvPairs []KVPair, kvMap map[K]V, expiration time.Duration) error {
	stor := p.config.Storage(ctx)
	batches := easy.Split(kvPairs, p.config.MSetBatchSize)
	for _, bat := range batches {
		err := stor.MSet(ctx, bat, expiration)
		if err != nil {
			return err
		}
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.MSet(kvMap, p.config.LRUExpiration)
	}
	return nil
}

// MDelete deletes multiple values from Cache.
func (p *Cache[K, V]) MDelete(ctx context.Context, pks []K) error {
	if len(pks) == 0 {
		return nil
	}

	keys := make([]string, 0, len(pks))
	for _, pk := range pks {
		key := p.config.KeyFunc(pk)
		keys = append(keys, key)
	}

	stor := p.config.Storage(ctx)
	batches := easy.Split(keys, p.config.MDeleteBatchSize)
	for _, bat := range batches {
		err := stor.MDelete(ctx, bat...)
		if err != nil {
			return err
		}
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.MDelete(pks...)
	}
	return nil
}
