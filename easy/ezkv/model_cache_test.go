package ezkv

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jxskiss/gopkg/v2/easy"
	"github.com/jxskiss/gopkg/v2/perf/lru"
	"github.com/jxskiss/gopkg/v2/utils/compress"
)

var (
	testModelList = []*TestModel{
		{
			IntId:  111,
			StrId:  "aaa",
			Field3: strings.Repeat("aaa", 100),
		},
		{
			IntId:  112,
			StrId:  "aab",
			Field3: strings.Repeat("aab", 100),
		},
	}
	testModelMapInt = map[int64]*TestModel{
		113: {
			IntId:  113,
			StrId:  "aac",
			Field3: strings.Repeat("aac", 100),
		},
	}
	testModelMapStr = map[string]*TestModel{
		"aac": {
			IntId:  113,
			StrId:  "aac",
			Field3: strings.Repeat("aac", 100),
		},
	}
	testIntIds       = []int64{111, 112, 113}
	testStrIds       = []string{"aaa", "aab", "aac"}
	testDeleteIntIds = []int64{111, 112}
	testDeleteStrIds = []string{"aab", "aac"}
)

func TestIsSetCacheError(t *testing.T) {
	assert.True(t, IsSetCacheError(&setCacheError{errors.New("test error")}))
}

func TestCache(t *testing.T) {
	mcInt := makeTestingCache(t,
		func(m *TestModel) int64 {
			return m.IntId
		})
	mcStr := makeTestingCache(t,
		func(m *TestModel) string {
			return m.StrId
		})

	ctx := context.Background()
	var modelList []*TestModel
	var modelIntMap map[int64]*TestModel
	var modelStrMap map[string]*TestModel
	var err error

	modelList, err = mcInt.BatchGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	modelList, err = mcStr.BatchGetSlice(ctx, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	modelIntMap, err = mcInt.BatchGetMap(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelIntMap, 0)

	modelStrMap, err = mcStr.BatchGetMap(ctx, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelStrMap, 0)

	// we can populate cache using either a list or a map
	err = mcInt.BatchSetSlice(ctx, testModelList, 0)
	assert.Nil(t, err)
	err = mcInt.BatchSetMap(ctx, testModelMapInt, 0)
	assert.Nil(t, err)

	err = mcStr.BatchSetSlice(ctx, testModelList, 0)
	assert.Nil(t, err)
	err = mcStr.BatchSetMap(ctx, testModelMapStr, 0)
	assert.Nil(t, err)

	modelList, err = mcInt.BatchGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 3)

	modelList, err = mcStr.BatchGetSlice(ctx, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 3)

	modelIntMap, err = mcInt.BatchGetMap(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelIntMap, 3)

	modelStrMap, err = mcStr.BatchGetMap(ctx, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelStrMap, 3)

	err = mcInt.Delete(ctx, testDeleteIntIds...)
	assert.Nil(t, err)
	assert.Len(t, mcInt.config.Storage(ctx).(*memoryStorage).data, 1)

	err = mcStr.Delete(ctx, testDeleteStrIds...)
	assert.Nil(t, err)
	assert.Len(t, mcStr.config.Storage(ctx).(*memoryStorage).data, 1)
}

func TestCacheWithLRUCache(t *testing.T) {
	mc := makeTestingCache(t,
		func(m *TestModel) int64 {
			return m.IntId
		})
	mc.config.LRUCache = lru.NewCache[int64, *TestModel](10)
	mc.config.LRUExpiration = time.Minute

	ctx := context.Background()
	var modelList []*TestModel
	var modelMap map[int64]*TestModel
	var err error

	modelList, err = mc.BatchGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	err = mc.BatchSetSlice(ctx, testModelList, 0)
	assert.Nil(t, err)
	assert.Len(t, mc.config.Storage(ctx).(*memoryStorage).data, 2)

	got1, exists1, expired1 := mc.config.LRUCache.Get(111)
	assert.NotNil(t, got1)
	assert.True(t, exists1 && !expired1)

	got2, exists2, expired2 := mc.config.LRUCache.Get(112)
	assert.NotNil(t, got2)
	assert.True(t, exists2 && !expired2)

	got3, exists3, expired3 := mc.config.LRUCache.Get(113)
	assert.Nil(t, got3)
	assert.False(t, exists3 || expired3)

	modelList, err = mc.BatchGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 2)

	modelMap, err = mc.BatchGetMap(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 2)

	err = mc.Delete(ctx, testIntIds...)
	assert.Nil(t, err)
	assert.Len(t, mc.config.Storage(ctx).(*memoryStorage).data, 0)
	got1, exists1, expired1 = mc.config.LRUCache.Get(111)
	assert.Nil(t, got1)
	assert.False(t, exists1 || expired1)
}

