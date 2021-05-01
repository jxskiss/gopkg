package reflectx

import (
	"fmt"
	"github.com/jxskiss/gopkg/internal/linkname"
	"reflect"
	"strings"
)

// GetType gets the type defined by the given fully-qualified name.
// If the specified type does not exist, or inactive (haven't been
// compiled into the binary), it panics.
func GetType(name string) *RType {
	sections, offsets := linkname.Reflect_typelinks()
	for i, base := range sections {
		for _, offset := range offsets[i] {
			typ := (*RType)(linkname.Reflect_resolveTypeOff(base, offset))
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
	panic(fmt.Sprintf("reflectx: cannot find type %s, maybe inactive", name))
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
