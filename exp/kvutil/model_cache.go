package kvutil

import (
	"context"
	"encoding"
	"errors"
	"reflect"
	"time"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/collection/set"
	"github.com/jxskiss/gopkg/v2/easy"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/perf/lru"
	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

// DefaultBatchSize is the default batch size for batch operations.
const DefaultBatchSize = 200

// DefaultLoaderBatchSize is the default batch size for calling Loader.
const DefaultLoaderBatchSize = 500

// DefaultLRUExpiration is the default expiration time for data in LRU cache.
const DefaultLRUExpiration = time.Second

var ErrDataNotFound = errors.New("data not found")

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
// Users may use any key-value storage to implement this.
type Storage interface {
	MGet(ctx context.Context, keys ...string) ([][]byte, error)
	MSet(ctx context.Context, kvPairs []KVPair, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

// Loader loads data from underlying persistent storage.
type Loader[K comparable, V Model] func(ctx context.Context, pks []K) (map[K]V, error)

// CacheConfig configures a Cache instance.
type CacheConfig[K comparable, V Model] struct {

	// Storage must return a Storage implementation which will be used
	// as the underlying key-value storage.
	Storage func(ctx context.Context) Storage

	// IDFunc returns the primary key of a Model object.
	IDFunc func(V) K

	// KeyFunc specifies the key function to use with the storage.
	KeyFunc Key

	// MGetBatchSize optionally specifies the batch size for one MGet
	// calling to storage. The default is 200.
	MGetBatchSize int

	// MSetBatchSize optionally specifies the batch size for one MSet
	// calling to storage. The default is 200.
	MSetBatchSize int

	// DeleteBatchSize optionally specifies the batch size for one Delete
	// calling to storage. The default is 200.
	DeleteBatchSize int

	// LRUCache optionally enables LRU cache, which may help to improve
	// the performance for high concurrency use-case.
	LRUCache lru.Interface[K, V]

	// LRUExpiration specifies the expiration time for data in LRU cache.
	// The default is one second.
	LRUExpiration time.Duration

	// Loader optionally specifies a function to load data from underlying
	// persistent storage when the data is missing from cache.
	Loader Loader[K, V]

	// LoaderBatchSize optionally specifies the batch size for calling
	// Loader. The default is 500.
	LoaderBatchSize int

	// CacheExpiration specifies the expiration time to cache the data to
	// Storage, when Loader is configured and data are loaded by Loader.
	// The default is zero, which means no expiration.
	CacheExpiration time.Duration

	// CacheLoaderResultAsync makes the Cache to save data from Loader
	// to Storage async, errors returned from Storage.MSet will be ignored.
	// The default is false, it reports errors to the caller.
	CacheLoaderResultAsync bool
}

func (p *CacheConfig[_, _]) checkAndSetDefaults() {
	if p.MGetBatchSize <= 0 {
		p.MGetBatchSize = DefaultBatchSize
	}
	if p.MSetBatchSize <= 0 {
		p.MSetBatchSize = DefaultBatchSize
	}
	if p.DeleteBatchSize <= 0 {
		p.DeleteBatchSize = DefaultBatchSize
	}
	if p.LRUExpiration <= 0 {
		p.LRUExpiration = DefaultLRUExpiration
	}
	if p.LoaderBatchSize <= 0 {
		p.LoaderBatchSize = DefaultLoaderBatchSize
	}
}

func buildNewElemFunc[V any]() func() V {
	var x V
	typ := reflectx.RTypeOf(x)
	if typ.Kind() == reflect.Ptr {
		valTyp := typ.Elem()
		return func() V {
			elem := linkname.Reflect_unsafe_New(unsafe.Pointer(valTyp))
			return typ.PackInterface(elem).(V)
		}
	}
	return func() V {
		return *new(V)
	}
}

// NewCache returns a new Cache instance.
func NewCache[K comparable, V Model](config *CacheConfig[K, V]) *Cache[K, V] {
	config.checkAndSetDefaults()
	newElemFn := buildNewElemFunc[V]()
	return &Cache[K, V]{
		config:      config,
		newElemFunc: newElemFn,
	}
}

// Cache encapsulates frequently used batching cache operations,
// such as MGet, MSet and Delete.
//
// A Cache must not be copied after initialized.
type Cache[K comparable, V Model] struct {
	config *CacheConfig[K, V]

	newElemFunc func() V
}

// Get queries Cache for a given pk.
//
// If pk cannot be found either in the cache nor from the Loader,
// it returns an error ErrDataNotFound.
func (p *Cache[K, V]) Get(ctx context.Context, pk K) (V, error) {
	if p.config.LRUCache != nil {
		val, exists := p.config.LRUCache.GetNotStale(pk)
		if exists {
			return val, nil
		}
	}

	var zeroVal V
	stor := p.config.Storage(ctx)
	key := p.config.KeyFunc(pk)
	cacheResult, err := stor.MGet(ctx, key)
	if err != nil && p.config.Loader == nil {
		return zeroVal, err
	}
	if len(cacheResult) > 0 && len(cacheResult[0]) > 0 {
		elem := p.newElemFunc()
		err = elem.UnmarshalBinary(cacheResult[0])
		if err != nil {
			return zeroVal, err
		}
		if p.config.LRUCache != nil {
			p.config.LRUCache.Set(pk, elem, p.config.LRUExpiration)
		}
		return elem, nil
	}
	if p.config.Loader != nil {
		var loaderResult map[K]V
		loaderResult, err = p.config.Loader(ctx, []K{pk})
		if err != nil {
			return zeroVal, err
		}
		elem, exists := loaderResult[pk]
		if exists {
			if p.config.CacheLoaderResultAsync {
				go func() {
					_ = p.Set(ctx, pk, elem, p.config.CacheExpiration)
				}()
			} else {
				err = p.Set(ctx, pk, elem, p.config.CacheExpiration)
				if err != nil {
					return zeroVal, err
				}
			}
			return elem, nil
		}
	}
	return zeroVal, ErrDataNotFound
}

// MGetSlice queries Cache and returns the cached values as a slice
// of type []V.
func (p *Cache[K, V]) MGetSlice(ctx context.Context, pks []K) ([]V, error) {
	if len(pks) == 0 {
		return nil, nil
	}

	// pk 去重
	pks = easy.Unique(pks, false)

	out := make([]V, 0, len(pks))
	valfunc := func(_ K, elem V) {
		out = append(out, elem)
	}
	err := p.mget(ctx, pks, valfunc)
	return out, err
}

// MGetMap queries Cache and returns the cached values as a map
// of type map[K]V.
func (p *Cache[K, V]) MGetMap(ctx context.Context, pks []K) (map[K]V, error) {
	if len(pks) == 0 {
		return nil, nil
	}

	// pk 去重
	pks = easy.Unique(pks, false)

	out := make(map[K]V, len(pks))
	valfunc := func(pk K, elem V) {
		out[pk] = elem
	}
	err := p.mget(ctx, pks, valfunc)
	return out, err
}

func (p *Cache[K, V]) mget(ctx context.Context, pks []K, f func(pk K, elem V)) error {
	var lruMissingPKs []K
	var lruMissingKeys []string
	if p.config.LRUCache != nil {
		lruResult := p.config.LRUCache.MGetNotStale(pks...)
		lruMissingKeys = make([]string, 0, len(pks)-len(lruResult))
		for _, pk := range pks {
			if elem, ok := lruResult[pk]; ok {
				f(pk, elem)
			} else {
				key := p.config.KeyFunc(pk)
				lruMissingPKs = append(lruMissingPKs, pk)
				lruMissingKeys = append(lruMissingKeys, key)
			}
		}
	} else {
		lruMissingPKs = pks
		lruMissingKeys = make([]string, len(pks))
		for i, pk := range pks {
			key := p.config.KeyFunc(pk)
			lruMissingKeys[i] = key
		}
	}

	stor := p.config.Storage(ctx)

	var err error
	var batchValues [][]byte
	var fromCache map[K]V
	if p.config.LRUCache != nil {
		fromCache = make(map[K]V, len(lruMissingKeys))
	}
	var cachedPKs = set.NewWithSize[K](len(lruMissingKeys))
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
			pk := p.config.IDFunc(elem)
			if fromCache != nil {
				fromCache[pk] = elem
			}
			f(pk, elem)
			cachedPKs.Add(pk)
		}
	}

	// load from underlying persistent storage if configured
	var fromLoader map[K]V
	if p.config.Loader != nil && cachedPKs.Size() < len(lruMissingPKs) {
		cacheMissingPKs := cachedPKs.FilterNotContains(lruMissingPKs)
		fromLoader = make(map[K]V, len(cacheMissingPKs))
		batchPKs := easy.Split(cacheMissingPKs, p.config.LoaderBatchSize)
		for _, bat := range batchPKs {
			batchResult, err := p.config.Loader(ctx, bat)
			if err != nil {
				return err
			}
			for pk, elem := range batchResult {
				f(pk, elem)
				fromLoader[pk] = elem
			}
		}
	}

	if len(fromCache) > 0 {
		p.config.LRUCache.MSet(fromCache, p.config.LRUExpiration)
	}
	if len(fromLoader) > 0 {
		if p.config.CacheLoaderResultAsync {
			go func() {
				_ = p.MSetMap(ctx, fromLoader, p.config.CacheExpiration)
			}()
		} else {
			err = p.MSetMap(ctx, fromLoader, p.config.CacheExpiration)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Set writes a key value pair to Cache.
func (p *Cache[K, V]) Set(ctx context.Context, pk K, elem V, expiration time.Duration) error {
	key := p.config.KeyFunc(pk)
	buf, err := elem.MarshalBinary()
	if err != nil {
		return err
	}
	kvPairs := []KVPair{{K: key, V: buf}}
	stor := p.config.Storage(ctx)
	err = stor.MSet(ctx, kvPairs, expiration)
	if err != nil {
		return err
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.Set(pk, elem, p.config.LRUExpiration)
	}
	return nil
}

// MSetSlice writes the given models to Cache.
func (p *Cache[K, V]) MSetSlice(ctx context.Context, models []V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}

	stor := p.config.Storage(ctx)
	batchSize := min(p.config.MSetBatchSize, len(models))
	kvPairs := make([]KVPair, 0, batchSize)
	for _, batchModels := range easy.Split(models, batchSize) {
		kvPairs = kvPairs[:0]
		var kvMap map[K]V
		if p.config.LRUCache != nil {
			kvMap = make(map[K]V, len(batchModels))
		}
		for _, elem := range batchModels {
			buf, err := elem.MarshalBinary()
			if err != nil {
				return err
			}
			pk := p.config.IDFunc(elem)
			key := p.config.KeyFunc(pk)
			kvPairs = append(kvPairs, KVPair{key, buf})
			if p.config.LRUCache != nil {
				kvMap[pk] = elem
			}
		}
		err := stor.MSet(ctx, kvPairs, expiration)
		if err != nil {
			return err
		}
		if p.config.LRUCache != nil {
			p.config.LRUCache.MSet(kvMap, p.config.LRUExpiration)
		}
	}
	return nil
}

// MSetMap writes the given models to Cache.
func (p *Cache[K, V]) MSetMap(ctx context.Context, models map[K]V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}

	stor := p.config.Storage(ctx)
	batchSize := min(p.config.MSetBatchSize, len(models))
	kvPairs := make([]KVPair, 0, batchSize)
	for pk, elem := range models {
		buf, err := elem.MarshalBinary()
		if err != nil {
			return err
		}
		key := p.config.KeyFunc(pk)
		kvPairs = append(kvPairs, KVPair{key, buf})
		if len(kvPairs) == batchSize {
			err = stor.MSet(ctx, kvPairs, expiration)
			if err != nil {
				return err
			}
			kvPairs = kvPairs[:0]
		}
	}
	if len(kvPairs) > 0 {
		err := stor.MSet(ctx, kvPairs, expiration)
		if err != nil {
			return err
		}
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.MSet(models, p.config.LRUExpiration)
	}
	return nil
}

// Delete deletes key values from Cache.
func (p *Cache[K, V]) Delete(ctx context.Context, pks ...K) error {
	if len(pks) == 0 {
		return nil
	}
	if len(pks) > 1 {
		return p.mDelete(ctx, pks)
	}

	pk := pks[0]
	key := p.config.KeyFunc(pk)
	stor := p.config.Storage(ctx)
	err := stor.Delete(ctx, key)
	if err != nil {
		return err
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.Delete(pk)
	}
	return nil
}

// mDelete deletes multiple key values from Cache.
func (p *Cache[K, V]) mDelete(ctx context.Context, pks []K) error {
	if len(pks) == 0 {
		return nil
	}

	keys := make([]string, 0, len(pks))
	for _, pk := range pks {
		key := p.config.KeyFunc(pk)
		keys = append(keys, key)
	}

	stor := p.config.Storage(ctx)
	batches := easy.Split(keys, p.config.DeleteBatchSize)
	for _, bat := range batches {
		err := stor.Delete(ctx, bat...)
		if err != nil {
			return err
		}
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.MDelete(pks...)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
