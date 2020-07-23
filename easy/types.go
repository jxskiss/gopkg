package easy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/jxskiss/gopkg/reflectx"
	"reflect"
	"sort"
	"strconv"
	"unsafe"
)

var (
	binEncoding = binary.LittleEndian
	binMagic32  = []byte("EZY0")
	binMagic64  = []byte("EZY1")
)

type Int32s []int32

func (p Int32s) Len() int           { return len(p) }
func (p Int32s) Less(i, j int) bool { return p[i] < p[j] }
func (p Int32s) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenient method.
func (p Int32s) Sort() { sort.Sort(p) }

func (p Int32s) Uint32s_() []uint32 {
	return *(*[]uint32)(unsafe.Pointer(&p))
}

func (p Int32s) Int64s() []int64 {
	out := make([]int64, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int64(p[i])
	}
	return out
}

func (p Int32s) Uint64s() []uint64 {
	out := make([]uint64, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = uint64(p[i])
	}
	return out
}

func (p Int32s) Ints_() []int {
	if reflectx.IsPlatform32bit {
		return *(*[]int)(unsafe.Pointer(&p))
	}
	out := make([]int, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int(p[i])
	}
	return out
}

func (p Int32s) Uints_() []uint {
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

func (p Int32s) Drop(x int32, inPlace bool) Int32s {
	var out = p[:0]
	if !inPlace {
		out = make([]int32, 0, len(p))
	}
	for i := 0; i < len(p); i++ {
		if p[i] != x {
			out = append(out, p[i])
		}
	}
	return out
}

func (p Int32s) Marshal() []byte {
	bufLen := 4 + 4*len(p)
	out := make([]byte, bufLen)
	copy(out, binMagic32)
	buf := out[4:]
	for i, x := range p {
		binEncoding.PutUint32(buf[4*i:4*(i+1)], uint32(x))
	}
	return out
}

func (p *Int32s) Unmarshal(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	if len(buf) < 4 || !bytes.Equal(buf[:4], binMagic32) {
		return errors.New("invalid bytes format")
	}
	buf = buf[4:]
	if len(buf)%4 != 0 {
		return fmt.Errorf("invalid bytes with length=%d", len(buf))
	}
	slice := *p
	if cap(slice)-len(slice) < len(buf)/4 {
		slice = make([]int32, 0, len(buf)/4)
	}
	for i := 0; i < len(buf); i += 4 {
		x := binEncoding.Uint32(buf[i : i+4])
		slice = append(slice, int32(x))
	}
	*p = slice
	return nil
}

func ToInt32s_(intSlice interface{}) Int32s {
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
	assertSliceOfIntegers("ToInt32s_", sliceTyp)

	if reflectx.Is32bitInt(sliceTyp.Elem().Kind()) {
		return reflectx.CastInt32Slice(intSlice)
	}
	return reflectx.ConvertInt32Slice(intSlice)
}

type Int64s []int64

func (p Int64s) Len() int           { return len(p) }
func (p Int64s) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64s) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p Int64s) Sort() { sort.Sort(p) }

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
	if reflectx.IsPlatform64bit {
		return *(*[]int)(unsafe.Pointer(&p))
	}
	out := make([]int, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int(p[i])
	}
	return out
}

func (p Int64s) Uints_() []uint {
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

func (p Int64s) Marshal32() []byte {
	bufLen := 4 + 4*len(p)
	out := make([]byte, bufLen)
	copy(out, binMagic32)
	buf := out[4:]
	for i, x := range p {
		binEncoding.PutUint32(buf[4*i:4*(i+1)], uint32(x))
	}
	return out
}

func (p Int64s) Marshal64() []byte {
	bufLen := 4 + 8*len(p)
	out := make([]byte, bufLen)
	copy(out, binMagic64)
	buf := out[4:]
	for i, x := range p {
		binEncoding.PutUint64(buf[8*i:8*(i+1)], uint64(x))
	}
	return out
}

func (p *Int64s) Unmarshal(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	if len(buf) < 4 {
		return errors.New("invalid bytes format")
	}
	switch {
	case bytes.Equal(buf[:4], binMagic32):
		return p.unmarshal32(buf[4:])
	case bytes.Equal(buf[:4], binMagic64):
		return p.unmarshal64(buf[4:])
	}
	return errors.New("invalid bytes format")
}

func (p *Int64s) unmarshal32(buf []byte) error {
	if len(buf)%4 != 0 {
		return fmt.Errorf("invalid bytes with length=%d", len(buf))
	}
	slice := *p
	if cap(slice)-len(slice) < len(buf)/4 {
		slice = make([]int64, 0, len(buf)/4)
	}
	for i := 0; i < len(buf); i += 4 {
		x := binEncoding.Uint32(buf[i : i+4])
		slice = append(slice, int64(x))
	}
	*p = slice
	return nil
}

func (p *Int64s) unmarshal64(buf []byte) error {
	if len(buf)%8 != 0 {
		return fmt.Errorf("invalid bytes with length=%d", len(buf))
	}
	slice := *p
	if cap(slice)-len(slice) < len(buf)/8 {
		slice = make([]int64, 0, len(buf)/8)
	}
	for i := 0; i < len(buf); i += 8 {
		x := binEncoding.Uint64(buf[i : i+8])
		slice = append(slice, int64(x))
	}
	*p = slice
	return nil
}

func ToInt64s_(intSlice interface{}) Int64s {
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
	assertSliceOfIntegers("ToInt64s_", sliceTyp)

	if reflectx.Is64bitInt(sliceTyp.Elem().Kind()) {
		return reflectx.CastInt64Slice(intSlice)
	}
	return reflectx.ConvertInt64Slice(intSlice)
}

type Strings []string

func (p Strings) Len() int           { return len(p) }
func (p Strings) Less(i, j int) bool { return p[i] < p[j] }
func (p Strings) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p Strings) Sort() { sort.Sort(p) }

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

type Bytes = reflectx.Bytes

func ToBytes_(b interface{}) Bytes {
	return reflectx.ToBytes_(b)
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
