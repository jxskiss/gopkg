package kvutil

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/jxskiss/gopkg/lru"
	"github.com/jxskiss/gopkg/serialize"
	"github.com/stretchr/testify/assert"
)

var (
	dummyData     = serialize.Int32List{1, 3, 5, 7, 9}
	testModelList = []*TestModel{
		{
			IntId: 111,
			StrId: "aaa",
			Dummy: dummyData,
		},
		{
			IntId: 112,
			StrId: "aab",
			Dummy: dummyData,
		},
	}
	testModelMapInt = map[int64]*TestModel{
		113: {
			IntId: 113,
			StrId: "aac",
			Dummy: dummyData,
		},
	}
	testModelMapStr = map[string]*TestModel{
		"aac": {
			IntId: 113,
			StrId: "aac",
			Dummy: dummyData,
		},
	}
	testIntIds       = []int64{111, 112, 113}
	testStrIds       = []string{"aaa", "aab", "aac"}
	testDeleteIntIds = []int64{111, 112}
	testDeleteStrIds = []string{"aab", "aac"}
)

func TestModelCacheIntId(t *testing.T) {
	mc := makeModelCache("TestModelCacheIntId",
		func(i interface{}) interface{} {
			return i.(*TestModel).IntId
		})

	ctx := context.Background()
	var modelList []*TestModel
	var modelMap = make(map[int64]*TestModel)
	var err error

	err = mc.MGetByIntKeys(ctx, &modelList, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	err = mc.MGetByIntKeys(ctx, &modelMap, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 0)

	// we can populate cache using either a list or a map
	err = mc.MSet(ctx, testModelList)
	assert.Nil(t, err)
	err = mc.MSet(ctx, testModelMapInt)
	assert.Nil(t, err)

	err = mc.MGetByIntKeys(ctx, &modelList, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 3)
	assert.Equal(t, dummyData, modelList[0].Dummy)

	err = mc.MGetByIntKeys(ctx, &modelMap, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 3)

	err = mc.MDeleteByIntKeys(ctx, testDeleteIntIds)
	assert.Nil(t, err)
	assert.Len(t, mc.ClientFunc(ctx).(*memoryStorage).data, 1)
}

func TestModelCacheStringId(t *testing.T) {
	mc := makeModelCache("TestModelCacheStringId",
		func(i interface{}) interface{} {
			return i.(*TestModel).StrId
		})

	ctx := context.Background()
	var modelList []*TestModel
	var modelMap = make(map[string]*TestModel)
	var err error

	err = mc.MGetByStringKeys(ctx, &modelList, testStrIds)
	assert.Nil(t, err)

	err = mc.MGetByStringKeys(ctx, &modelMap, testStrIds)
	assert.Nil(t, err)

	// we can populate cache using either a list or a map
	err = mc.MSet(ctx, testModelList)
	assert.Nil(t, err)
	err = mc.MSet(ctx, testModelMapStr)
	assert.Nil(t, err)

	err = mc.MGetByStringKeys(ctx, &modelList, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 3)
	assert.Equal(t, dummyData, modelList[0].Dummy)

	err = mc.MGetByStringKeys(ctx, &modelMap, testStrIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 3)

	err = mc.MDeleteByStringKeys(ctx, testDeleteStrIds)
	assert.Nil(t, err)
	assert.Len(t, mc.ClientFunc(ctx).(*memoryStorage).data, 1)
}

func TestModelCacheWithLruCache(t *testing.T) {
	mc := makeModelCache("TestModelCacheWithLruCache",
		func(i interface{}) interface{} {
			return i.(*TestModel).IntId
		})
	mc.LruCache = lru.NewCache(10)
	mc.LruExpiration = time.Minute

	ctx := context.Background()
	var modelList []*TestModel
	var modelMap = make(map[int64]*TestModel)
	var err error

	err = mc.MGetByIntKeys(ctx, &modelList, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 0)

	err = mc.MSet(ctx, testModelList)
	assert.Len(t, mc.ClientFunc(ctx).(*memoryStorage).data, 2)

	got1, exists1, expired1 := mc.LruCache.Get(mc.KeyFunc(111))
	assert.NotNil(t, got1)
	assert.True(t, exists1 && !expired1)

	got2, exists2, expired2 := mc.LruCache.Get(mc.KeyFunc(112))
	assert.NotNil(t, got2)
	assert.True(t, exists2 && !expired2)

	got3, exists3, expired3 := mc.LruCache.Get(mc.KeyFunc(113))
	assert.Nil(t, got3)
	assert.False(t, exists3 || expired3)

	err = mc.MGetByIntKeys(ctx, &modelList, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelList, 2)

	err = mc.MGetByIntKeys(ctx, &modelMap, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, modelMap, 2)

	err = mc.MDeleteByIntKeys(ctx, testIntIds)
	assert.Nil(t, err)
	assert.Len(t, mc.ClientFunc(ctx).(*memoryStorage).data, 0)
	got1, exists1, expired1 = mc.LruCache.Get(mc.KeyFunc(111))
	assert.Nil(t, got1)
	assert.False(t, exists1 || expired1)
}

func makeModelCache(testName string, idFunc func(i interface{}) interface{}) *ModelCache {
	km := KeyManager{}
	return &ModelCache{
		ClientFunc:    testClientFunc(testName),
		IdFunc:        idFunc,
		KeyFunc:       km.NewKey("test_model:{id}"),
		LruCache:      nil,
		MGetBatchSize: 2,
		MSetBatchSize: 2,
		Expiration:    0,
	}
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
	Dummy serialize.Int32List
}

func (t *TestModel) MarshalModel() ([]byte, error) {
	var buf []byte
	buf = append(buf, []byte(strconv.FormatInt(t.IntId, 10))...)
	buf = append(buf, []byte(t.StrId)...)
	pb, err := t.Dummy.MarshalProto()
	if err != nil {
		return nil, err
	}
	buf = append(buf, pb...)
	return buf, nil
}

func (t *TestModel) UnmarshalModel(b []byte) error {
	t.IntId, _ = strconv.ParseInt(string(b[:3]), 10, 64)
	t.StrId = string(b[3:6])
	return t.Dummy.UnmarshalProto(b[6:])
}
