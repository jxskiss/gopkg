package forceexport

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

var _itab_reflectType = func() unsafe.Pointer {
	typ := reflect.TypeOf(0)
	return (*iface)(unsafe.Pointer(&typ)).tab
}()

// GetType gets the type defined by the given fully-qualified name.
// If the specified type does not exist, or inactive (haven't been
// compiled into the binary), it panics.
func GetType(name string) reflect.Type {
	sections, offsets := typelinks()
	for i, base := range sections {
		for _, offset := range offsets[i] {
			_type := resolveTypeOff(base, offset)
			typ := *(*reflect.Type)(unsafe.Pointer(&iface{
				tab:  _itab_reflectType,
				data: _type,
			}))
			for typ.Name() == "" && typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			typName := typ.Name()
			if typName == "" || !strings.HasSuffix(name, typName) {
				continue
			}
			pkgPath := removeVendorPrefix(typ.PkgPath())
			if name == pkgPath+"."+typName {
				return typ
			}
		}
	}
	panic(fmt.Sprintf("forceexprt: cannot find type %s, maybe inactive", name))
}

func removeVendorPrefix(path string) string {
	const prefix = "/vendor/"
	const prefixLen = 8
	idx := strings.LastIndex(path, prefix)
	if idx >= 0 {
		path = path[idx+prefixLen:]
	}
	return path
}

type iface struct {
	tab  unsafe.Pointer // *itab
	data unsafe.Pointer
}

//go:linkname typelinks reflect.typelinks
func typelinks() ([]unsafe.Pointer, [][]int32)

//go:linkname resolveTypeOff reflect.resolveTypeOff
func resolveTypeOff(_ unsafe.Pointer, _ int32) unsafe.Pointer
