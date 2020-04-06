package reflectx

import (
	"github.com/jxskiss/gopkg/ptr"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type simple struct {
	A string
}

func Test_rtype(t *testing.T) {
	types := []interface{}{
		int8(1),
		ptr.Int8(1),
		int32(2),
		ptr.Int32(2),
		int(3),
		ptr.Int(3),
		"a",
		ptr.String("a"),
		simple{"b"},
		&simple{"b"},
	}
	for _, x := range types {
		rtype1 := EFaceOf(&x).RType
		rtype2 := RTypeOf(reflect.TypeOf(x))
		rtype3 := RTypeOf(reflect.ValueOf(x))
		assert.Equal(t, rtype1, rtype2)
		assert.Equal(t, rtype2, rtype3)
	}
}
