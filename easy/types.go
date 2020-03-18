package easy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

var (
	binEncoding = binary.LittleEndian
	binMagic    = []byte("EZY0")
)

const (
	// intSize is the size in bits of an int or uint value.
	intSize       = 32 << (^uint(0) >> 63)
	platform32bit = intSize == 32
	platform64bit = intSize == 64
)

type Int32s []int32

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
	if platform32bit {
		return *(*[]int)(unsafe.Pointer(&p))
	}
	out := make([]int, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int(p[i])
	}
	return out
}

func (p Int32s) Uints_() []uint {
	if platform32bit {
		return *(*[]uint)(unsafe.Pointer(&p))
	}
	out := make([]uint, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = uint(p[i])
	}
	return out
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
	copy(out, binMagic)
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
	if len(buf) < 4 || !bytes.Equal(buf[:4], binMagic) {
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
	if sliceTyp.Kind() != reflect.Slice || !isIntType(sliceTyp.Elem().Kind()) {
		panic("ToInt32s_: not a slice of integers")
	}

	tab := int64table[sliceTyp.Elem().Kind()]
	if tab.sz == 4 {
		return _castInt32s(intSlice)
	}
	return _convertInt32s(intSlice, tab.sz, tab.fn)
}

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
	if platform64bit {
		return *(*[]int)(unsafe.Pointer(&p))
	}
	out := make([]int, len(p))
	for i := len(p) - 1; i >= 0; i-- {
		out[i] = int(p[i])
	}
	return out
}

func (p Int64s) Uints_() []uint {
	if platform64bit {
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

func (p Int64s) Marshal32() []byte {
	bufLen := 4 + 4*len(p)
	out := make([]byte, bufLen)
	copy(out, binMagic)
	buf := out[4:]
	for i, x := range p {
		binEncoding.PutUint32(buf[4*i:4*(i+1)], uint32(x))
	}
	return out
}

func (p *Int64s) Unmarshal32(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	if len(buf) < 4 || !bytes.Equal(buf[:4], binMagic) {
		return errors.New("invalid bytes format")
	}
	buf = buf[4:]
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

func (p Int64s) Marshal64() []byte {
	bufLen := 4 + 8*len(p)
	out := make([]byte, bufLen)
	copy(out, binMagic)
	buf := out[4:]
	for i, x := range p {
		binEncoding.PutUint64(buf[8*i:8*(i+1)], uint64(x))
	}
	return out
}

func (p *Int64s) Unmarshal64(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	if len(buf) < 4 || !bytes.Equal(buf[:4], binMagic) {
		return errors.New("invalid bytes format")
	}
	buf = buf[4:]
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
	if sliceTyp.Kind() != reflect.Slice || !isIntType(sliceTyp.Elem().Kind()) {
		panic("ToInt64s_: not a slice of integers")
	}

	tab := int64table[sliceTyp.Elem().Kind()]
	if tab.sz == 8 {
		return _castInt64s(intSlice)
	}
	return _convertInt64s(intSlice, tab.sz, tab.fn)
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

type Bytes []byte

func ToBytes_(b interface{}) Bytes {
	switch b := b.(type) {
	case Bytes:
		return b
	case string:
		return s2b(b)
	case []byte:
		return b
	}
	panic("ToBytes_: invalid type (must be string/[]byte)")
}

func (p Bytes) String_() string { return b2s(p) }

func String_(b []byte) string {
	return b2s(b)
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

func _is64bitInt(kind reflect.Kind) bool {
	return isIntType(kind) && int64table[kind].sz == 8
}

func _is32bitInt(kind reflect.Kind) bool {
	return isIntType(kind) && int64table[kind].sz == 4
}

func isIntType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	default:
		return false
	}
}

func isIntTypeOrPtr(typ reflect.Type) bool {
	return isIntType(typ.Kind()) ||
		(typ.Kind() == reflect.Ptr && isIntType(typ.Elem().Kind()))
}

func isStringTypeOrPtr(typ reflect.Type) bool {
	return typ.Kind() == reflect.String ||
		(typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.String)
}

func reflectInt(v reflect.Value) int64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(v.Uint())
	}

	// shall not happen, type should be pre-checked
	panic("bug: not int type")
}
