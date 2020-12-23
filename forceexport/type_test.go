package forceexport

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type TestStruct struct {
	// pass
}

func TestRuntimeModuledata(t *testing.T) {
	var rtmdtype reflect.Type
	assert.NotPanics(t, func() {
		rtmdtype = GetType("runtime.moduledata")
	})
	assert.Equal(t, reflect.Struct, rtmdtype.Kind())

	nextfield, ok := rtmdtype.FieldByName("next")
	assert.True(t, ok)
	assert.Equal(t, reflect.Ptr, nextfield.Type.Kind())
}

func TestTypeEquality(t *testing.T) {
	ifacetype := GetType("github.com/jxskiss/gopkg/forceexport.iface")
	assert.Equal(t, reflect.TypeOf(iface{}), ifacetype)
	
	structtype := GetType("github.com/jxskiss/gopkg/forceexport.TestStruct")
	assert.Equal(t, reflect.TypeOf(TestStruct{}), structtype)
}