func TestCacheWithLoader(t *testing.T) {
	mc := makeTestingCache(t,
		func(m *TestModel) int64 {
			return m.IntId
		})
	mc.config.LRUCache = lru.NewShardedCache[int64, *TestModel](4, 30)
	mc.config.LRUExpiration = time.Second
	mc.config.Loader = testLoaderFunc
	mc.config.CacheExpiration = time.Hour

	ctx := context.Background()
	var modelList []*TestModel
	var modelMap map[int64]*TestModel
	var err error

	modelList, err = mc.BatchGetSlice(ctx, []int64{111, 112, 113})
	assert.Nil(t, err)
	assert.Len(t, modelList, 2)
	assert.Equal(t, 2, mc.config.LRUCache.Len())
	_, exists := mc.config.LRUCache.GetNotStale(int64(111))
	assert.True(t, exists)
	_, exists = mc.config.LRUCache.GetNotStale(int64(112))
	assert.False(t, exists)
	_, exists = mc.config.LRUCache.GetNotStale(int64(113))
	assert.True(t, exists)

	time.Sleep(mc.config.LRUExpiration)
	_, exists = mc.config.LRUCache.GetNotStale(int64(111))
	assert.False(t, exists)
	_, exists = mc.config.LRUCache.GetNotStale(int64(112))
	assert.False(t, exists)
	_, exists = mc.config.LRUCache.GetNotStale(int64(113))
	assert.False(t, exists)

	moreIntIds := []int64{111, 112, 113, 114, 115, 116, 117}
	modelMap, err = mc.BatchGetMap(ctx, moreIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 4)
	assert.Equal(t, 4, mc.config.LRUCache.Len())
	assert.ElementsMatch(t, []int64{111, 113, 115, 117}, easy.Keys(modelMap, nil))

	cacheKeys := make([]string, len(moreIntIds))
	for i, id := range moreIntIds {
		cacheKeys[i] = mc.config.KeyFunc(id)
	}
	fromCache, err := mc.config.Storage(ctx).Get(ctx, cacheKeys...)
	assert.Nil(t, err)
	assert.Len(t, fromCache, len(cacheKeys))
	valid := easy.Filter(func(_ int, elem []byte) bool { return len(elem) > 0 }, fromCache)
	assert.Len(t, valid, 4)
}

func TestCacheSingleKeyValue(t *testing.T) {
	mc := makeTestingCache(t,
		func(m *TestModel) int64 {
			return m.IntId
		})
	mc.config.LRUCache = lru.NewCache[int64, *TestModel](5)
	mc.config.LRUExpiration = time.Second
	mc.config.Loader = testLoaderFunc
	mc.config.CacheExpiration = time.Hour

	ctx := context.Background()
	val111, err := mc.Get(ctx, 111)
	assert.Nil(t, err)
	assert.Equal(t, int64(111), val111.IntId)
	_, exists := mc.config.LRUCache.GetNotStale(111)
	assert.True(t, exists)

	mc.config.LRUCache.Delete(111)
	val111, err = mc.Get(ctx, 111)
	assert.Nil(t, err)
	assert.Equal(t, int64(111), val111.IntId)
	_, exists = mc.config.LRUCache.GetNotStale(111)
	assert.True(t, exists)

	val112, err := mc.Get(ctx, 112)
	assert.Equal(t, ErrDataNotFound, err)
	assert.Nil(t, val112)
	_, exists = mc.config.LRUCache.GetNotStale(112)
	assert.False(t, exists)

	err = mc.Delete(ctx, 111)
	assert.Nil(t, err)
	_, exists = mc.config.LRUCache.GetNotStale(111)
	assert.False(t, exists)
}

