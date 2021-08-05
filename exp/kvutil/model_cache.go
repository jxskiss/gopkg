package kvutil

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jxskiss/gopkg/easy"
	"github.com/jxskiss/gopkg/lru"
)

// DefaultBatchSize is the default batch size for MGet and MSet operations.
const DefaultBatchSize = 500

// Model is the interface implemented by types that can be cached by ModelCache.
type Model interface {
	MarshalModel() ([]byte, error)
	UnmarshalModel(b []byte) error
}

// KVPair represents a key value pair to set to ModelCache.
type KVPair struct {
	K string
	V []byte
}

// Storage is the interface which provides storage for ModelCache.
// Uses may use any key-value storage to implement this.
type Storage interface {
	MGet(ctx context.Context, keys ...string) ([][]byte, error)
	MSet(ctx context.Context, kvPairs []KVPair, expiration time.Duration) error
	MDelete(ctx context.Context, keys ...string) error
}

// ModelCache encapsulates frequently used data model cache operations,
// such as MGet, MSet and MDelete and batching.
type ModelCache struct {

	// ClientFunc must returns a Storage implementation which will be used
	// as the underlying key-value storage.
	ClientFunc func(ctx context.Context) Storage

	// IdFunc returns the primary key of a Model object.
	IdFunc func(interface{}) interface{}

	// KeyFunc specifies the key function to use with the storage.
	KeyFunc Key

	// MGetBatchSize optionally gives the batch size for one MGet call to
	// storage. The default is DefaultBatchSize.
	MGetBatchSize int

	// MSetBatchSize optionally gives the batch size for one MSet call to
	// storage. The default is DefaultBatchSize.
	MSetBatchSize int

	// Expiration specifies the expiration time to cache model data in storage.
	Expiration time.Duration

	// LruCache optionally enables LRU cache, which may help to improve
	// the performance for high concurrency use-case.
	LruCache *lru.Cache

	// LruExpiration specifies the expiration time for data in LRU cache.
	LruExpiration time.Duration
}

// MGetByIntKeys retrieve multiple model values from cache using int64 keys.
//
// dstPtr must be either *map[int64]<Model> or *[]<Model>,
// otherwise it returns an error.
func (p *ModelCache) MGetByIntKeys(ctx context.Context, dstPtr interface{}, pks []int64) (err error) {
	if len(pks) == 0 {
		return nil
	}
	defer recovery("MGetByIntKeys", &err)

	var isSlicePtr, isMapPtr bool
	isSlicePtr = isCacheableSlicePointer(dstPtr)
	if !isSlicePtr {
		isMapPtr = isIntCacheableMapPointer(dstPtr)
		if !isMapPtr {
			return errors.New("dst is not cacheable model slice/map pointer")
		}
	}

	client := p.ClientFunc(ctx)
	batchSize := p.MGetBatchSize
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	// pk 去重
	pks = easy.UniqueInt64s(pks, false)

	keys := make([]string, 0, len(pks))
	lruItems := make([]interface{}, 0)
	for _, pk := range pks {
		key := p.KeyFunc(pk)
		if p.LruCache != nil {
			lruVal, exists := p.LruCache.GetNotStale(key)
			if exists && lruVal != nil {
				lruItems = append(lruItems, lruVal)
				continue
			}
		}
		keys = append(keys, key)
	}

	var batchKeys []string
	var batchValues [][]byte
	var batches = easy.SplitBatch(len(keys), batchSize)

	values := make([][]byte, 0, len(keys))
	for _, idx := range batches {
		batchKeys = keys[idx.I:idx.J]
		batchValues, err = client.MGet(ctx, batchKeys...)
		if err != nil {
			return err
		}
		values = append(values, batchValues...)
	}

	if isSlicePtr {
		return p.unmarshalCacheableSlice(dstPtr, lruItems, values)
	}
	return p.unmarshalCacheableMap(dstPtr, lruItems, values)
}

