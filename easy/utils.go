package easy

import "reflect"

func SetDefault(dst interface{}, value ...interface{}) {
	dstVal := reflect.ValueOf(dst)
	if reflect.Indirect(dstVal).IsZero() {
		kind := dstVal.Elem().Kind()
		for _, x := range value {
			xval := reflect.ValueOf(x)
			if !xval.IsZero() {
				switch kind {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					dstVal.Elem().SetInt(reflectInt(xval))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					dstVal.Elem().SetUint(uint64(reflectInt(xval)))
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
