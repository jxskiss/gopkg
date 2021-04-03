package reflectx

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestToRType(t *testing.T) {
	var x int64
	rtyp1 := RTypeOf(x)
	rtyp2 := ToRType(reflect.TypeOf(x))
	assert.Equal(t, rtyp1, rtyp2)
}

func TestRTypeOf(t *testing.T) {
	var x int64
	typ1 := RTypeOf(x).ReflectType()
	typ2 := reflect.TypeOf(x)
	assert.Equal(t, typ1, typ2)
}

func TestPtrTo(t *testing.T) {
	var x int64
	typ1 := PtrTo(RTypeOf(x)).ReflectType()
	typ2 := reflect.PtrTo(reflect.TypeOf(x))
	assert.Equal(t, typ1, typ2)
}