// MGetByStringKeys retrieve multiple model values from cache using string keys.
//
// dstPtr must be either *map[string]<Model> or *[]<Model>,
// otherwise it returns an error.
func (p *ModelCache) MGetByStringKeys(ctx context.Context, dstPtr interface{}, pks []string) (err error) {
	if len(pks) == 0 {
		return nil
	}
	defer recovery("MGetByStringKeys", &err)

	var isSlicePtr, isMapPtr bool
	isSlicePtr = isCacheableSlicePointer(dstPtr)
	if !isSlicePtr {
		isMapPtr = isStringCacheableMapPointer(dstPtr)
		if !isMapPtr {
			return errors.New("dst is not cacheable model slice/map pointer")
		}
	}

	client := p.ClientFunc(ctx)
	batchSize := p.MGetBatchSize
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	// pk 去重
	pks = easy.UniqueStrings(pks, false)

	keys := make([]string, 0, len(pks))
	lruItems := make([]interface{}, 0)
	for _, pk := range pks {
		key := p.KeyFunc(pk)
		if p.LruCache != nil {
			lruVal, exists := p.LruCache.GetNotStale(key)
			if exists && lruVal != nil {
				lruItems = append(lruItems, lruVal)
				continue
			}
		}
		keys = append(keys, key)
	}

	var batchKeys []string
	var batchValues [][]byte
	var batches = easy.SplitBatch(len(keys), batchSize)

	values := make([][]byte, 0, len(keys))
	for _, idx := range batches {
		batchKeys = keys[idx.I:idx.J]
		batchValues, err = client.MGet(ctx, batchKeys...)
		if err != nil {
			return err
		}
		values = append(values, batchValues...)
	}

	if isSlicePtr {
		return p.unmarshalCacheableSlice(dstPtr, lruItems, values)
	}
	return p.unmarshalCacheableMap(dstPtr, lruItems, values)
}

// MSet set multiple model values to cache.
//
// models must be either map[int64]<Model>, map[string]<Model> or []<Model>,
// otherwise it returns an error.
// If models is a map, the map's keys will be used to generate cache keys.
func (p *ModelCache) MSet(ctx context.Context, models interface{}) (err error) {
	defer recovery("MSet", &err)

	var isSlice, isMap bool
	isSlice = isCacheableSlice(models)
	if !isSlice {
		isMap = isCacheableMap(models)
		if !isMap {
			return errors.New("models is not cacheable slice/map")
		}
	}

	var kvPairs []KVPair
	var kvMap map[string]interface{}
	modelsVal := reflect.ValueOf(models)
	if isSlice {
		kvPairs, kvMap, err = p.marshalCacheableSlice(modelsVal)
	} else {
		kvPairs, kvMap, err = p.marshalCacheableMap(modelsVal)
	}
	if err != nil {
		return err
	}
	if len(kvPairs) == 0 {
		return nil
	}

	client := p.ClientFunc(ctx)
	batchSize := p.MSetBatchSize
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	batches := easy.SplitBatch(len(kvPairs), batchSize)
	for _, idx := range batches {
		batchPairs := kvPairs[idx.I:idx.J]
		err = client.MSet(ctx, batchPairs, p.Expiration)
		if err != nil {
			return err
		}
	}

	if p.LruCache != nil {
		p.LruCache.MSet(kvMap, p.LruExpiration)
	}

	return nil
}

// MDeleteByIntKeys delete cached models from cache using int64 keys.
func (p *ModelCache) MDeleteByIntKeys(ctx context.Context, pks []int64) (err error) {
	if len(pks) == 0 {
		return nil
	}
	defer recovery("MDeleteByIntKeys", &err)

	keys := make([]string, 0, len(pks))
	for _, pk := range pks {
		key := p.KeyFunc(pk)
		keys = append(keys, key)
	}
	client := p.ClientFunc(ctx)
	err = client.MDelete(ctx, keys...)
	if err != nil {
		return err
	}
	if p.LruCache != nil {
		p.LruCache.MDelString(keys...)
	}
	return nil
}

// MDeleteByStringKeys delete cached models from cache using string keys.
func (p *ModelCache) MDeleteByStringKeys(ctx context.Context, pks []string) (err error) {
	if len(pks) == 0 {
		return nil
	}
	defer recovery("MDeleteByStringKeys", &err)

	keys := make([]string, 0, len(pks))
	for _, pk := range pks {
		key := p.KeyFunc(pk)
		keys = append(keys, key)
	}
	client := p.ClientFunc(ctx)
	err = client.MDelete(ctx, keys...)
	if err != nil {
		return err
	}
	if p.LruCache != nil {
		p.LruCache.MDelString(keys...)
	}
	return nil
}

func (p *ModelCache) marshalCacheableSlice(sliceVal reflect.Value) (kvPairs []KVPair, kvMap map[string]interface{}, err error) {
	length := sliceVal.Len()
	kvPairs = make([]KVPair, 0, length)
	kvMap = make(map[string]interface{}, length)
	for i := 0; i < length; i++ {
		elem := sliceVal.Index(i).Interface()
		buf, err := elem.(Model).MarshalModel()
		if err != nil {
			return nil, nil, err
		}
		key := p.KeyFunc(p.IdFunc(elem))
		kvPairs = append(kvPairs, KVPair{
			K: key,
			V: buf,
		})
		kvMap[key] = elem
	}
	return
}

