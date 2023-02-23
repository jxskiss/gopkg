package kvutil

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/easy"
	"github.com/jxskiss/gopkg/v2/perf/lru"
)

var (
	testModelList = []*TestModel{
		{
			IntId: 111,
			StrId: "aaa",
		},
		{
			IntId: 112,
			StrId: "aab",
		},
	}
	testModelMapInt = map[int64]*TestModel{
		113: {
			IntId: 113,
			StrId: "aac",
		},
	}
	testModelMapStr = map[string]*TestModel{
		"aac": {
			IntId: 113,
			StrId: "aac",
		},
	}
	testIntIds       = []int64{111, 112, 113}
	testStrIds       = []string{"aaa", "aab", "aac"}
	testDeleteIntIds = []int64{111, 112}
	testDeleteStrIds = []string{"aab", "aac"}
)

func TestCache(t *testing.T) {
	mcInt := makeTestingCache("TestIntCache",
		func(m *TestModel) int64 {
			return m.IntId
		})
	mcStr := makeTestingCache("TestStrCache",
		func(m *TestModel) string {
			return m.StrId
		})

	ctx := context.Background()
	var modelList []*TestModel
	var modelIntMap = make(map[int64]*TestModel)
	var modelStrMap = make(map[string]*TestModel)
	var err error

	modelList, err = mcInt.MGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	modelList, err = mcStr.MGetSlice(ctx, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	modelIntMap, err = mcInt.MGetMap(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelIntMap, 0)

	modelStrMap, err = mcStr.MGetMap(ctx, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelStrMap, 0)

	// we can populate cache using either a list or a map
	err = mcInt.MSetSlice(ctx, testModelList, 0)
	assert.Nil(t, err)
	err = mcInt.MSetMap(ctx, testModelMapInt, 0)
	assert.Nil(t, err)

	err = mcStr.MSetSlice(ctx, testModelList, 0)
	assert.Nil(t, err)
	err = mcStr.MSetMap(ctx, testModelMapStr, 0)
	assert.Nil(t, err)

	modelList, err = mcInt.MGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 3)

	modelList, err = mcStr.MGetSlice(ctx, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 3)

	modelIntMap, err = mcInt.MGetMap(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelIntMap, 3)

	modelStrMap, err = mcStr.MGetMap(ctx, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelStrMap, 3)

	err = mcInt.MDelete(ctx, testDeleteIntIds)
	assert.Nil(t, err)
	assert.Len(t, mcInt.config.Storage(ctx).(*memoryStorage).data, 1)

	err = mcStr.MDelete(ctx, testDeleteStrIds)
	assert.Nil(t, err)
	assert.Len(t, mcStr.config.Storage(ctx).(*memoryStorage).data, 1)
}

func TestCacheWithLRUCache(t *testing.T) {
	mc := makeTestingCache("TestCacheWithLRUCache",
		func(m *TestModel) int64 {
			return m.IntId
		})
	mc.config.LRUCache = lru.NewCache[int64, *TestModel](10)
	mc.config.LRUExpiration = time.Minute

	ctx := context.Background()
	var modelList []*TestModel
	var modelMap = make(map[int64]*TestModel)
	var err error

	modelList, err = mc.MGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	err = mc.MSetSlice(ctx, testModelList, 0)
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

	modelList, err = mc.MGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 2)

	modelMap, err = mc.MGetMap(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 2)

	err = mc.MDelete(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, mc.config.Storage(ctx).(*memoryStorage).data, 0)
	got1, exists1, expired1 = mc.config.LRUCache.Get(111)
	assert.Nil(t, got1)
	assert.False(t, exists1 || expired1)
}

func TestCacheWithLoader(t *testing.T) {
	mc := makeTestingCache("TestCacheWithLoader",
		func(m *TestModel) int64 {
			return m.IntId
		})
	mc.config.LRUCache = lru.NewShardedCache[int64, *TestModel](4, 30)
	mc.config.LRUExpiration = time.Second
	mc.config.Loader = testLoaderFunc
	mc.config.CacheExpiration = time.Hour

	ctx := context.Background()
	var modelList []*TestModel
	var modelMap = make(map[int64]*TestModel)
	var err error

	modelList, err = mc.MGetSlice(ctx, []int64{111, 112, 113})
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
	modelMap, err = mc.MGetMap(ctx, moreIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 4)
	assert.Equal(t, 4, mc.config.LRUCache.Len())
	assert.ElementsMatch(t, []int64{111, 113, 115, 117}, easy.Keys(modelMap, nil))

	cacheKeys := make([]string, len(moreIntIds))
	for i, id := range moreIntIds {
		cacheKeys[i] = mc.config.KeyFunc(id)
	}
	fromCache, err := mc.config.Storage(ctx).MGet(ctx, cacheKeys...)
	assert.Nil(t, err)
	assert.Len(t, fromCache, len(cacheKeys))
	valid := easy.Filter(func(_ int, elem []byte) bool { return len(elem) > 0 }, fromCache)
	assert.Len(t, valid, 4)
}

func TestCacheSingleKeyValue(t *testing.T) {
	mc := makeTestingCache("TestCacheSingleKeyValue",
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

func makeTestingCache[K comparable, V Model](testName string, idFunc func(V) K) *Cache[K, V] {
	km := KeyManager{}
	return NewCache(&CacheConfig[K, V]{
		Storage:          testClientFunc(testName),
		IDFunc:           idFunc,
		KeyFunc:          km.NewKey("test_model:{id}"),
		MGetBatchSize:    2,
		MSetBatchSize:    2,
		MDeleteBatchSize: 2,
	})
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

func (m *memoryStorage) MGet(ctx context.Context, keys ...string) ([][]byte, error) {
	out := make([][]byte, 0, len(keys))
	for _, k := range keys {
		out = append(out, m.data[k])
	}
	return out, nil
}

func (m *memoryStorage) MSet(ctx context.Context, kvPairs []KVPair, expiration time.Duration) error {
	for _, kv := range kvPairs {
		m.data[kv.K] = kv.V
	}
	return nil
}

func (m *memoryStorage) MDelete(ctx context.Context, keys ...string) error {
	for _, k := range keys {
		delete(m.data, k)
	}
	return nil
}

type TestModel struct {
	IntId int64
	StrId string
}

func (t *TestModel) MarshalBinary() ([]byte, error) {
	var buf []byte
	buf = append(buf, []byte(strconv.FormatInt(t.IntId, 10))...)
	buf = append(buf, []byte(t.StrId)...)
	return buf, nil
}

func (t *TestModel) UnmarshalBinary(b []byte) error {
	t.IntId, _ = strconv.ParseInt(string(b[:3]), 10, 64)
	t.StrId = string(b[3:6])
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