func TestCacheWithCompression(t *testing.T) {
	mcInt := makeTestingCache(t,
		func(m *TestModel) int64 {
			return m.IntId
		})
	codec, codecErr := compress.NewGzipCodec(gzip.BestSpeed)
	if codecErr != nil {
		t.Fatal(codecErr)
	}
	compressor, configErr := compress.NewCompressor(compress.CompressorConfig{
		Codec:             codec,
		Threshold:         1,
		MinReductionRatio: 0.00001,
	})
	if configErr != nil {
		t.Fatal(configErr)
	}
	mcInt.config.Compressor = compressor

	ctx := context.Background()
	var modelList []*TestModel
	var err error

	modelList, err = mcInt.BatchGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	// we can populate cache using either a list or a map
	err = mcInt.BatchSetSlice(ctx, testModelList, 0)
	assert.Nil(t, err)

	var rawSize int
	for _, m := range testModelList {
		data, _ := m.MarshalBinary()
		rawSize += len(data)
	}
	var compressedSize int
	for _, data := range mcInt.config.Storage(ctx).(*memoryStorage).data {
		compressedSize += len(data)
	}
	assert.Less(t, compressedSize, rawSize)

	modelList, err = mcInt.BatchGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 2)
}

func TestNewModelCacheValidation(t *testing.T) {
	for _, field := range []string{"config", "Storage", "IDFunc", "KeyFunc"} {
		t.Run(field, func(t *testing.T) {
			cfg := &ModelCacheConfig[int64, *TestModel]{
				Storage: testClientFunc(t.Name()),
				IDFunc:  func(m *TestModel) int64 { return m.IntId },
				KeyFunc: func(id int64) string { return fmt.Sprint(id) },
			}
			switch field {
			case "config":
				cfg = nil
			case "Storage":
				cfg.Storage = nil
			case "IDFunc":
				cfg.IDFunc = nil
			case "KeyFunc":
				cfg.KeyFunc = nil
			}
			cache, err := NewModelCache(cfg)
			require.Nil(t, cache)
			require.ErrorContains(t, err, field)
		})
	}

	t.Run("interface model", func(t *testing.T) {
		cache, err := NewModelCache(&ModelCacheConfig[int64, Model]{
			Storage: testClientFunc(t.Name()),
			IDFunc:  func(m Model) int64 { return m.(*TestModel).IntId },
			KeyFunc: func(id int64) string { return fmt.Sprint(id) },
		})
		require.Nil(t, cache)
		require.ErrorContains(t, err, "concrete")
	})

	t.Run("defaults", func(t *testing.T) {
		cache, err := NewModelCache(&ModelCacheConfig[int64, *TestModel]{
			Storage:         testClientFunc(t.Name()),
			IDFunc:          func(m *TestModel) int64 { return m.IntId },
			KeyFunc:         func(id int64) string { return fmt.Sprint(id) },
			BatchGetSize:    -1,
			BatchSetSize:    -1,
			BatchDeleteSize: -1,
			LoaderBatchSize: -1,
			LRUExpiration:   -1,
		})
		require.NoError(t, err)
		ctx := context.Background()
		model := &TestModel{IntId: 111, StrId: "abc"}
		require.NoError(t, cache.BatchSetMap(ctx, map[int64]*TestModel{111: model}, 0))
		got, err := cache.BatchGetMap(ctx, []int64{111})
		require.NoError(t, err)
		require.Equal(t, map[int64]*TestModel{111: model}, got)
		require.NoError(t, cache.Delete(ctx, 111, 112))
		_, err = cache.Get(ctx, 111)
		require.ErrorIs(t, err, ErrDataNotFound)
	})
}

func TestModelCacheBatchLRU(t *testing.T) {
	ctx := context.Background()
	for _, operation := range []string{"set", "delete"} {
		for _, batchSize := range []int{1, 2} {
			for _, failAt := range []int{0, 1, 2} {
				t.Run(fmt.Sprintf("%s/batch=%d/fail=%d", operation, batchSize, failAt), func(t *testing.T) {
					cache := makeTestingCache(t, func(m *TestModel) int64 { return m.IntId })
					stor := &failingBatchStorage{
						memoryStorage: memoryStorage{data: make(map[string][]byte)},
						failAt:        failAt,
						err:           errors.New("batch failed"),
					}
					cache.config.Storage = func(context.Context) Storage { return stor }
					cache.config.LRUCache = lru.NewCache[int64, *TestModel](10)
					cache.config.BatchSetSize = batchSize
					cache.config.BatchDeleteSize = batchSize
					models := make(map[int64]*TestModel)
					for _, id := range []int64{111, 112, 113} {
						require.NoError(t, cache.Set(ctx, id, &TestModel{IntId: id, StrId: "old"}, 0))
						models[id] = &TestModel{IntId: id, StrId: "new"}
					}
					var err error
					if operation == "set" {
						err = cache.BatchSetMap(ctx, models, 0)
					} else {
						err = cache.Delete(ctx, 111, 112, 113)
					}
					if failAt == 0 {
						require.NoError(t, err)
					} else {
						require.ErrorIs(t, err, stor.err)
					}
					for _, id := range []int64{111, 112, 113} {
						data := stor.data[cache.config.KeyFunc(id)]
						cached, exists := cache.config.LRUCache.GetNotStale(id)
						if data == nil {
							require.False(t, exists, "deleted key %d remains in LRU", id)
							continue
						}
						var stored TestModel
						require.NoError(t, stored.UnmarshalBinary(data))
						require.True(t, exists)
						require.Equal(t, &stored, cached, "LRU differs from storage for key %d", id)
					}
				})
			}
		}
	}
}

