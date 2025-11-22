package sqlutil

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
	"github.com/jxskiss/gopkg/v2/internal/testpb"
)

func TestTypes(t *testing.T) {
	var testcases = []struct {
		a driver.Valuer
		b sql.Scanner
	}{
		{
			a: &JSONMap{Map: ezmap.Map{"a": "1", "b": float64(2), "c": []any{"a", "b", "c"}}},
			b: &JSONMap{},
		},
	}

	for _, tc := range testcases {
		buf, err := tc.a.Value()
		assert.Nil(t, err)
		assert.NotNil(t, buf)

		err = tc.b.Scan(buf)
		assert.Nil(t, err)
		assert.Equal(t, tc.a.(*JSONMap).GetString("a"), "1")
		assert.Equal(t, tc.a.(*JSONMap).GetFloat("b"), float64(2))
		assert.Equal(t, tc.a.(*JSONMap).MustGet("c"), tc.b.(*JSONMap).MustGet("c"))
	}
}

type Record struct {
	ID      int64      `json:"id" db:"db"`           // bigint unsigned not null primary_key
	Column1 string     `json:"column1" db:"column1"` // varchar(32) not null
	Extra   LazyBinary `json:"extra" db:"extra"`     // blob
}

func (p *Record) GetExtra() (map[string]any, error) {
	unmarshaler := func(b []byte) (any, error) {
		var out = map[string]any{}
		var err error
		if len(b) > 0 {
			err = json.Unmarshal(b, &out)
		}
		return out, err
	}

	extra, err := p.Extra.Get(unmarshaler)
	if err != nil {
		return nil, err
	}
	return extra.(map[string]any), nil
}

func (p *Record) SetExtra(extra map[string]any) {
	buf, _ := json.Marshal(extra)
	p.Extra.Set(buf, extra)
}

func TestLazyBinary(t *testing.T) {
	extra := map[string]any{
		"a": "123",
		"b": "234",
		"c": "345",
	}
	extraBuf, _ := json.Marshal(extra)

	row := &Record{}
	assert.Len(t, row.Extra.GetBytes(), 0)
	_, err := row.GetExtra()
	assert.Nil(t, err)

	row.Extra.Set(extraBuf, nil)
	assert.Len(t, row.Extra.GetBytes(), len(extraBuf))

	got1, _ := row.GetExtra()
	got2, _ := row.GetExtra()
	assert.Equal(t, extra, got1)
	assert.Equal(t, extra, got2)

	delete(extra, "a")
	row.SetExtra(extra)
	got3, _ := row.GetExtra()
	assert.Equal(t, extra, got3)

	row.Extra.Set(extraBuf, nil)
	got4, _ := row.GetExtra()
	assert.Len(t, got4, 3)
}

func TestBitmap(t *testing.T) {
	bm := NewBitmap[int](nil)
	assert.False(t, bm.Get(0x1))
	assert.False(t, bm.Get(0x2))
	assert.False(t, bm.Get(0x4))
	assert.Equal(t, 0, bm.Underlying())

	bm.Set(0x1)
	bm.Set(0x4)
	bm.Set(0x8)
	assert.True(t, bm.Get(0x1))
	assert.False(t, bm.Get(0x2))
	assert.True(t, bm.Get(0x4))
	assert.True(t, bm.Get(0x8))
	assert.False(t, bm.Get(0x10))
	assert.Equal(t, 13, bm.Underlying())

	bm.Set(0x10)
	bm.Clear(0x4)
	assert.False(t, bm.Get(0x4))
	bm.Clear(0x9)
	assert.False(t, bm.Get(0x1))
	assert.False(t, bm.Get(0x8))
	assert.Equal(t, 16, bm.Underlying())

	value, err := bm.Value()
	assert.Nil(t, err)
	assert.Equal(t, int64(16), value)
}

