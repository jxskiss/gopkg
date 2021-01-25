package sqlutil

import (
	"database/sql"
	"database/sql/driver"
	"github.com/jxskiss/gopkg/serialize"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestTypes(t *testing.T) {
	var testcases = []struct {
		a driver.Valuer
		b sql.Scanner
	}{
		{
			a: JSONInt32s{1, 2, 3, 438138181},
			b: &JSONInt32s{},
		},
		{
			a: JSONInt64s{1, 2, 3, 38828198418419919},
			b: &JSONInt64s{},
		},
		{
			a: JSONStrings{"a", "b", "c"},
			b: &JSONStrings{},
		},
		{
			a: JSONStringMap{"a": "123", "b": "234", "c": "345"},
			b: &JSONStringMap{},
		},
		{
			a: JSONDict{"a": "1", "b": float64(2), "c": []interface{}{"a", "b", "c"}},
			b: &JSONDict{},
		},
		{
			a: PBInt32s{1, 3, 48483219},
			b: &PBInt32s{},
		},
		{
			a: PBInt64s{1, 3, 431943119414314},
			b: &PBInt64s{},
		},
		{
			a: PBStrings{"a", "b", "c"},
			b: &PBStrings{},
		},
		{
			a: PBStringMap{"a": "123", "b": "234", "c": "345"},
			b: &PBStringMap{},
		},
	}

	for _, tc := range testcases {
		buf, err := tc.a.Value()
		assert.Nil(t, err)
		assert.NotNil(t, buf)

		err = tc.b.Scan(buf)
		assert.Nil(t, err)
		assert.Equal(t, tc.a, reflect.ValueOf(tc.b).Elem().Interface())
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
