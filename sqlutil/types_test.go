package sqlutil

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/jxskiss/gopkg/gemap"
	"github.com/jxskiss/gopkg/serialize"
	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	var testcases = []struct {
		a driver.Valuer
		b sql.Scanner
	}{
		{
			a: &JSON{Map: gemap.Map{"a": "1", "b": float64(2), "c": []interface{}{"a", "b", "c"}}},
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
		assert.Equal(t, tc.a.(*JSON).GetFloat64("b"), float64(2))
		assert.Equal(t, tc.a.(*JSON).MustGet("c"), tc.b.(*JSON).MustGet("c"))
	}
}

type Record struct {
	ID      int64      `json:"id" db:"db"`           // bigint unsigned not null primary_key
	Column1 string     `json:"column1" db:"column1"` // varchar(32) not null
	Extra   LazyBinary `json:"extra" db:"extra"`     // blob
}

func (p *Record) GetExtra() (map[string]string, error) {
	unmarshaler := func(b []byte) (interface{}, error) {
		var out = map[string]string{}
		var err error
		if len(b) > 0 {
			err = (*serialize.StringMap)(&out).UnmarshalProto(b)
		}
		return out, err
	}

	extra, err := p.Extra.Get(unmarshaler)
	if err != nil {
		return nil, err
	}
	return extra.(map[string]string), nil
}

func (p *Record) SetExtra(extra map[string]string) {
	buf, _ := serialize.StringMap(extra).MarshalProto()
	p.Extra.Set(buf, extra)
}

func TestLazyBinary(t *testing.T) {
	extra := map[string]string{
		"a": "123",
		"b": "234",
		"c": "345",
	}
	extraBuf, _ := serialize.StringMap(extra).MarshalProto()

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

func TestJSONMarshal(t *testing.T) {
	data := JSON{
		Map: map[string]interface{}{
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
	assert.Equal(t, gemap.Map{"a": 123.0}, obj.Map)

	data = []byte("null")
	obj = JSON{}
	err = obj.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Nil(t, obj.Map)
}
