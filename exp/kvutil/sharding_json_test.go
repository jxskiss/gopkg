package kvutil

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestJSONShardData struct {
	TotalNum int32  `json:"totalNum"`
	ShardNum int32  `json:"shardNum"`
	Digest   []byte `json:"digest"`

	// Data ...
	// As an example, JSON data is always valid UTF-8 strings,
	// string and []byte are exchangeable here,
	// using string here avoids unnecessary base64 encoding and decoding
	// when doing JSON serialization.
	Data string `json:"data"`
}

type TestJSONShardingModel struct {
	ID        int64  `json:"id,omitempty"`
	InnerData string `json:"innerData,omitempty"`

	// ShardData helps to do big value sharding.
	ShardData *TestJSONShardData `json:"shardData"`
}

func (m *TestJSONShardingModel) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m *TestJSONShardingModel) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *TestJSONShardingModel) GetShardingData() (ShardingData, bool) {
	if m.ShardData == nil {
		return ShardingData{}, false
	}
	return ShardingData{
		TotalNum: m.ShardData.TotalNum,
		ShardNum: m.ShardData.ShardNum,
		Digest:   m.ShardData.Digest,
		Data:     []byte(m.ShardData.Data),
	}, true
}

func (m *TestJSONShardingModel) SetShardingData(data ShardingData) {
	m.ShardData = &TestJSONShardData{
		TotalNum: data.TotalNum,
		ShardNum: data.ShardNum,
		Digest:   data.Digest,
		Data:     string(data.Data),
	}
}

var (
	testJSONShardingModelList = []*TestJSONShardingModel{
		{
			ID:        111,
			InnerData: "test",
		},
		{
			ID:        112,
			InnerData: strings.Repeat("test ", 10),
		},
		{
			ID:        113,
			InnerData: strings.Repeat("test ", 50),
		},
	}
)

//nolint:dupl
func TestShardingCache_JSON(t *testing.T) {
	kf := KeyFactory{}
	cfg := &ShardingCacheConfig[int64, *TestJSONShardingModel]{
		Storage: testClientFunc("testShardingCache_JSON"),
		IDFunc: func(model *TestJSONShardingModel) int64 {
			return model.ID
		},
		KeyFunc:         kf.NewKey("testShardingCache:{id}"),
		ShardingSize:    50,
		MGetBatchSize:   2,
		MSetBatchSize:   2,
		DeleteBatchSize: 2,
	}

	ctx := context.Background()
	sc := NewShardingCache[int64, *TestJSONShardingModel](cfg)

	t.Run("Get / not found", func(t *testing.T) {
		gotModel, err := sc.Get(ctx, testIntIds[0])
		assert.Equal(t, ErrDataNotFound, err)
		assert.Nil(t, gotModel)
	})

	t.Run("MGetMap / not found", func(t *testing.T) {
		modelMap, errMap, err := sc.MGetSlice(ctx, testIntIds)
		assert.Nil(t, err)
		assert.Len(t, errMap, 0)
		assert.Len(t, modelMap, 0)
	})

	t.Run("Set", func(t *testing.T) {
		clearMemoryStorage(ctx, sc.config.Storage)
		stor := getMemoryStorage(ctx, sc.config.Storage)
		_ = stor

		err1 := sc.Set(ctx, 111, testJSONShardingModelList[0], 0)
		assert.Nil(t, err1)

		err2 := sc.Set(ctx, 112, testJSONShardingModelList[1], 0)
		assert.Nil(t, err2)

		err3 := sc.Set(ctx, 113, testJSONShardingModelList[2], 0)
		assert.Nil(t, err3)

		got1, err1 := sc.Get(ctx, 111)
		assert.Nil(t, err1)
		assert.NotNil(t, got1)
		assert.Nil(t, got1.ShardData)
		assert.Equal(t, testJSONShardingModelList[0].InnerData, got1.InnerData)

		got2, err2 := sc.Get(ctx, 112)
		assert.Nil(t, err2)
		assert.NotNil(t, got2)
		assert.Nil(t, got2.ShardData)
		assert.Equal(t, testJSONShardingModelList[1].InnerData, got2.InnerData)

		got3, err3 := sc.Get(ctx, 113)
		assert.Nil(t, err3)
		assert.NotNil(t, got3)
		assert.Nil(t, got3.ShardData)
		assert.Equal(t, testJSONShardingModelList[2].InnerData, got3.InnerData)

		mgetRet, errMap, err := sc.MGetMap(ctx, []int64{111, 112, 113, 114})
		assert.Nil(t, err)
		assert.Len(t, errMap, 0)
		assert.Len(t, mgetRet, 3)
		assert.Equal(t, testJSONShardingModelList[0].InnerData, mgetRet[111].InnerData)
		assert.Equal(t, testJSONShardingModelList[1].InnerData, mgetRet[112].InnerData)
		assert.Equal(t, testJSONShardingModelList[2].InnerData, mgetRet[113].InnerData)
	})

	t.Run("MSetSlice", func(t *testing.T) {
		clearMemoryStorage(ctx, sc.config.Storage)
		stor := getMemoryStorage(ctx, sc.config.Storage)
		_ = stor

		err := sc.MSetSlice(ctx, testJSONShardingModelList, 0)
		require.Nil(t, err)

		mgetRet, errMap, err := sc.MGetMap(ctx, []int64{111, 112, 113, 114})
		assert.Nil(t, err)
		assert.Len(t, errMap, 0)
		assert.Len(t, mgetRet, 3)
		assert.Equal(t, testJSONShardingModelList[0].InnerData, mgetRet[111].InnerData)
		assert.Equal(t, testJSONShardingModelList[1].InnerData, mgetRet[112].InnerData)
		assert.Equal(t, testJSONShardingModelList[2].InnerData, mgetRet[113].InnerData)
	})

	t.Run("Delete", func(t *testing.T) {
		clearMemoryStorage(ctx, sc.config.Storage)
		stor := getMemoryStorage(ctx, sc.config.Storage)
		_ = stor

		err := sc.MSetSlice(ctx, testJSONShardingModelList, 0)
		require.Nil(t, err)

		err = sc.Delete(ctx, false, 111, 112)
		require.Nil(t, err)

		got1, err1 := sc.Get(ctx, 111)
		assert.Equal(t, ErrDataNotFound, err1)
		assert.Nil(t, got1)

		got2, err2 := sc.Get(ctx, 112)
		assert.Equal(t, ErrDataNotFound, err2)
		assert.Nil(t, got2)

		err = sc.Delete(ctx, true, 113)
		require.Nil(t, err)

		got3, err3 := sc.Get(ctx, 113)
		assert.Equal(t, ErrDataNotFound, err3)
		assert.Nil(t, got3)

		assert.Nil(t, stor.data[sc.config.KeyFunc(111)])
		assert.Nil(t, stor.data[sc.config.KeyFunc(112)])
		assert.NotNil(t, stor.data[GetShardKey(sc.config.KeyFunc(112), 1)])
		assert.Nil(t, stor.data[sc.config.KeyFunc(113)])
		assert.Nil(t, stor.data[GetShardKey(sc.config.KeyFunc(113), 1)])
		assert.Nil(t, stor.data[GetShardKey(sc.config.KeyFunc(113), 2)])
		assert.Nil(t, stor.data[GetShardKey(sc.config.KeyFunc(113), 3)])
	})

}