func makeTestingCache[K comparable, V Model](t testing.TB, idFunc func(V) K) *ModelCache[K, V] {
	t.Helper()
	kf := KeyFactory{}
	keyFun := kf.NewKey("test_model:{id}")
	cache, err := NewModelCache(&ModelCacheConfig[K, V]{
		Storage:         testClientFunc(t.Name()),
		IDFunc:          idFunc,
		KeyFunc:         func(pk K) string { return keyFun(pk) },
		BatchGetSize:    2,
		BatchSetSize:    2,
		BatchDeleteSize: 2,
	})
	require.NoError(t, err)
	return cache
}

func testClientFunc(testName string) func(ctx context.Context) Storage {
	data := make(map[string][]byte)
	return func(ctx context.Context) Storage {
		return &memoryStorage{data: data}
	}
}

type memoryStorage struct {
	data map[string][]byte
}

func (m *memoryStorage) Get(ctx context.Context, keys ...string) ([][]byte, error) {
	out := make([][]byte, 0, len(keys))
	for _, k := range keys {
		out = append(out, m.data[k])
	}
	return out, nil
}

func (m *memoryStorage) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *memoryStorage) BatchSet(ctx context.Context, keys []string, values [][]byte, expiration time.Duration) error {
	for i, k := range keys {
		m.data[k] = values[i]
	}
	return nil
}

func (m *memoryStorage) Delete(ctx context.Context, keys ...string) error {
	for _, k := range keys {
		delete(m.data, k)
	}
	return nil
}

func clearMemoryStorage(ctx context.Context, cliFunc func(ctx context.Context) Storage) {
	stor := cliFunc(ctx).(*memoryStorage)
	stor.data = make(map[string][]byte)
}

func getMemoryStorage(ctx context.Context, cliFunc func(ctx context.Context) Storage) *memoryStorage {
	stor := cliFunc(ctx).(*memoryStorage)
	return stor
}

type TestModel struct {
	IntId  int64
	StrId  string
	Field3 string
}

func (m *TestModel) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 0, 20+len(m.StrId)+len(m.Field3))
	buf = append(buf, []byte(strconv.FormatInt(m.IntId, 10))...)
	buf = append(buf, []byte(m.StrId)...)
	buf = append(buf, []byte(m.Field3)...)
	return buf, nil
}

func (m *TestModel) UnmarshalBinary(b []byte) error {
	m.IntId, _ = strconv.ParseInt(string(b[:3]), 10, 64)
	m.StrId = string(b[3:6])
	m.Field3 = string(b[6:])
	return nil
}

func testLoaderFunc(ctx context.Context, ids []int64) (map[int64]*TestModel, error) {
	out := make(map[int64]*TestModel, len(ids))
	for _, id := range ids {
		if id%2 == 0 {
			continue
		}
		out[id] = &TestModel{
			IntId: id,
			StrId: strconv.FormatInt(id, 10),
		}
	}
	return out, nil
}

type failingBatchStorage struct {
	memoryStorage
	calls  int
	failAt int
	err    error
}

func (s *failingBatchStorage) BatchSet(ctx context.Context, keys []string, values [][]byte, expiration time.Duration) error {
	s.calls++
	if s.calls == s.failAt {
		return s.err
	}
	return s.memoryStorage.BatchSet(ctx, keys, values, expiration)
}

func (s *failingBatchStorage) Delete(ctx context.Context, keys ...string) error {
	s.calls++
	if s.calls == s.failAt {
		return s.err
	}
	return s.memoryStorage.Delete(ctx, keys...)
}
