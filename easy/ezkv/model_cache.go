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
	"github.com/jxskiss/gopkg/v2/perf/lru"
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
// Model types must be concrete. For pointer types, UnmarshalBinary must work
// on a newly allocated zero-valued element; other types must support their zero value.
// MarshalBinary must be safe for concurrent reads of an immutable model.
type Model interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// Storage is the interface which provides storage for ModelCache.
// Users may use any key-value storage to implement this.
// Implementations must support concurrent calls when ModelCache is shared.
// Writes and deletes may partially succeed on error; ModelCache does not roll them back.
type Storage interface {
	// Get returns one value per key, in the same order, using nil or empty values
	// for misses. On error, it must return no values. Returned buffers must remain
	// valid and must not be modified by Storage after returning.
	Get(ctx context.Context, keys ...string) ([][]byte, error)
	// Set must consume or copy its inputs before returning, without modifying them.
	// A zero expiration means no expiration.
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	// BatchSet pairs keys[i] with values[i]; both slices have the same length.
	// It has the same buffer ownership and expiration requirements as Set.
	BatchSet(ctx context.Context, keys []string, values [][]byte, expiration time.Duration) error
	// Delete must treat missing keys as successfully deleted and consume or copy
	// the keys before returning.
	Delete(ctx context.Context, keys ...string) error
}

// Loader loads data from underlying persistent storage.
// It must return only requested keys, omit missing models, and never return nil
// models. Each map key must equal IDFunc(model), including after serialization.
// The returned map and models must not be modified by the Loader after returning.
// Loader must support concurrent calls when ModelCache is shared; calls for the
// same key are not coalesced. Results returned with an error are discarded.
type Loader[K comparable, V Model] func(ctx context.Context, pks []K) (map[K]V, error)

// ModelCacheConfig configures a ModelCache instance.
// NewModelCache fills defaults in this struct and retains its pointer.
// Do not mutate it while the cache is in use. Configured callbacks and components
// must support concurrent use when the cache is shared.
type ModelCacheConfig[K comparable, V Model] struct {

	// BizName is used to identify the cache instance.
	BizName string

	// Storage must return a Storage implementation which will be used
	// as the underlying key-value storage.
	// This function is required and must never return nil.
	Storage func(ctx context.Context) Storage

	// IDFunc returns the primary key of a Model object.
	// It is required and must return the same key before and after serialization.
	IDFunc func(V) K

	// KeyFunc specifies the key function to use with the storage.
	// It is required and must be stable and collision-free for distinct primary keys.
	KeyFunc func(pk K) string

	// BatchGetSize limits keys per Storage.Get call. Nonpositive values use
	// DefaultBatchSize (100).
	BatchGetSize int

	// BatchSetSize limits models per Storage.BatchSet call. Nonpositive values use
	// DefaultBatchSize (100).
	BatchSetSize int

	// BatchDeleteSize limits keys per Storage.Delete call. Nonpositive values use
	// DefaultBatchSize (100).
	BatchDeleteSize int

	// LRUCache optionally enables LRU cache, which may help to improve
	// the performance for high concurrency use-case.
	LRUCache lru.Interface[K, V]

	// LRUExpiration specifies the expiration time for data in LRU cache.
	// Nonpositive values use DefaultLRUExpiration (five seconds).
	// This TTL is independent of Storage expiration and may outlive it.
	LRUExpiration time.Duration

	// Loader optionally specifies a function to load data from underlying
	// persistent storage when the data is missing from cache.
	//
	// With a Loader, storage read and decoding errors are treated as cache misses.
	// Without a Loader, Get returns these errors, while BatchGetSlice and BatchGetMap
	// return storage read errors but log and skip values that fail decoding.
	// Loader errors are always returned. Synchronous writeback errors are returned
	// as IsSetCacheError with valid loaded data.
	Loader Loader[K, V]

	// LoaderBatchSize limits keys per Loader call. Nonpositive values use
	// DefaultLoaderBatchSize (300).
	LoaderBatchSize int

	// CacheExpiration specifies the expiration time to cache the data to
	// Storage, when Loader is configured and data are loaded by Loader.
	// The default is zero, which means no expiration.
	CacheExpiration time.Duration

	// CacheLoaderResultAsync writes Loader results in a background goroutine using
	// the request context; cancellation can prevent writeback. Writeback errors are
	// logged through ErrorLogger instead of returned to the caller.
	// The default is false, which waits for writeback and reports its errors.
	CacheLoaderResultAsync bool

	// Compressor optionally specifies a compressor to use for
	// compressing and decompressing cached data.
	// The default is nil, which means no compression.
	Compressor compress.Compressor

	// DisableCompressionWrites bypasses Compressor on writes, storing the raw
	// MarshalBinary result without escaping magic prefixes or framing empty data.
	// Reads still use Compressor, which must accept unframed data in this mode.
	DisableCompressionWrites bool

	// ErrorLogger optionally specifies a function to log ignored errors.
	ErrorLogger func(ctx context.Context, err error, msg string)
}

func (p *ModelCacheConfig[_, _]) checkAndSetDefaults() error {
	if p == nil {
		return errors.New("ezkv: config must not be nil")
	}
	if p.Storage == nil {
		return errors.New("ezkv: ModelCacheConfig.Storage must not be nil")
	}
	if p.IDFunc == nil {
		return errors.New("ezkv: ModelCacheConfig.IDFunc must not be nil")
	}
	if p.KeyFunc == nil {
		return errors.New("ezkv: ModelCacheConfig.KeyFunc must not be nil")
	}
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
	return nil
}

