package reflectx

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type simple struct {
	A string
}

func TestRType(t *testing.T) {
	var (
		oneI8    int8  = 1
		twoI32   int32 = 2
		threeInt int   = 3
		strA           = "a"
	)
	types := []interface{}{
		int8(1),
		&oneI8,
		int32(2),
		&twoI32,
		int(3),
		&threeInt,
		"a",
		&strA,
		simple{"b"},
		&simple{"b"},
	}
	for _, x := range types {
		rtype1 := EfaceOf(&x).RType
		rtype2 := RTypeOf(reflect.TypeOf(x))
		rtype3 := RTypeOf(reflect.ValueOf(x))
		assert.Equal(t, rtype1, rtype2)
		assert.Equal(t, rtype2, rtype3)
	}
}

func TestToRType(t *testing.T) {
	var x int64
	rtyp1 := RTypeOf(x)
	rtyp2 := ToRType(reflect.TypeOf(x))
	assert.Equal(t, rtyp1, rtyp2)
}

func TestRTypeOf(t *testing.T) {
	var x int64
	typ1 := RTypeOf(x).ToType()
	typ2 := reflect.TypeOf(x)
	assert.Equal(t, typ1, typ2)
}

func TestPtrTo(t *testing.T) {
	var x int64
	typ1 := PtrTo(RTypeOf(x)).ToType()
	typ2 := reflect.PtrTo(reflect.TypeOf(x))
	assert.Equal(t, typ1, typ2)
}