func (p *ModelCache) marshalCacheableMap(mapVal reflect.Value) (kvPairs []KVPair, kvMap map[string]interface{}, err error) {
	length := mapVal.Len()
	kvPairs = make([]KVPair, 0, length)
	kvMap = make(map[string]interface{}, length)
	iter := mapVal.MapRange()
	for iter.Next() {
		keyVal := iter.Key()
		elem := iter.Value().Interface()
		buf, err := elem.(Model).MarshalModel()
		if err != nil {
			return nil, nil, err
		}
		key := p.KeyFunc(keyVal.Interface())
		kvPairs = append(kvPairs, KVPair{
			K: key,
			V: buf,
		})
		kvMap[key] = elem
	}
	return
}

func (p *ModelCache) unmarshalCacheableSlice(slicePtr interface{}, lruItems []interface{}, values [][]byte) error {
	slicePtrVal := reflect.ValueOf(slicePtr)
	sliceTyp := slicePtrVal.Type().Elem()
	elemTyp := sliceTyp.Elem()
	sliceVal := reflect.MakeSlice(sliceTyp, 0, len(lruItems)+len(values))
	for _, item := range lruItems {
		sliceVal = reflect.Append(sliceVal, reflect.ValueOf(item))
	}
	for _, val := range values {
		if len(val) == 0 {
			continue
		}
		elemVal := reflect.New(elemTyp.Elem())
		err := elemVal.Interface().(Model).UnmarshalModel(val)
		if err != nil {
			return err
		}
		sliceVal = reflect.Append(sliceVal, elemVal)
	}
	slicePtrVal.Elem().Set(sliceVal)
	return nil
}

func (p *ModelCache) unmarshalCacheableMap(mapPtr interface{}, lruItems []interface{}, values [][]byte) error {
	mapPtrVal := reflect.ValueOf(mapPtr)
	mapTyp := mapPtrVal.Type().Elem()
	elemTyp := mapTyp.Elem()
	mapVal := reflect.MakeMapWithSize(mapTyp, len(lruItems)+len(values))
	for _, item := range lruItems {
		pk := p.IdFunc(item)
		mapVal.SetMapIndex(reflect.ValueOf(pk), reflect.ValueOf(item))
	}
	for _, val := range values {
		if len(val) == 0 {
			continue
		}
		elemVal := reflect.New(elemTyp.Elem())
		err := elemVal.Interface().(Model).UnmarshalModel(val)
		if err != nil {
			return err
		}
		pk := p.IdFunc(elemVal.Interface())
		mapVal.SetMapIndex(reflect.ValueOf(pk), elemVal)
	}
	mapPtrVal.Elem().Set(mapVal)
	return nil
}

var cacheableSliceInterfaceTyp = reflect.TypeOf((*Model)(nil)).Elem()

func isCacheableSlice(models interface{}) bool {
	typ := reflect.TypeOf(models)
	return isCacheableSliceType(typ)
}

func isCacheableSliceType(sliceTyp reflect.Type) bool {
	if sliceTyp.Kind() != reflect.Slice {
		return false
	}
	elemTyp := sliceTyp.Elem()
	if elemTyp.Implements(cacheableSliceInterfaceTyp) {
		return true
	}
	return false
}

func isCacheableMap(models interface{}) bool {
	typ := reflect.TypeOf(models)
	if typ.Kind() != reflect.Map {
		return false
	}
	keyTyp := typ.Key()
	keyKind := keyTyp.Kind()
	if keyKind != reflect.Int64 && keyKind != reflect.String {
		return false
	}
	elemTyp := typ.Elem()
	if elemTyp.Implements(cacheableSliceInterfaceTyp) {
		return true
	}
	return false
}

func isCacheableSlicePointer(slicePtr interface{}) bool {
	typ := reflect.TypeOf(slicePtr)
	if typ.Kind() != reflect.Ptr {
		return false
	}
	sliceTyp := typ.Elem()
	return isCacheableSliceType(sliceTyp)
}

func isIntCacheableMapPointer(mapPtr interface{}) bool {
	typ := reflect.TypeOf(mapPtr)
	if typ.Kind() != reflect.Ptr {
		return false
	}
	mapTyp := typ.Elem()
	if mapTyp.Kind() != reflect.Map {
		return false
	}
	if mapTyp.Key().Kind() != reflect.Int64 {
		return false
	}
	elemTyp := mapTyp.Elem()
	if elemTyp.Implements(cacheableSliceInterfaceTyp) {
		return true
	}
	return false
}

func isStringCacheableMapPointer(mapPtr interface{}) bool {
	typ := reflect.TypeOf(mapPtr)
	if typ.Kind() != reflect.Ptr {
		return false
	}
	mapTyp := typ.Elem()
	if mapTyp.Kind() != reflect.Map {
		return false
	}
	if mapTyp.Key().Kind() != reflect.String {
		return false
	}
	elemTyp := mapTyp.Elem()
	if elemTyp.Implements(cacheableSliceInterfaceTyp) {
		return true
	}
	return false
}

func recovery(where string, err *error) {
	r := recover()
	if r != nil {
		*err = fmt.Errorf("panic %s: %v", where, err)
	}
}
