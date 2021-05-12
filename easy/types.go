package easy

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/jxskiss/gopkg/reflectx"
	"github.com/jxskiss/gopkg/set"
)

type Int32s []int32

func (p Int32s) AsUint32s_() []uint32 {
	return *(*[]uint32)(unsafe.Pointer(&p))
}

func (p Int32s) ToInt64s() []int64 {
	out := make([]int64, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int64(p[i])
	}
	return out
}

func (p Int32s) ToUint64s() []uint64 {
	out := make([]uint64, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = uint64(p[i])
	}
	return out
}

func (p Int32s) AsInts_() []int {
	if reflectx.IsPlatform32bit {
		return *(*[]int)(unsafe.Pointer(&p))
	}
	out := make([]int, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int(p[i])
	}
	return out
}

func (p Int32s) AsUints_() []uint {
	if reflectx.IsPlatform32bit {
		return *(*[]uint)(unsafe.Pointer(&p))
	}
	out := make([]uint, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = uint(p[i])
	}
	return out
}

func (p Int32s) castType(typ reflect.Type) interface{} {
	return reflectx.CastSlice(p, typ)
}

func (p Int32s) Copy() Int32s {
	out := make([]int32, len(p))
	copy(out, p)
	return out
}

func (p Int32s) ToStrings() Strings {
	out := make([]string, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = strconv.FormatInt(int64(p[i]), 10)
	}
	return out
}

func (p Int32s) ToMap() map[int32]bool {
	out := make(map[int32]bool, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[p[i]] = true
	}
	return out
}

func (p Int32s) ToStringMap() map[string]bool {
	out := make(map[string]bool, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		x := strconv.FormatInt(int64(p[i]), 10)
		out[x] = true
	}
	return out
}

func (p Int32s) ToInterfaceSlice() []interface{} {
	out := make([]interface{}, len(p))
	for i := 0; i < len(p); i++ {
		out[i] = p[i]
	}
	return out
}

