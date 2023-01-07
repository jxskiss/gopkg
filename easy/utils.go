package easy

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
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

// SetNonZeroValue sets to dst only when val is not zero value of it's type.
func SetNonZeroValue[T comparable](dst **T, val T) {
	var zero T
	if val != zero {
		*dst = &val
	}
}

// Caller returns function name, filename, and the line number of the caller.
// The argument skip is the number of stack frames to ascend, with 0
// identifying the caller of Caller.
func Caller(skip int) (name, file string, line int) {
	pc, file, line, _ := runtime.Caller(skip + 1)
	name = runtime.FuncForPC(pc).Name()
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			name = name[i+1:]
			break
		}
	}
	pathSepCnt := 0
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			pathSepCnt++
			if pathSepCnt == 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return
}

// CallerName returns the function name of the direct caller.
// This is a convenient wrapper around Caller.
func CallerName() string {
	name, _, _ := Caller(1)
	return name
}

// SingleJoin joins the given text segments using sep.
// No matter whether a segment begins or ends with sep or not, it
// guarantees that only one sep appears between two segments.
func SingleJoin(sep string, text ...string) string {
	if len(text) == 0 {
		return ""
	}
	result := text[0]
	for _, next := range text[1:] {
		asep := strings.HasSuffix(result, sep)
		bsep := strings.HasPrefix(next, sep)
		switch {
		case asep && bsep:
			result += next[len(sep):]
		case !asep && !bsep:
			result += sep + next
		default:
			result += next
		}
	}
	return result
}

// SlashJoin joins the given path segments using "/".
// No matter whether a segment begins or ends with "/" or not, it guarantees
// that only one "/" appears between two segments.
func SlashJoin(path ...string) string {
	return SingleJoin("/", path...)
}
