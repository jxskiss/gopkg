package forceexport

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

// GetType gets the type defined by the given fully-qualified name.
// If the specified type does not exist, or inactive (haven't been
// compiled into the binary), it panics.
func GetType(name string) *reflectx.RType {
	sections, offsets := linkname.Reflect_typelinks()
	for i, base := range sections {
		for _, offset := range offsets[i] {
			typ := (*reflectx.RType)(linkname.Reflect_resolveTypeOff(base, offset))
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
	panic(fmt.Sprintf("forceexport: cannot find type %s, maybe inactive", name))
}

// ScanType scans type information which are available from reflect.typelinks.
// For each type, it calls f with the type's fully-qualified name and type.
func ScanType(f func(name string, typ *reflectx.RType)) {
	sections, offsets := linkname.Reflect_typelinks()
	for i, base := range sections {
		for _, offset := range offsets[i] {
			typ := (*reflectx.RType)(linkname.Reflect_resolveTypeOff(base, offset))
			for typ.Name() == "" && typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			typName := typ.Name()
			if typName == "" {
				continue
			}
			pkgPath := removeVendorPrefix(typ.PkgPath())
			if pkgPath == "" {
				continue
			}
			fullName := fmt.Sprintf("%s.%s", pkgPath, typName)
			f(fullName, typ)
		}
	}
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
