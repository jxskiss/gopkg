package kvutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/jxskiss/gopkg/v2/easy"
	"github.com/jxskiss/gopkg/v2/exp/kvutil/sharding_protobuf_test"
)

type TestProtoShardingModel struct {
	Entity *sharding_protobuf_test.TestShardingModel
}

func (m *TestProtoShardingModel) MarshalBinary() (data []byte, err error) {
	return proto.Marshal(m.Entity)
}

func (m *TestProtoShardingModel) UnmarshalBinary(data []byte) error {
	entity := &sharding_protobuf_test.TestShardingModel{}
	err := proto.Unmarshal(data, entity)
	if err != nil {
		return err
	}
	m.Entity = entity
	return nil
}

func (m *TestProtoShardingModel) GetShardingData() (ShardingData, bool) {
	if m.Entity == nil || m.Entity.ShardData == nil {
		return ShardingData{}, false
	}
	shardData := ShardingData{
		TotalNum: m.Entity.ShardData.TotalNum,
		ShardNum: m.Entity.ShardData.ShardNum,
		Digest:   m.Entity.ShardData.Digest,
		Data:     m.Entity.ShardData.Data,
	}
	return shardData, true
}

func (m *TestProtoShardingModel) SetShardingData(data ShardingData) {
	if m.Entity == nil {
		m.Entity = &sharding_protobuf_test.TestShardingModel{}
	}
	m.Entity.ShardData = &sharding_protobuf_test.ShardingData{
		TotalNum: data.TotalNum,
		ShardNum: data.ShardNum,
		Digest:   data.Digest,
		Data:     data.Data,
	}
}

var (
	testProtobufShardingModelList = []*TestProtoShardingModel{
		{
			Entity: &sharding_protobuf_test.TestShardingModel{
				Id:      111,
				BizData: []byte("test"),
			},
		},
		{
			Entity: &sharding_protobuf_test.TestShardingModel{
				Id:      112,
				BizData: easy.Repeat([]byte("test "), 10),
			},
		},
		{
			Entity: &sharding_protobuf_test.TestShardingModel{
				Id:      113,
				BizData: easy.Repeat([]byte("test "), 50),
			},
		},
	}
)

//nolint:dupl
func TestShardingCache_Protobuf(t *testing.T) {
	kf := KeyFactory{}
	cfg := &ShardingCacheConfig[int64, *TestProtoShardingModel]{
		Storage: testClientFunc("testShardingCache_Protobuf"),
		IDFunc: func(model *TestProtoShardingModel) int64 {
			return model.Entity.GetId()
		},
		KeyFunc:         kf.NewKey("testShardingCache:{id}"),
		ShardingSize:    10,
		MGetBatchSize:   2,
		MSetBatchSize:   2,
		DeleteBatchSize: 2,
	}

	ctx := context.Background()
	sc := NewShardingCache[int64, *TestProtoShardingModel](cfg)

	t.Run("Get / not found", func(t *testing.T) {
		gotModel, err := sc.Get(ctx, testIntIds[0])
		assert.Equal(t, ErrDataNotFound, err)
		assert.Nil(t, gotModel)
	})

	t.Run("MGet / not found", func(t *testing.T) {
		modelMap, errMap, err := sc.MGet(ctx, testIntIds)
		assert.Nil(t, err)
		assert.Len(t, errMap, 0)
		assert.Len(t, modelMap, 0)
	})

	t.Run("Set", func(t *testing.T) {
		clearMemoryStorage(ctx, sc.config.Storage)
		stor := getMemoryStorage(ctx, sc.config.Storage)
		_ = stor

		err1 := sc.Set(ctx, 111, testProtobufShardingModelList[0], 0)
		assert.Nil(t, err1)

		err2 := sc.Set(ctx, 112, testProtobufShardingModelList[1], 0)
		assert.Nil(t, err2)

		err3 := sc.Set(ctx, 113, testProtobufShardingModelList[2], 0)
		assert.Nil(t, err3)

		got1, err1 := sc.Get(ctx, 111)
		assert.Nil(t, err1)
		assert.NotNil(t, got1)
		assert.Nil(t, got1.Entity.ShardData)
		assert.Equal(t, testProtobufShardingModelList[0].Entity.BizData, got1.Entity.BizData)

		got2, err2 := sc.Get(ctx, 112)
		assert.Nil(t, err2)
		assert.NotNil(t, got2)
		assert.Nil(t, got2.Entity.ShardData)
		assert.Equal(t, testProtobufShardingModelList[1].Entity.BizData, got2.Entity.BizData)

		got3, err3 := sc.Get(ctx, 113)
		assert.Nil(t, err3)
		assert.NotNil(t, got3)
		assert.Nil(t, got3.Entity.ShardData)
		assert.Equal(t, testProtobufShardingModelList[2].Entity.BizData, got3.Entity.BizData)

		mgetRet, errMap, err := sc.MGet(ctx, []int64{111, 112, 113, 114})
		assert.Nil(t, err)
		assert.Len(t, errMap, 0)
		assert.Len(t, mgetRet, 3)
		assert.Equal(t, testProtobufShardingModelList[0].Entity.BizData, mgetRet[111].Entity.BizData)
		assert.Equal(t, testProtobufShardingModelList[1].Entity.BizData, mgetRet[112].Entity.BizData)
		assert.Equal(t, testProtobufShardingModelList[2].Entity.BizData, mgetRet[113].Entity.BizData)
	})

	t.Run("MSet", func(t *testing.T) {
		clearMemoryStorage(ctx, sc.config.Storage)
		stor := getMemoryStorage(ctx, sc.config.Storage)
		_ = stor

		err := sc.MSet(ctx, testProtobufShardingModelList, 0)
		require.Nil(t, err)

		mgetRet, errMap, err := sc.MGet(ctx, []int64{111, 112, 113, 114})
		assert.Nil(t, err)
		assert.Len(t, errMap, 0)
		assert.Len(t, mgetRet, 3)
		assert.Equal(t, testProtobufShardingModelList[0].Entity.BizData, mgetRet[111].Entity.BizData)
		assert.Equal(t, testProtobufShardingModelList[1].Entity.BizData, mgetRet[112].Entity.BizData)
		assert.Equal(t, testProtobufShardingModelList[2].Entity.BizData, mgetRet[113].Entity.BizData)
	})

	t.Run("Delete", func(t *testing.T) {
		clearMemoryStorage(ctx, sc.config.Storage)
		stor := getMemoryStorage(ctx, sc.config.Storage)
		_ = stor

		err := sc.MSet(ctx, testProtobufShardingModelList, 0)
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
