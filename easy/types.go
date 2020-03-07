package easy

import (
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

var ErrNotSliceOfInt = errors.New("not a slice of integers")

// intSize is the size in bits of an int or uint value.
const intSize = 32 << (^uint(0) >> 63)

type Int64s []int64

func (p Int64s) Uint64s_() []uint64 {
	return *(*[]uint64)(unsafe.Pointer(&p))
}

func (p Int64s) Int32s() []int32 {
	out := make([]int32, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int32(p[i])
	}
	return out
}

func (p Int64s) Uint32s() []uint32 {
	out := make([]uint32, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = uint32(p[i])
	}
	return out
}

func (p Int64s) Ints_() []int {
	if intSize == 64 {
		return *(*[]int)(unsafe.Pointer(&p))
	}
	out := make([]int, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int(p[i])
	}
	return out
}

func (p Int64s) Uints_() []uint {
	if intSize == 64 {
		return *(*[]uint)(unsafe.Pointer(&p))
	}
	out := make([]uint, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = uint(p[i])
	}
	return out
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

func (p Int64s) Drop(x int64, inPlace bool) Int64s {
	var out = p[:0]
	if !inPlace {
		out = make([]int64, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		if p[i] != x {
			out = append(out, p[i])
		}
	}
	return out
}

func ToInt64s_(intSlice interface{}) Int64s {
	switch slice := intSlice.(type) {
	case []int64:
		return slice
	case []uint64:
		return *(*[]int64)(unsafe.Pointer(&slice))
	case []int, []uint, []uintptr:
		if intSize == 64 {
			iface := *(*[2]unsafe.Pointer)(unsafe.Pointer(&slice))
			return *(*[]int64)(iface[1])
		}
	}

	sliceTyp := reflect.TypeOf(intSlice)
	if sliceTyp.Kind() != reflect.Slice || !isIntType(sliceTyp.Elem()) {
		panic(ErrNotSliceOfInt)
	}

	sliceVal := reflect.ValueOf(intSlice)
	out := make([]int64, sliceVal.Len())
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = reflectInt(sliceVal.Index(i))
	}
	return out
}

type Strings []string

func (p Strings) Copy() Strings {
	out := make([]string, len(p))
	copy(out, p)
	return out
}

func (p Strings) ToInt64s() Int64s {
	out := make([]int64, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		x, err := strconv.ParseInt(p[i], 10, 64)
		if err == nil {
			out[i] = x
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

func (p Strings) Drop(x string, inPlace bool) Strings {
	var out = p[:0]
	if !inPlace {
		out = make([]string, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		if p[i] != x {
			out = append(out, p[i])
		}
	}
	return out
}

type bytes_ []byte

func Bytes_(b interface{}) bytes_ {
	switch b := b.(type) {
	case string:
		return s2b(b)
	case []byte:
		return b
	}
	panic("invalid type for bytes (string/[]byte)")
}

func (p bytes_) String() string { return b2s(p) }

func (p bytes_) Bytes() []byte { return p }

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}
