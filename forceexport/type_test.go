package forceexport

import (
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/reflectx"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	// pass
}

func TestScanType(t *testing.T) {
	got := make([]string, 0)
	ScanType(func(name string, typ *reflectx.RType) {
		if strings.HasPrefix(name, "github.com/jxskiss/gopkg") {
			got = append(got, name)
		}
	})
	assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/forceexport.iface")
	assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/forceexport.moduledata")
	assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/forceexport.TestStruct")
	assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/internal/rtype.RType")
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
	ifacetype := GetType("github.com/jxskiss/gopkg/v2/forceexport.iface").ToType()
	assert.Equal(t, reflect.TypeOf(iface{}), ifacetype)

	structtype := GetType("github.com/jxskiss/gopkg/v2/forceexport.TestStruct").ToType()
	assert.Equal(t, reflect.TypeOf(TestStruct{}), structtype)
}

// iface is a copy type of runtime.iface.
type iface struct {
	tab  unsafe.Pointer // *itab
	data unsafe.Pointer
}
