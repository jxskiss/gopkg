package forceexport

import (
	"bytes"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

type TestStruct struct {
	// pass
}

func TestScanTypes(t *testing.T) {
	pkgPrefix := []byte("github.com/jxskiss/gopkg")

	t.Run("break iteration", func(t *testing.T) {
		var got []string
		ScanTypes(func(name []byte, typ *reflectx.RType) bool {
			if bytes.HasPrefix(name, pkgPrefix) {
				got = append(got, string(name))
				return false
			}
			return true
		})
		assert.Len(t, got, 1)
	})

	t.Run("iterate all", func(t *testing.T) {
		var got []string
		ScanTypes(func(name []byte, typ *reflectx.RType) bool {
			if bytes.HasPrefix(name, pkgPrefix) {
				got = append(got, string(name))
			}
			return true
		})
		assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/internal/linkname.functab")
		assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/internal/linkname.Runtime_moduledata")
		assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/unsafe/forceexport.iface")
		assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/unsafe/forceexport.TestStruct")
		assert.Contains(t, got, "github.com/jxskiss/gopkg/v2/unsafe/reflectx.RType")
	})
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
	ifacetype := GetType("github.com/jxskiss/gopkg/v2/unsafe/forceexport.iface").ToReflectType()
	assert.Equal(t, reflect.TypeOf(iface{}), ifacetype)

	structtype := GetType("github.com/jxskiss/gopkg/v2/unsafe/forceexport.TestStruct").ToReflectType()
	assert.Equal(t, reflect.TypeOf(TestStruct{}), structtype)
}

// iface is a copy type of runtime.iface.
type iface struct {
	tab  unsafe.Pointer // *itab
	data unsafe.Pointer
}
