package forceexport

import (
	"github.com/jxskiss/gopkg/reflectx"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"unsafe"
)

type TestStruct struct {
	// pass
}

func TestRuntimeModuledata(t *testing.T) {
	var rtmdtype *reflectx.RType
	assert.NotPanics(t, func() {
		rtmdtype = GetType("runtime.moduledata")
	})
	assert.Equal(t, reflect.Struct, rtmdtype.Kind())

	nextfield, ok := rtmdtype.FieldByName("next")
	assert.True(t, ok)
	assert.Equal(t, reflect.Ptr, nextfield.Type.Kind())
}

func TestTypeEquality(t *testing.T) {
	ifacetype := GetType("github.com/jxskiss/gopkg/forceexport.iface").ToType()
	assert.Equal(t, reflect.TypeOf(iface{}), ifacetype)

	structtype := GetType("github.com/jxskiss/gopkg/forceexport.TestStruct").ToType()
	assert.Equal(t, reflect.TypeOf(TestStruct{}), structtype)
}

// iface is a copy type of runtime.iface.
type iface struct {
	tab  unsafe.Pointer // *itab
	data unsafe.Pointer
}