func TestJSONMapMarshal(t *testing.T) {
	data := JSONMap{
		Map: map[string]any{
			"a": 123,
		},
	}
	want := `{"a":123}`
	got, err := data.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, want, string(got))

	data = JSONMap{}
	want = "null"
	got, err = data.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, want, string(got))
}

func TestJSONMapUnmarshal(t *testing.T) {
	data := []byte(`{"a":123}`)
	obj := JSONMap{}
	err := obj.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, ezmap.Map{"a": 123.0}, obj.Map)

	data = []byte("null")
	obj = JSONMap{}
	err = obj.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Nil(t, obj.Map)
}

func TestLazyJSON(t *testing.T) {
	type Extra1 struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}
	type Record1 struct {
		ID      int64            `json:"id" db:"db"`
		Column1 string           `json:"column1" db:"column1"`
		Extra   LazyJSON[Extra1] `json:"extra" db:"extra"`
	}

	row1 := &Record1{}
	err := row1.Extra.Scan([]byte(`{"field1":"abc","field2":123}`))
	assert.Nil(t, err)
	extra1, err := row1.Extra.Get()
	assert.Nil(t, err)
	assert.Equal(t, Extra1{Field1: "abc", Field2: 123}, *extra1)
	extra2, err := row1.Extra.Get()
	assert.Nil(t, err)
	assert.Equal(t, extra1, extra2) // same object

	value, err := row1.Extra.Value()
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"field1":"abc","field2":123}`), value)

	err = row1.Extra.Set(&Extra1{Field1: "def", Field2: 456})
	assert.Nil(t, err)
	extra3, err := row1.Extra.Get()
	assert.Nil(t, err)
	assert.Equal(t, Extra1{Field1: "def", Field2: 456}, *extra3)

	buf1 := row1.Extra.GetBytes()
	assert.Equal(t, []byte(`{"field1":"def","field2":456}`), buf1)

	buf2, err := json.Marshal(row1.Extra)
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"field1":"def","field2":456}`), buf2)

	err = json.Unmarshal([]byte("null"), &row1.Extra)
	assert.Nil(t, err)
	extra4, err := row1.Extra.Get()
	assert.Nil(t, err)
	assert.Nil(t, extra4)
}

func TestLasyProtobuf(t *testing.T) {
	type Record1 struct {
		ID      int64                              `json:"id" db:"db"`
		Column1 string                             `json:"column1" db:"column1"`
		Extra   LazyProtobuf[*testpb.ShardingData] `json:"extra" db:"extra"`
	}

	obj := &testpb.ShardingData{
		TotalNum: 100,
		ShardNum: 10,
		Digest:   []byte("123456"),
		Data:     []byte("abcdef"),
	}
	objBuf, err := proto.Marshal(obj)
	assert.Nil(t, err)

	row1 := &Record1{}
	err = row1.Extra.Scan(objBuf)
	assert.Nil(t, err)
	extra1, err := row1.Extra.Get()
	assert.Nil(t, err)
	assert.Equal(t, obj.TotalNum, extra1.TotalNum)
	assert.Equal(t, obj.ShardNum, extra1.ShardNum)
	assert.Equal(t, obj.Digest, extra1.Digest)
	assert.Equal(t, obj.Data, extra1.Data)

	value, err := row1.Extra.Value()
	assert.Nil(t, err)
	assert.Equal(t, objBuf, value)

	row1.Extra.SetBytes(objBuf)
	assert.Nil(t, atomic.LoadPointer(&row1.Extra.lb.obj))

	err = row1.Extra.Set(obj)
	assert.Nil(t, err)
	extra2, err := row1.Extra.Get()
	assert.Nil(t, err)
	assert.Equal(t, obj, extra2)

	err = row1.Extra.Set(nil)
	assert.Nil(t, err)
	extra3, err := row1.Extra.Get()
	assert.Nil(t, err)
	assert.Nil(t, extra3)
	assert.Empty(t, row1.Extra.GetBytes())
}