func buildNewElemFunc[V any]() func() V {
	var x V
	typ := reflect.TypeOf(x)
	if typ.Kind() == reflect.Pointer {
		valTyp := typ.Elem()
		return func() V {
			ptr := reflect.New(valTyp).UnsafePointer()
			return *(*V)(unsafe.Pointer(&ptr))
		}
	}
	return func() (value V) { return }
}

// NewModelCache returns a new ModelCache instance.
// It returns an error for invalid config, or an interface model type V.
// It does not call the configured functions. It fills defaults in config and
// retains it; config must not be changed while the cache is in use.
func NewModelCache[K comparable, V Model](config *ModelCacheConfig[K, V]) (*ModelCache[K, V], error) {
	if err := config.checkAndSetDefaults(); err != nil {
		return nil, err
	}
	if reflect.TypeOf((*V)(nil)).Elem().Kind() == reflect.Interface {
		return nil, errors.New("ezkv: model type V must be concrete, not an interface")
	}
	newElemFn := buildNewElemFunc[V]()
	return &ModelCache[K, V]{
		config:      config,
		newElemFunc: newElemFn,
	}, nil
}

// ModelCache encapsulates frequently used batching cache operations,
// such as MGet, MSet and Delete.
//
// A ModelCache must not be copied after initialized.
// Models returned by Get, BatchGetSlice and BatchGetMap are shared, read-only
// objects. Callers must not modify them, including referenced slices, maps and
// pointers; make a deep copy before making changes. Models passed to write
// methods or returned by Loader must also remain immutable once handed to the
// cache, since LRU entries and asynchronous writeback can retain them.
// Concurrent operations do not provide transactional consistency between Storage
// and LRU; cached data may remain stale until its LRU expiration.
type ModelCache[K comparable, V Model] struct {
	config *ModelCacheConfig[K, V]

	newElemFunc func() V
}

// Get queries ModelCache for a given pk.
//
// If pk cannot be found either in the cache nor from the Loader,
// it returns an error ErrDataNotFound.
// The returned model is read-only, as described by ModelCache.
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
// Duplicate keys are queried once.
// Missing values and values that fail decoding are omitted unless supplied by Loader.
// Decoding errors are logged and skipped even without a Loader.
// Results are read-only and have no guaranteed order.
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
// Duplicate keys are queried once.
// Missing values and values that fail decoding are omitted unless supplied by Loader.
// Decoding errors are logged and skipped even without a Loader.
// The models in the returned map are read-only.
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
// elem must be non-nil and immutable, and pk must equal IDFunc(elem).
// expiration applies to Storage; the LRU uses its independent LRUExpiration.
func (p *ModelCache[K, V]) Set(ctx context.Context, pk K, elem V, expiration time.Duration) error {
	key := p.config.KeyFunc(pk)
	buf, err := elem.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal model: %w", err)
	}
	if p.config.Compressor != nil && !p.config.DisableCompressionWrites {
		buf = p.config.Compressor.Compress(ctx, buf).Data
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
// Models must be non-nil and immutable. Each successful batch updates the LRU.
// On error, earlier successful batches remain applied; the failing batch may
// have partially changed Storage, but its LRU entries are not updated.
// expiration applies to Storage; the LRU uses its independent LRUExpiration.
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
			if compressor != nil && !p.config.DisableCompressionWrites {
				buf = compressor.Compress(ctx, buf).Data
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
// Each map key must equal IDFunc(model); models must be non-nil and immutable.
// Batches are processed in unspecified order. Each successful batch updates the
// LRU. On error, earlier successful batches remain applied; the failing batch may
// have partially changed Storage, but its LRU entries are not updated.
// expiration applies to Storage; the LRU uses its independent LRUExpiration.
func (p *ModelCache[K, V]) BatchSetMap(ctx context.Context, models map[K]V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}

	compressor := p.config.Compressor
	stor := p.config.Storage(ctx)
	batchSize := min(p.config.BatchSetSize, len(models))
	keys := make([]string, 0, batchSize)
	values := make([][]byte, 0, batchSize)
	var kvMap map[K]V
	for pk, elem := range models {
		buf, err := elem.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal model: %w", err)
		}
		if compressor != nil && !p.config.DisableCompressionWrites {
			buf = compressor.Compress(ctx, buf).Data
		}
		key := p.config.KeyFunc(pk)
		if p.config.LRUCache != nil {
			if len(keys) == 0 {
				kvMap = make(map[K]V, batchSize)
			}
			kvMap[pk] = elem
		}
		keys = append(keys, key)
		values = append(values, buf)
		if len(keys) == batchSize {
			err = stor.BatchSet(ctx, keys, values, expiration)
			if err != nil {
				return fmt.Errorf("write storage: %w", err)
			}
			if p.config.LRUCache != nil {
				p.config.LRUCache.MSet(kvMap, p.config.LRUExpiration)
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
		if p.config.LRUCache != nil {
			p.config.LRUCache.MSet(kvMap, p.config.LRUExpiration)
		}
	}
	return nil
}

// Delete deletes key values from ModelCache.
// Each successful batch removes its keys from the LRU. On error, earlier
// successful batches remain applied; the failing batch may have partially
// changed Storage, but its LRU entries are not removed.
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
	for i, bat := range batches {
		err := stor.Delete(ctx, bat...)
		if err != nil {
			return fmt.Errorf("write storage: %w", err)
		}
		if p.config.LRUCache != nil {
			start := i * p.config.BatchDeleteSize
			p.config.LRUCache.MDelete(pks[start : start+len(bat)]...)
		}
	}
	return nil
}
