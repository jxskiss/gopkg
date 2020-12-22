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

func GetType(name string) reflect.Type {
	sections, offsets := typelinks()
	for i, base := range sections {

	LoopOffset:
		for _, offset := range offsets[i] {
			_type := resolveTypeOff(base, offset)
			typ := *(*reflect.Type)(unsafe.Pointer(&iface{
				tab:  _itab_reflectType,
				data: _type,
			}))
			for typ.Name() == "" {
				if typ.Kind() != reflect.Ptr {
					continue LoopOffset
				}
				typ = typ.Elem()
			}
			pkgPath := removeVendorPrefix(typ.PkgPath())
			if !strings.HasPrefix(name, pkgPath) {
				continue
			}
			if name == pkgPath+"."+typ.Name() {
				return typ
			}
		}
	}
	panic(fmt.Sprintf("forceexprt: cannot find type %s, may be inactive", name))
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