func (p Int32s) Drop(inPlace bool, vals ...int32) Int32s {
	var valSet = set.NewInt32(vals...)
	var out = p[:0]
	if !inPlace {
		out = make([]int32, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		if !valSet.Contains(p[i]) {
			out = append(out, p[i])
		}
	}
	return out
}

func (p Int32s) DropFunc(inPlace bool, predicate func(x int32) bool) Int32s {
	var out = p[:0]
	if !inPlace {
		out = make([]int32, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		if !predicate(p[i]) {
			out = append(out, p[i])
		}
	}
	return out
}

func AsInt32s_(intSlice interface{}) Int32s {
	if intSlice == nil {
		return nil
	}

	switch slice := intSlice.(type) {
	case Int32s:
		return slice
	case []int32:
		return slice
	case []uint32:
		return *(*[]int32)(unsafe.Pointer(&slice))
	case Strings:
		return slice.ToInt32s()
	case []string:
		return Strings(slice).ToInt32s()
	}

	sliceTyp := reflect.TypeOf(intSlice)
	assertSliceOfIntegers("AsInt32s_", sliceTyp)

	if reflectx.Is32bitInt(sliceTyp.Elem().Kind()) {
		return reflectx.CastInt32Slice(intSlice)
	}
	return reflectx.ConvertInt32Slice(intSlice)
}

type Int64s []int64

func (p Int64s) AsUint64s_() []uint64 {
	return *(*[]uint64)(unsafe.Pointer(&p))
}

func (p Int64s) ToInt32s() []int32 {
	out := make([]int32, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int32(p[i])
	}
	return out
}

func (p Int64s) ToUint32s() []uint32 {
	out := make([]uint32, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = uint32(p[i])
	}
	return out
}

func (p Int64s) AsInts_() []int {
	if reflectx.IsPlatform64bit {
		return *(*[]int)(unsafe.Pointer(&p))
	}
	out := make([]int, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int(p[i])
	}
	return out
}

func (p Int64s) AsUints_() []uint {
	if reflectx.IsPlatform64bit {
		return *(*[]uint)(unsafe.Pointer(&p))
	}
	out := make([]uint, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = uint(p[i])
	}
	return out
}

func (p Int64s) castType(typ reflect.Type) interface{} {
	return reflectx.CastSlice(p, typ)
}

func (p Int64s) Copy() Int64s {
	out := make([]int64, len(p))
	copy(out, p)
	return out
}

func (p Int64s) ToStrings() Strings {
	out := make([]string, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = strconv.FormatInt(p[i], 10)
	}
	return out
}

func (p Int64s) ToMap() map[int64]bool {
	out := make(map[int64]bool, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[p[i]] = true
	}
	return out
}

func (p Int64s) ToStringMap() map[string]bool {
	out := make(map[string]bool, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		x := strconv.FormatInt(p[i], 10)
		out[x] = true
	}
	return out
}

func (p Int64s) ToInterfaceSlice() []interface{} {
	out := make([]interface{}, len(p))
	for i := 0; i < len(p); i++ {
		out[i] = p[i]
	}
	return out
}

func (p Int64s) Drop(inPlace bool, vals ...int64) Int64s {
	var valSet = set.NewInt64(vals...)
	var out = p[:0]
	if !inPlace {
		out = make([]int64, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		if !valSet.Contains(p[i]) {
			out = append(out, p[i])
		}
	}
	return out
}

func (p Int64s) DropFunc(inPlace bool, predicate func(x int64) bool) Int64s {
	var out = p[:0]
	if !inPlace {
		out = make([]int64, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		for !predicate(p[i]) {
			out = append(out, p[i])
		}
	}
	return out
}

func AsInt64s_(intSlice interface{}) Int64s {
	if intSlice == nil {
		return nil
	}

	switch slice := intSlice.(type) {
	case Int64s:
		return slice
	case []int64:
		return slice
	case []uint64:
		return *(*[]int64)(unsafe.Pointer(&slice))
	case Strings:
		return slice.ToInt64s()
	case []string:
		return Strings(slice).ToInt64s()
	}

	sliceTyp := reflect.TypeOf(intSlice)
	assertSliceOfIntegers("AsInt64s_", sliceTyp)

	if reflectx.Is64bitInt(sliceTyp.Elem().Kind()) {
		return reflectx.CastInt64Slice(intSlice)
	}
	return reflectx.ConvertInt64Slice(intSlice)
}

type Strings []string

func ToStrings_(slice interface{}) Strings {
	switch slice := slice.(type) {
	case []string:
		return slice
	case Strings:
		return slice
	case [][]byte:
		out := make([]string, len(slice))
		for i, x := range slice {
			out[i] = string(x)
		}
		return out
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("ToStrings_: " + errNotSliceType)
	}
	sliceVal := reflect.ValueOf(slice)
	out := make([]string, sliceVal.Len())
	for i := len(out) - 1; i >= 0; i-- {
		str := fmt.Sprint(sliceVal.Index(i).Interface())
		out[i] = str
	}
	return out
}

func (p Strings) castType(typ reflect.Type) interface{} {
	return reflectx.CastSlice(p, typ)
}

func (p Strings) Copy() Strings {
	out := make([]string, len(p))
	copy(out, p)
	return out
}

func (p Strings) ToInt32s() Int32s {
	out := make([]int32, 0, len(p))
	for i := 0; i < len(p); i++ {
		x, err := strconv.ParseInt(p[i], 10, 64)
		if err == nil {
			out = append(out, int32(x))
		}
	}
	return out
}

func (p Strings) ToInt64s() Int64s {
	out := make([]int64, 0, len(p))
	for i := 0; i < len(p); i++ {
		x, err := strconv.ParseInt(p[i], 10, 64)
		if err == nil {
			out = append(out, x)
		}
	}
	return out
}

func (p Strings) ToMap() map[string]bool {
	out := make(map[string]bool, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[p[i]] = true
	}
	return out
}

func (p Strings) ToInterfaceSlice() []interface{} {
	out := make([]interface{}, len(p))
	for i := 0; i < len(p); i++ {
		out[i] = p[i]
	}
	return out
}

func (p Strings) Drop(inPlace bool, vals ...string) Strings {
	var valSet = set.NewString(vals...)
	var out = p[:0]
	if !inPlace {
		out = make([]string, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		if !valSet.Contains(p[i]) {
			out = append(out, p[i])
		}
	}
	return out
}

func (p Strings) DropFunc(inPlace bool, predicate func(x string) bool) Strings {
	var out = p[:0]
	if !inPlace {
		out = make([]string, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		if !predicate(p[i]) {
			out = append(out, p[i])
		}
	}
	return out
}

type Bytes = reflectx.Bytes

func ToBytes_(s string) Bytes {
	return reflectx.ToBytes_(s)
}

func String_(b []byte) string {
	return reflectx.String_(b)
}

func IsNil(x interface{}) bool {
	if x == nil {
		return true
	}
	val := reflect.ValueOf(x)
	if isNillableKind(val.Kind()) {
		return val.IsNil()
	}
	return false
}

func isNillableKind(kind reflect.Kind) bool {
	switch kind {
	case
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice,
		reflect.UnsafePointer:
		return true
	}
	return false
}
