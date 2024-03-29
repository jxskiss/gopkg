package sqlutil

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

func TestTypes(t *testing.T) {
	var testcases = []struct {
		a driver.Valuer
		b sql.Scanner
	}{
		{
			a: &JSON{Map: ezmap.Map{"a": "1", "b": float64(2), "c": []any{"a", "b", "c"}}},
			b: &JSON{},
		},
	}

	for _, tc := range testcases {
		buf, err := tc.a.Value()
		assert.Nil(t, err)
		assert.NotNil(t, buf)

		err = tc.b.Scan(buf)
		assert.Nil(t, err)
		assert.Equal(t, tc.a.(*JSON).GetString("a"), "1")
		assert.Equal(t, tc.a.(*JSON).GetFloat("b"), float64(2))
		assert.Equal(t, tc.a.(*JSON).MustGet("c"), tc.b.(*JSON).MustGet("c"))
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

func TestJSONMarshal(t *testing.T) {
	data := JSON{
		Map: map[string]any{
			"a": 123,
		},
	}
	want := `{"a":123}`
	got, err := data.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, want, string(got))

	data = JSON{}
	want = "null"
	got, err = data.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, want, string(got))
}

func TestJSONUnmarshal(t *testing.T) {
	data := []byte(`{"a":123}`)
	obj := JSON{}
	err := obj.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, ezmap.Map{"a": 123.0}, obj.Map)

	data = []byte("null")
	obj = JSON{}
	err = obj.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Nil(t, obj.Map)
}
