package kvutil

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jxskiss/gopkg/v2/easy"
)

// ShardingModel is the interface implemented by types that can be cached
// by ShardingCache.
//
// A ShardingModel implementation type should save ShardingData with it,
// and can tell whether it's a shard or a complete model.
type ShardingModel interface {
	Model

	// GetShardingData returns whether a model is a shard or a complete model.
	// When the returned bool value is true, the returned ShardingData
	// can be a zero value, and shall not be used.
	GetShardingData() (data ShardingData, isShard bool)

	// SetShardingData sets a shard data to a new model created when
	// doing serialization to split data into shards and save to storage.
	SetShardingData(data ShardingData)
}

// ShardingData holds sharding information and partial data of a sharding.
//
// Example:
//
//	// protobuf
//	message ShardingData {
//		int32 total_num = 1;
//		int32 shard_num = 2;
//		bytes digest = 3;
//		bytes data = 4;
//	}
type ShardingData struct {
	TotalNum int32  `protobuf:"varint,1,opt,name=total_num,json=totalNum,proto3" json:"total_num,omitempty"`
	ShardNum int32  `protobuf:"varint,2,opt,name=shard_num,json=shardNum,proto3" json:"shard_num,omitempty"`
	Digest   []byte `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
	Data     []byte `protobuf:"bytes,4,opt,name=data,json=data,proto3" json:"data,omitempty"`
}

// ShardingCacheConfig configures a ShardingCache instance.
type ShardingCacheConfig[K comparable, V ShardingModel] struct {

	// Storage must return a Storage implementation which will be used
	// as the underlying key-value storage.
	Storage func(ctx context.Context) Storage

	// IDFunc returns the primary key of a ShardingModel object.
	IDFunc func(V) K

	// KeyFunc specifies the key function to use with the storage.
	KeyFunc Key

	// ShardingSize configures the maximum length of data in a shard.
	// When the serialization result of V is longer than ShardingSize,
	// the data will be split into shards to save to storage.
	//
	// ShardingSize must be greater than zero, else it panics.
	ShardingSize int

	// MGetBatchSize optionally specifies the batch size for one MGet
	// calling to storage. The default is 200.
	MGetBatchSize int

	// MSetBatchSize optionally specifies the batch size for one MSet
	// calling to storage. The default is 200.
	MSetBatchSize int

	// DeleteBatchSize optionally specifies the batch size for one Delete
	// calling to storage. The default is 200.
	DeleteBatchSize int
}

func (p *ShardingCacheConfig[K, V]) checkAndSetDefaults() {
	if p.ShardingSize <= 0 {
		panic("kvutil: ShardingCacheConfig.ShardingSize must be greater than zero")
	}
	if p.MGetBatchSize <= 0 {
		p.MGetBatchSize = DefaultBatchSize
	}
	if p.MSetBatchSize <= 0 {
		p.MSetBatchSize = DefaultBatchSize
	}
	if p.DeleteBatchSize <= 0 {
		p.DeleteBatchSize = DefaultBatchSize
	}
}

// NewShardingCache returns a new ShardingCache instance.
func NewShardingCache[K comparable, V ShardingModel](config *ShardingCacheConfig[K, V]) *ShardingCache[K, V] {
	config.checkAndSetDefaults()
	newElemFn := buildNewElemFunc[V]()
	return &ShardingCache[K, V]{
		config:      config,
		newElemFunc: newElemFn,
	}
}

// ShardingCache implements common cache operations for big cache value,
// it helps to split big value into shards according to
// ShardingCacheConfig.ShardingSize.
//
// When saving data to cache storage, it checks length of the serialization
// result, if it does not exceed ShardingSize, it saves one key-value
// to storage, else it splits the serialization result to multiple shards
// and saves multiple key-values to storage.
//
// When doing query, it first loads the first shard from storage and checks
// whether there are more shards, if yes, it builds the keys of other shards
// using the information in the first shard, then reads the other shards,
// and concat all data to deserialize the complete model.
//
// A ShardingCache must not be copied after initialization.
type ShardingCache[K comparable, V ShardingModel] struct {
	config *ShardingCacheConfig[K, V]

	newElemFunc func() V
	queryPool   sync.Pool
}

// Set writes a key value pair to ShardingCache.
func (p *ShardingCache[K, V]) Set(ctx context.Context, pk K, elem V, expiration time.Duration) error {
	_ = pk
	return p.MSetSlice(ctx, []V{elem}, expiration)
}

// MSetSlice serializes and writes multiple models to ShardingCache.
func (p *ShardingCache[K, V]) MSetSlice(ctx context.Context, models []V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}
	kvPairs, err := p.marshalModels(models)
	if err != nil {
		return fmt.Errorf("cannot marshal models: %w", err)
	}
	stor := p.config.Storage(ctx)
	return msetToStorage(ctx, stor, kvPairs, expiration, p.config.MSetBatchSize)
}

// MSetMap serializes and writes multiple models to ShardingCache.
func (p *ShardingCache[K, V]) MSetMap(ctx context.Context, models map[K]V, expiration time.Duration) error {
	if len(models) == 0 {
		return nil
	}
	valueSlice := easy.Values(models)
	return p.MSetSlice(ctx, valueSlice, expiration)
}

// Delete deletes key values from ShardingCache.
//
// By default, it only deletes the first shards from storage,
// if the underlying storage is Redis, the other shards shall be evicted
// when they are expired.
// If the underlying storage does not support auto eviction, or the data
// does not expire, or user want to release storage space actively,
// deleteAllShards should be set to true, which indicates it to read
// the first shard from storage and checks whether there are more shards,
// it yes, it builds the keys of other shards using the information
// in the first shard, then deletes all shards from storage.
func (p *ShardingCache[K, V]) Delete(ctx context.Context, deleteAllShards bool, pks ...K) error {
	if len(pks) == 0 {
		return nil
	}
	if deleteAllShards {
		return p.deleteAllShards(ctx, pks)
	}

	keys := make([]string, 0, len(pks))
	for _, pk := range pks {
		keys = append(keys, p.config.KeyFunc(pk))
	}

	stor := p.config.Storage(ctx)
	batches := easy.Split(keys, p.config.DeleteBatchSize)
	for _, bat := range batches {
		err := stor.Delete(ctx, bat...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *ShardingCache[K, V]) deleteAllShards(ctx context.Context, pks []K) error {
	keys := make([]string, 0, len(pks))
	for _, pk := range pks {
		keys = append(keys, p.config.KeyFunc(pk))
	}

	stor := p.config.Storage(ctx)
	mgetRet, err := mgetFromStorage(ctx, stor, keys, p.config.MGetBatchSize)
	if err != nil {
		return err
	}
	for i, data := range mgetRet {
		if len(data) == 0 {
			continue
		}
		key := keys[i]
		elem := p.newElemFunc()
		err = elem.UnmarshalBinary(data)
		if err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}
		shard0, isShard := elem.GetShardingData()
		if isShard {
			for j := 1; j < int(shard0.TotalNum); j++ {
				ithKey := GetShardKey(key, j)
				keys = append(keys, ithKey)
			}
		}
	}

	batches := easy.Split(keys, p.config.DeleteBatchSize)
	for _, bat := range batches {
		err = stor.Delete(ctx, bat...)
		if err != nil {
			return fmt.Errorf("write storage: %w", err)
		}
	}
	return nil
}

func (p *ShardingCache[K, V]) marshalModels(entityList []V) (result []KVPair, err error) {
	idFunc := p.config.IDFunc
	keyFunc := p.config.KeyFunc
	newElemFunc := p.newElemFunc
	shardSize := p.config.ShardingSize
	for _, elem := range entityList {
		buf, err := elem.MarshalBinary()
		if err != nil {
			return nil, err
		}
		pk := idFunc(elem)
		key := keyFunc(pk)
		if len(buf) <= shardSize {
			result = append(result, KVPair{K: key, V: buf})
			continue
		}

		// Split big value into shards.
		totalNum := (len(buf) + shardSize - 1) / shardSize
		digest := calcDigest(buf)
		shard0 := newElemFunc()
		shard0.SetShardingData(ShardingData{
			TotalNum: int32(totalNum),
			ShardNum: 0,
			Digest:   digest,
			Data:     buf[:shardSize],
		})
		shard0Buf, err := shard0.MarshalBinary()
		if err != nil {
			return nil, err
		}

		result = append(result, KVPair{K: key, V: shard0Buf})
		for num := 1; num < totalNum; num++ {
			i := num * shardSize
			j := min((num+1)*shardSize, len(buf))
			ithKey := GetShardKey(key, num)
			ithShard := newElemFunc()
			ithShard.SetShardingData(ShardingData{
				TotalNum: int32(totalNum),
				ShardNum: int32(num),
				Digest:   digest,
				Data:     buf[i:j],
			})
			ithBuf, err := ithShard.MarshalBinary()
			if err != nil {
				return nil, err
			}
			result = append(result, KVPair{K: ithKey, V: ithBuf})
		}
	}
	return result, nil
}

// Get queries ShardingCache for a given pk.
//
// If pk cannot be found in the cache, it returns an error ErrDataNotFound.
func (p *ShardingCache[K, V]) Get(ctx context.Context, pk K) (V, error) {
	var zeroVal V
	stor := p.config.Storage(ctx)
	key := p.config.KeyFunc(pk)
	cacheResult, err := stor.MGet(ctx, key)
	if err != nil {
		return zeroVal, fmt.Errorf("query storage: %w", err)
	}

	if len(cacheResult) == 0 || len(cacheResult[0]) == 0 {
		return zeroVal, ErrDataNotFound
	}

	elem := p.newElemFunc()
	err = elem.UnmarshalBinary(cacheResult[0])
	if err != nil {
		return zeroVal, fmt.Errorf("cannot unmarshal data: %w", err)
	}
	shard0, isShard := elem.GetShardingData()
	if !isShard {
		return elem, nil
	}

	ithKeys := make([]string, 0, shard0.TotalNum-1)
	for i := 1; i < int(shard0.TotalNum); i++ {
		ithKey := GetShardKey(key, i)
		ithKeys = append(ithKeys, ithKey)
	}
	ithCacheResult, err := mgetFromStorage(ctx, stor, ithKeys, p.config.MGetBatchSize)
	if err != nil {
		return zeroVal, err
	}

	buf := shard0.Data
	for i := 0; i < len(ithKeys); i++ {
		ithKey := ithKeys[i]
		ithRet := ithCacheResult[i]
		if len(ithRet) == 0 {
			return zeroVal, fmt.Errorf("sharding data not found: %s", ithKey)
		}
		ithVal := p.newElemFunc()
		err = ithVal.UnmarshalBinary(ithRet)
		if err != nil {
			return zeroVal, fmt.Errorf("cannot unmarshal data: %w", err)
		}
		shardNum := getShardNumFromKey(ithKey)
		ithShard, isShard := ithVal.GetShardingData()
		if !isShard || ithShard.ShardNum != shardNum {
			return zeroVal, fmt.Errorf("sharding data shardNum not match: %s", ithKey)
		}
		if !bytes.Equal(ithShard.Digest, shard0.Digest) {
			return zeroVal, fmt.Errorf("sharding data digest not match: %s", ithKey)
		}
		buf = append(buf, ithShard.Data...)
	}
	elem = p.newElemFunc()
	err = elem.UnmarshalBinary(buf)
	if err != nil {
		return zeroVal, fmt.Errorf("cannot unmarshal data: %w", err)
	}
	return elem, nil
}

// MGetSlice queries ShardingCache for multiple pks and returns
// the cached values as a slice of type []V.
// Note the returned values may be less than requested.
func (p *ShardingCache[K, V]) MGetSlice(ctx context.Context, pks []K) (
	result []V, errMap map[K]error, storageErr error) {
	mapRet, errMap, storageErr := p.MGetMap(ctx, pks)
	if len(mapRet) > 0 {
		result = make([]V, 0, len(mapRet))
		for _, pk := range pks {
			if val, ok := mapRet[pk]; ok {
				result = append(result, val)
			}
		}
	}
	return result, errMap, storageErr
}

// MGetMap queries ShardingCache for multiple pks and returns
// the cached values as a map of type map[K]V.
// Note the returned values may be less than requested.
func (p *ShardingCache[K, V]) MGetMap(ctx context.Context, pks []K) (
	result map[K]V, errMap map[K]error, storageErr error) {
	type SQ = shardingQuery[K, V]
	var query *SQ
	if x, ok := p.queryPool.Get().(*SQ); ok {
		query = x
	} else {
		query = &SQ{}
	}
	defer func() {
		query.reset()
		p.queryPool.Put(query)
	}()

	query.stor = p.config.Storage(ctx)
	query.mgetBatchSize = p.config.MGetBatchSize
	query.keyFunc = p.config.KeyFunc
	query.newElemFunc = p.newElemFunc
	query.pks = pks
	query.Do(ctx)
	return query.Result()
}

type shardingQuery[K comparable, V ShardingModel] struct {
	stor          Storage
	mgetBatchSize int
	keyFunc       Key
	newElemFunc   func() V

	pks  []K
	keys []string

	shardingPKs  []K
	shard0Data   []ShardingData
	ithShardKeys []string

	result     map[K]V
	retErrs    map[K]error
	storageErr error
}

func (sq *shardingQuery[K, V]) reset() {
	sq.pks = nil
	sq.keys = sq.keys[:0]
	sq.shardingPKs = sq.shardingPKs[:0]
	sq.shard0Data = sq.shard0Data[:0]
	sq.ithShardKeys = sq.ithShardKeys[:0]
	sq.result = nil
	sq.retErrs = nil
	sq.storageErr = nil
}

func (sq *shardingQuery[K, V]) setError(pk K, err error) {
	if sq.retErrs[pk] == nil {
		if sq.retErrs == nil {
			sq.retErrs = make(map[K]error)
		}
		sq.retErrs[pk] = err
	}
}

func (sq *shardingQuery[K, V]) setStorageError(err error) {
	if sq.storageErr == nil {
		sq.storageErr = err
	}
}

func (sq *shardingQuery[K, V]) Do(ctx context.Context) {
	for _, pk := range sq.pks {
		key := sq.keyFunc(pk)
		sq.keys = append(sq.keys, key)
	}

	mgetRet, err := mgetFromStorage(ctx, sq.stor, sq.keys, sq.mgetBatchSize)
	if err != nil {
		sq.setStorageError(err)
		return
	}

	newElemFunc := sq.newElemFunc
	sq.result = make(map[K]V, len(sq.keys))
	sq.retErrs = make(map[K]error, len(sq.keys))
	for i, buf := range mgetRet {
		if len(buf) == 0 {
			continue
		}

		pk := sq.pks[i]
		key := sq.keys[i]
		elem := newElemFunc()
		err = elem.UnmarshalBinary(buf)
		if err != nil {
			tmpErr := fmt.Errorf("cannot unmarshal data: %w", err)
			sq.setError(pk, tmpErr)
			continue
		}

		shard0, isShard := elem.GetShardingData()
		if !isShard {
			sq.result[pk] = elem
			continue
		}

		// The cache data is a shard, we need to read data from all shards.
		sq.shardingPKs = append(sq.shardingPKs, pk)
		sq.shard0Data = append(sq.shard0Data, shard0)
		for j := 1; j < int(shard0.TotalNum); j++ {
			ithKey := GetShardKey(key, j)
			sq.ithShardKeys = append(sq.ithShardKeys, ithKey)
		}
	}

	if len(sq.shardingPKs) > 0 {
		sq.queryAndMergeShardingData(ctx)
	}
}

type ithResult struct {
	Value any
	Err   error
}

func (sq *shardingQuery[K, V]) queryAndMergeShardingData(ctx context.Context) {
	ithShardKeys := sq.ithShardKeys
	shardRet, err := mgetFromStorage(ctx, sq.stor, ithShardKeys, sq.mgetBatchSize)
	if err != nil {
		sq.setStorageError(err)
		return
	}

	newElemFunc := sq.newElemFunc
	ithShardMap := make(map[string]ithResult, len(ithShardKeys))
	for i, ithKey := range ithShardKeys {
		buf := shardRet[i]
		if len(buf) == 0 {
			continue
		}
		elem := newElemFunc()
		err = elem.UnmarshalBinary(buf)
		if err != nil {
			tmpErr := fmt.Errorf("cannot unmarshal data: %w", err)
			ithShardMap[ithKey] = ithResult{Err: tmpErr}
			continue
		}
		ithShardMap[ithKey] = ithResult{Value: elem}
	}

mergeShards:
	for i, pk := range sq.shardingPKs {
		key := sq.keyFunc(pk)
		shard0 := sq.shard0Data[i]
		buf := shard0.Data
		for j := 1; j < int(shard0.TotalNum); j++ {
			ithKey := GetShardKey(key, j)
			ithRet, exists := ithShardMap[ithKey]
			if !exists {
				tmpErr := fmt.Errorf("sharding data not found: %s", ithKey)
				sq.retErrs[pk] = tmpErr
				continue mergeShards
			}
			if ithRet.Err != nil {
				sq.retErrs[pk] = ithRet.Err
				continue mergeShards
			}
			ithVal := ithRet.Value.(V)
			ithShard, isShard := ithVal.GetShardingData()
			if !isShard || ithShard.ShardNum != int32(j) {
				tmpErr := fmt.Errorf("sharding data is invalid: %s", ithKey)
				sq.retErrs[pk] = tmpErr
				continue mergeShards
			}
			if !bytes.Equal(ithShard.Digest, shard0.Digest) {
				tmpErr := fmt.Errorf("sharding data digest not match: %s", ithKey)
				sq.retErrs[pk] = tmpErr
				continue mergeShards
			}
			buf = append(buf, ithShard.Data...)
		}
		elem := newElemFunc()
		err = elem.UnmarshalBinary(buf)
		if err != nil {
			tmpErr := fmt.Errorf("cannot unmarshal data: %w", err)
			sq.retErrs[pk] = tmpErr
			continue mergeShards
		}
		sq.result[pk] = elem
	}
}

func (sq *shardingQuery[K, V]) Result() (map[K]V, map[K]error, error) {
	return sq.result, sq.retErrs, sq.storageErr
}

const shardNumSep = "__"

// GetShardKey returns the shard key for given key and index of shard.
func GetShardKey(key string, i int) string {
	if i == 0 {
		return key
	}
	return key + shardNumSep + strconv.Itoa(i)
}

func getShardNumFromKey(key string) int32 {
	sepIdx := strings.LastIndex(key, shardNumSep)
	if sepIdx <= 0 {
		return 0
	}
	shardNum, _ := strconv.Atoi(key[sepIdx+len(shardNumSep):])
	return int32(shardNum)
}

func calcDigest(data []byte) []byte {
	sum := sha1.Sum(data)
	return sum[:]
}

func mgetFromStorage(ctx context.Context, stor Storage, keys []string, batchSize int) ([][]byte, error) {
	ret := make([][]byte, 0, len(keys))
	for _, batchKeys := range easy.Split(keys, batchSize) {
		batchRet, err := stor.MGet(ctx, batchKeys...)
		if err != nil {
			return nil, fmt.Errorf("query storage: %w", err)
		}
		ret = append(ret, batchRet...)
	}
	return ret, nil
}

func msetToStorage(ctx context.Context, stor Storage, kvPairs []KVPair, expiration time.Duration, batchSize int) error {
	for _, batchKVPairs := range easy.Split(kvPairs, batchSize) {
		err := stor.MSet(ctx, batchKVPairs, expiration)
		if err != nil {
			return fmt.Errorf("write storage: %w", err)
		}
	}
	return nil
}
