package sqlutil

import (
	"database/sql"
	"database/sql/driver"
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
