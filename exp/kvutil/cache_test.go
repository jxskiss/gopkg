package kvutil

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/lru"
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
	mc := makeTestingCache("TestCache",
		func(m *TestModel) int64 {
			return m.IntId
		})

	ctx := context.Background()
	var modelList []*TestModel
	var modelMap = make(map[int64]*TestModel)
	var err error

	modelList, err = mc.MGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	modelMap, err = mc.MGetMap(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 0)

	// we can populate cache using either a list or a map
	err = mc.MSetSlice(ctx, testModelList, 0)
	assert.Nil(t, err)
	err = mc.MSetMap(ctx, testModelMapInt, 0)
	assert.Nil(t, err)

	modelList, err = mc.MGetSlice(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 3)

	modelMap, err = mc.MGetMap(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 3)

	err = mc.MDelete(ctx, testDeleteIntIds)
	assert.Nil(t, err)
	assert.Len(t, mc.config.Storage(ctx).(*memoryStorage).data, 1)
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

func makeTestingCache[K comparable, V Model](testName string, idFunc func(V) K) *Cache[K, V] {
	km := KeyManager{}
	return NewCache(&CacheConfig[K, V]{
		Storage:          testClientFunc(testName),
		IdFunc:           idFunc,
		KeyFunc:          km.NewKey("test_model:{id}"),
		MGetBatchSize:    2,
		MSetBatchSize:    2,
		MDeleteBatchSize: 2,
	})
}

var testStorage = make(map[string]map[string][]byte)

func testClientFunc(testName string) func(ctx context.Context) Storage {
	if testStorage[testName] == nil {
		testStorage[testName] = make(map[string][]byte)
	}
	data := testStorage[testName]
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
