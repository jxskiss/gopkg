package easy

import (
	"github.com/jxskiss/gopkg/reflectx"
	"reflect"
)

// SetDefault checks whether dst points to a zero value, if yes, it sets
// the first non-zero value to dst.
// dst must be a pointer to same type as value, else it panics.
func SetDefault(dst interface{}, value ...interface{}) {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || !reflect.Indirect(dstVal).IsValid() {
		panic("SetDefault: dst must be a non-nil pointer")
	}
	if reflect.Indirect(dstVal).IsZero() {
		kind := dstVal.Elem().Kind()
		for _, x := range value {
			xval := reflect.ValueOf(x)
			if !xval.IsZero() {
				switch kind {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					dstVal.Elem().SetInt(reflectx.ReflectInt(xval))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					dstVal.Elem().SetUint(uint64(reflectx.ReflectInt(xval)))
				case reflect.Float32, reflect.Float64:
					dstVal.Elem().SetFloat(xval.Float())
				default:
					dstVal.Elem().Set(xval)
				}
				break
			}
		}
	}
}
