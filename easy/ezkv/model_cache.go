package ezkv

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/collection/set"
	"github.com/jxskiss/gopkg/v2/easy"
	"github.com/jxskiss/gopkg/v2/internal"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/perf/lru"
	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
	"github.com/jxskiss/gopkg/v2/utils/compress"
)

const (
	// DefaultBatchSize is the default batch size for batch operations.
	DefaultBatchSize = 100

	// DefaultLoaderBatchSize is the default batch size for calling Loader.
	DefaultLoaderBatchSize = 300

	// DefaultLRUExpiration is the default expiration time for data in LRU cache.
	DefaultLRUExpiration = 5 * time.Second
)

var ErrDataNotFound = errors.New("data not found")

type setCacheError struct {
	error
}

func (e *setCacheError) Error() string { return "set cache data: " + e.error.Error() }
func (e *setCacheError) Unwrap() error { return e.error }

// IsSetCacheError tells whether an error returned from this package is
// occurred when set data to underlying cache storage.
func IsSetCacheError(err error) bool {
	return errors.As(err, new(*setCacheError))
}

// Model is the interface implemented by types that can be cached by ModelCache.
type Model interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// Storage is the interface which provides storage for ModelCache.
// Users may use any key-value storage to implement this.
type Storage interface {
	Get(ctx context.Context, keys ...string) ([][]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	BatchSet(ctx context.Context, keys []string, values [][]byte, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

// Loader loads data from underlying persistent storage.
type Loader[K comparable, V Model] func(ctx context.Context, pks []K) (map[K]V, error)

// ModelCacheConfig configures a ModelCache instance.
type ModelCacheConfig[K comparable, V Model] struct {

	// BizName is used to identify the cache instance.
	BizName string

	// Storage must return a Storage implementation which will be used
	// as the underlying key-value storage.
	Storage func(ctx context.Context) Storage

	// IDFunc returns the primary key of a Model object.
	IDFunc func(V) K

	// KeyFunc specifies the key function to use with the storage.
	KeyFunc func(pk K) string

	// BatchGetSize optionally specifies the batch size for one BatchGet
	// calling to storage. The default is 200.
	BatchGetSize int

	// BatchSetSize optionally specifies the batch size for one Set
	// calling to storage. The default is 200.
	BatchSetSize int

	// BatchDeleteSize optionally specifies the batch size for one Delete
	// calling to storage. The default is 200.
	BatchDeleteSize int

	// LRUCache optionally enables LRU cache, which may help to improve
	// the performance for high concurrency use-case.
	LRUCache lru.Interface[K, V]

	// LRUExpiration specifies the expiration time for data in LRU cache.
	// The default is one second.
	LRUExpiration time.Duration

	// Loader optionally specifies a function to load data from underlying
	// persistent storage when the data is missing from cache.
	//
	// If Loader is not nil, all errors happened when processing cache
	// are treat as cache-miss and ignored, Loader will be called to load data
	// from underlying persistent storage.
	// If you want to fine-grained control the error handling,
	// config this to nil and handle cache errors by yourself.
	Loader Loader[K, V]

	// LoaderBatchSize optionally specifies the batch size for calling
	// Loader. The default is 500.
	LoaderBatchSize int

	// CacheExpiration specifies the expiration time to cache the data to
	// Storage, when Loader is configured and data are loaded by Loader.
	// The default is zero, which means no expiration.
	CacheExpiration time.Duration

	// CacheLoaderResultAsync makes the ModelCache to save data from Loader
	// to Storage async, errors returned from Storage.MSet will be ignored.
	// The default is false, it reports errors to the caller.
	CacheLoaderResultAsync bool

	// Compressor optionally specifies a compressor to use for
	// compressing and decompressing cached data.
	// The default is nil, which means no compression.
	Compressor compress.Compressor

	// ErrorLogger optionally specifies a function to log ignored errors.
	ErrorLogger func(ctx context.Context, err error, msg string)
}

func (p *ModelCacheConfig[_, _]) checkAndSetDefaults() {
	if p.BatchGetSize <= 0 {
		p.BatchGetSize = DefaultBatchSize
	}
	if p.BatchSetSize <= 0 {
		p.BatchSetSize = DefaultBatchSize
	}
	if p.BatchDeleteSize <= 0 {
		p.BatchDeleteSize = DefaultBatchSize
	}
	if p.LRUExpiration <= 0 {
		p.LRUExpiration = DefaultLRUExpiration
	}
	if p.LoaderBatchSize <= 0 {
		p.LoaderBatchSize = DefaultLoaderBatchSize
	}
	if p.ErrorLogger == nil {
		p.ErrorLogger = internal.DefaultLoggerError
	}
}

func buildNewElemFunc[V any]() func() V {
	var x V
	typ := reflectx.RTypeOf(x)
	if typ.Kind() == reflect.Ptr {
		valTyp := typ.Elem()
		return func() V {
			ptr := linkname.Reflect_unsafe_New(unsafe.Pointer(valTyp))
			return *(*V)(unsafe.Pointer(&ptr))
		}
	}
	return func() (value V) { return }
}

// NewModelCache returns a new ModelCache instance.
func NewModelCache[K comparable, V Model](config *ModelCacheConfig[K, V]) *ModelCache[K, V] {
	config.checkAndSetDefaults()
	newElemFn := buildNewElemFunc[V]()
	return &ModelCache[K, V]{
		config:      config,
		newElemFunc: newElemFn,
	}
}

// ModelCache encapsulates frequently used batching cache operations,
// such as MGet, MSet and Delete.
//
// A ModelCache must not be copied after initialized.
type ModelCache[K comparable, V Model] struct {
	config *ModelCacheConfig[K, V]

	newElemFunc func() V
}

// Get queries ModelCache for a given pk.
//
// If pk cannot be found either in the cache nor from the Loader,
// it returns an error ErrDataNotFound.
//
// Error may occur during setting data to cache, while we do get data
// from Loader, in this case the returned value is valid, but the error
// is returned together with the value, user can use IsSetCacheError
// to check it.
func (p *ModelCache[K, V]) Get(ctx context.Context, pk K) (V, error) {
	if p.config.LRUCache != nil {
		val, exists := p.config.LRUCache.GetNotStale(pk)
		if exists {
			return val, nil
		}
	}

	var zeroVal V
	stor := p.config.Storage(ctx)
	key := p.config.KeyFunc(pk)
	cacheResult, err := stor.Get(ctx, key)
	if err != nil && p.config.Loader == nil {
		return zeroVal, fmt.Errorf("query storage: %w", err)
	}
	if len(cacheResult) > 0 && len(cacheResult[0]) > 0 {
		val := cacheResult[0]
		elem, success, err1 := p.decodeCacheValue(ctx, pk, val)
		if err1 != nil {
			return zeroVal, err1
		}
		if success {
			if p.config.LRUCache != nil {
				p.config.LRUCache.Set(pk, elem, p.config.LRUExpiration)
			}
			return elem, nil
		}
	}
	if p.config.Loader != nil {
		var loaderResult map[K]V
		loaderResult, err = p.config.Loader(ctx, []K{pk})
		if err != nil {
			return zeroVal, fmt.Errorf("execute loader: %w", err)
		}
		elem, exists := loaderResult[pk]
		if exists {
			if p.config.CacheLoaderResultAsync {
				go func() {
					defer func() { recover() }()
					err1 := p.Set(ctx, pk, elem, p.config.CacheExpiration)
					if err1 != nil {
						p.config.ErrorLogger(ctx, err1, fmt.Sprintf("[ModelCache] set cache failed, bizName= %s, pk= %v", p.config.BizName, pk))
					}
				}()
			} else {
				err = p.Set(ctx, pk, elem, p.config.CacheExpiration)
				if err != nil {
					return elem, &setCacheError{err}
				}
			}
			return elem, nil
		}
	}
	return zeroVal, ErrDataNotFound
}

func (p *ModelCache[K, V]) decodeCacheValue(ctx context.Context, pk K, cacheVal []byte) (elem V, success bool, err error) {
	var zeroVal V
	val := cacheVal
	if p.config.Compressor != nil {
		val, _, err = p.config.Compressor.Decompress(ctx, val)
		if err != nil {
			if p.config.Loader == nil {
				return zeroVal, false, err
			}
			// treat as cache-miss, log and go ahead
			p.config.ErrorLogger(ctx, err, fmt.Sprintf("[ModelCache] decompress failed, bizName= %s, pk= %v", p.config.BizName, pk))
			return zeroVal, false, nil
		}
	}
	elem = p.newElemFunc()
	err = elem.UnmarshalBinary(val)
	if err != nil {
		if p.config.Loader == nil {
			return zeroVal, false, fmt.Errorf("unmarshal model: %w", err)
		}
		// treat as cache-miss, log and go ahead
		p.config.ErrorLogger(ctx, err, fmt.Sprintf("[ModelCache] unmarshal failed, bizName= %s, pk= %v", p.config.BizName, pk))
		return zeroVal, false, nil
	}
	return elem, true, nil
}

// BatchGetSlice queries ModelCache and returns the cached values as a slice
// of type []V.
// Note the returned values may be less than requested.
//
// Error may occur during setting data to cache, while we do get data
// from Loader, in this case the returned value is valid, but the error
// is returned together with the value, user can use IsSetCacheError
// to check it.
// But for any error that IsSetCacheError returns false, the returned
// data is incomplete and shall not be used.
func (p *ModelCache[K, V]) BatchGetSlice(ctx context.Context, pks []K) ([]V, error) {
	if len(pks) == 0 {
		return nil, nil
	}

	// pk 去重
	pks = easy.Unique(pks, false)

	out := make([]V, 0, len(pks))
	valfunc := func(_ K, elem V) {
		out = append(out, elem)
	}
	err := p.mGet(ctx, pks, valfunc)
	return out, err
}

// BatchGetMap queries ModelCache and returns the cached values as a map
// of type map[K]V.
// Note the returned values may be less than requested.
//
// Error may occur during setting data to cache, while we do get data
// from Loader, in this case the returned value is valid, but the error
// is returned together with the value, user can use IsSetCacheError
// to check it.
// But for any error that IsSetCacheError returns false, the returned
// data is incomplete and shall not be used.
func (p *ModelCache[K, V]) BatchGetMap(ctx context.Context, pks []K) (map[K]V, error) {
	if len(pks) == 0 {
		return nil, nil
	}

	// pk 去重
	pks = easy.Unique(pks, false)

	out := make(map[K]V, len(pks))
	valfunc := func(pk K, elem V) {
		out[pk] = elem
	}
	err := p.mGet(ctx, pks, valfunc)
	return out, err
}

func (p *ModelCache[K, V]) mGet(ctx context.Context, pks []K, f func(pk K, elem V)) error {
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

	compressor := p.config.Compressor
	stor := p.config.Storage(ctx)

	var err error
	var batchValues [][]byte
	var fromCache map[K]V
	if p.config.LRUCache != nil {
		fromCache = make(map[K]V, len(lruMissingKeys))
	}
	var cachedPKs = set.NewWithSize[K](len(lruMissingKeys))
	var batchKeys = easy.Split(lruMissingKeys, p.config.BatchGetSize)
	for _, bat := range batchKeys {
		batchValues, err = stor.Get(ctx, bat...)
		if err != nil {
			if p.config.Loader == nil {
				return fmt.Errorf("query storage: %w", err)
			}
			// treat as cache-miss, log and continue
			p.config.ErrorLogger(ctx, err, fmt.Sprintf("[ModelCache] query storage failed, bizName= %s, keys= %v", p.config.BizName, bat))
			continue
		}
		for i, val := range batchValues {
			if len(val) == 0 {
				continue
			}
			if compressor != nil {
				val, _, err = compressor.Decompress(ctx, val)
				if err != nil {
					p.config.ErrorLogger(ctx, err, fmt.Sprintf("[ModelCache] decompress failed, bizName= %s, key= %s", p.config.BizName, bat[i]))
					continue
				}
			}
			elem := p.newElemFunc()
			err = elem.UnmarshalBinary(val)
			if err != nil {
				p.config.ErrorLogger(ctx, err, fmt.Sprintf("[ModelCache] unmarshal failed, bizName= %s, key= %s", p.config.BizName, bat[i]))
				continue
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
				return fmt.Errorf("execute loader: %w", err)
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
				defer func() { recover() }()
				err1 := p.BatchSetMap(ctx, fromLoader, p.config.CacheExpiration)
				if err1 != nil {
					p.config.ErrorLogger(ctx, err1, fmt.Sprintf("[ModelCache] batch set cache failed, bizName= %s", p.config.BizName))
				}
			}()
		} else {
			err = p.BatchSetMap(ctx, fromLoader, p.config.CacheExpiration)
			if err != nil {
				return &setCacheError{err}
			}
		}
	}

	return nil
}

// Set writes a key value pair to ModelCache.
func (p *ModelCache[K, V]) Set(ctx context.Context, pk K, elem V, expiration time.Duration) error {
	key := p.config.KeyFunc(pk)
	buf, err := elem.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal model: %w", err)
	}
	if p.config.Compressor != nil {
		buf, _, _ = p.config.Compressor.Compress(ctx, buf)
	}
	stor := p.config.Storage(ctx)
	err = stor.Set(ctx, key, buf, expiration)
	if err != nil {
		return fmt.Errorf("write storage: %w", err)
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.Set(pk, elem, p.config.LRUExpiration)
	}
	return nil
}

// BatchSetSlice writes the given models to ModelCache.
func (p *ModelCache[K, V]) BatchSetSlice(ctx context.Context, models []V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}

	compressor := p.config.Compressor
	stor := p.config.Storage(ctx)
	batchSize := min(p.config.BatchSetSize, len(models))
	keys := make([]string, 0, batchSize)
	values := make([][]byte, 0, batchSize)
	for _, batchModels := range easy.Split(models, batchSize) {
		keys = keys[:0]
		values = values[:0]
		var kvMap map[K]V
		if p.config.LRUCache != nil {
			kvMap = make(map[K]V, len(batchModels))
		}
		for _, elem := range batchModels {
			buf, err := elem.MarshalBinary()
			if err != nil {
				return fmt.Errorf("marshal model: %w", err)
			}
			if compressor != nil {
				buf, _, _ = compressor.Compress(ctx, buf)
			}
			pk := p.config.IDFunc(elem)
			key := p.config.KeyFunc(pk)
			keys = append(keys, key)
			values = append(values, buf)
			if p.config.LRUCache != nil {
				kvMap[pk] = elem
			}
		}
		err := stor.BatchSet(ctx, keys, values, expiration)
		if err != nil {
			return fmt.Errorf("write storage: %w", err)
		}
		if p.config.LRUCache != nil {
			p.config.LRUCache.MSet(kvMap, p.config.LRUExpiration)
		}
	}
	return nil
}

// BatchSetMap writes the given models to ModelCache.
func (p *ModelCache[K, V]) BatchSetMap(ctx context.Context, models map[K]V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}

	compressor := p.config.Compressor
	stor := p.config.Storage(ctx)
	batchSize := min(p.config.BatchSetSize, len(models))
	keys := make([]string, 0, batchSize)
	values := make([][]byte, 0, batchSize)
	for pk, elem := range models {
		buf, err := elem.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal model: %w", err)
		}
		if compressor != nil {
			buf, _, _ = compressor.Compress(ctx, buf)
		}
		key := p.config.KeyFunc(pk)
		keys = append(keys, key)
		values = append(values, buf)
		if len(keys) == batchSize {
			err = stor.BatchSet(ctx, keys, values, expiration)
			if err != nil {
				return fmt.Errorf("write storage: %w", err)
			}
			keys = keys[:0]
			values = values[:0]
		}
	}
	if len(keys) > 0 {
		err := stor.BatchSet(ctx, keys, values, expiration)
		if err != nil {
			return fmt.Errorf("write storage: %w", err)
		}
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.MSet(models, p.config.LRUExpiration)
	}
	return nil
}

// Delete deletes key values from ModelCache.
func (p *ModelCache[K, V]) Delete(ctx context.Context, pks ...K) error {
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
		return fmt.Errorf("write storage: %w", err)
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.Delete(pk)
	}
	return nil
}

// mDelete deletes multiple key values from ModelCache.
func (p *ModelCache[K, V]) mDelete(ctx context.Context, pks []K) error {
	if len(pks) == 0 {
		return nil
	}

	keys := make([]string, 0, len(pks))
	for _, pk := range pks {
		key := p.config.KeyFunc(pk)
		keys = append(keys, key)
	}

	stor := p.config.Storage(ctx)
	batches := easy.Split(keys, p.config.BatchDeleteSize)
	for _, bat := range batches {
		err := stor.Delete(ctx, bat...)
		if err != nil {
			return fmt.Errorf("write storage: %w", err)
		}
	}
	if p.config.LRUCache != nil {
		p.config.LRUCache.MDelete(pks...)
	}
	return nil
}
