package easy

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/jxskiss/gopkg/internal/unsafeheader"
	"github.com/jxskiss/gopkg/reflectx"
)

const (
	maxInsertGrowth = 1024
)

//nolint:unused
const (
	errNotSliceType           = "not slice type"
	errNotSliceOfInt          = "not a slice of integers"
	errElemTypeNotMatchSlice  = "elem type does not match slice"
	errElemNotStructOrPointer = "elem is not struct or pointer to struct"
	errStructFieldNotProvided = "struct field is not provided"
	errStructFieldNotExists   = "struct field not exists"
	errStructFieldIsNotInt    = "struct field is not integer or pointer"
	errStructFieldIsNotStr    = "struct field is not string or pointer"
)

func panicNilParams(where string, params ...interface{}) {
	for i := 0; i < len(params); i += 2 {
		arg := params[i].(string)
		val := params[i+1]
		if val == nil {
			panic(fmt.Sprintf("%s: param %s is nil interface", where, arg))
		}
	}
}

func assertSliceOfIntegers(where string, sliceTyp reflect.Type) {
	if sliceTyp.Kind() != reflect.Slice || !reflectx.IsIntType(sliceTyp.Elem().Kind()) {
		panic(where + ":" + errNotSliceOfInt)
	}
}

func assertSliceAndElemType(where string, sliceVal reflect.Value, elemTyp reflect.Type) (reflect.Value, bool) {
	if sliceVal.Kind() != reflect.Slice {
		panic(where + ": " + errNotSliceType)
	}
	intTypeNotMatch := false
	sliceTyp := sliceVal.Type()
	if elemTyp != sliceTyp.Elem() {
		// int-family
		if reflectx.IsIntType(sliceTyp.Elem().Kind()) &&
			reflectx.IsIntType(elemTyp.Kind()) {
			intTypeNotMatch = true
		} else {
			panic(where + ": " + errElemTypeNotMatchSlice)
		}
	}
	return sliceVal, intTypeNotMatch
}

func assertSliceElemStructAndField(where string, sliceTyp reflect.Type, field string) reflect.StructField {
	if field == "" {
		panic(where + ": " + errStructFieldNotProvided)
	}
	if sliceTyp.Kind() != reflect.Slice {
		panic(where + ": " + errNotSliceType)
	}
	elemTyp := sliceTyp.Elem()
	elemIsPtr := elemTyp.Kind() == reflect.Ptr
	if !(elemTyp.Kind() == reflect.Struct ||
		(elemIsPtr && elemTyp.Elem().Kind() == reflect.Struct)) {
		panic(where + ": " + errElemNotStructOrPointer)
	}
	var fieldInfo reflect.StructField
	var ok bool
	if elemIsPtr {
		fieldInfo, ok = elemTyp.Elem().FieldByName(field)
	} else {
		fieldInfo, ok = elemTyp.FieldByName(field)
	}
	if !ok {
		panic(where + ": " + errStructFieldNotExists)
	}
	return fieldInfo
}

// InSlice is reserved for a generic implementation.

// InSliceFunc iterates the given slice, it calls predicate(i) for i in range [0, n)
// where n is the length of the slice.
// When predicate(i) returns true, it stops and returns true.
//
// The parameter predicate must be not nil, otherwise it panics.
func InSliceFunc(slice interface{}, predicate func(i int) bool) bool {
	if slice == nil {
		return false
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("InSliceFunc: " + errNotSliceType)
	}
	_, header := reflectx.UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		if predicate(i) {
			return true
		}
	}
	return false
}

// InInt32s tells whether the int32 value elem is in the slice.
func InInt32s(slice []int32, elem int32) bool {
	for _, x := range slice {
		if x == elem {
			return true
		}
	}
	return false
}

// InInt64s tells whether the int64 value elem is in the slice.
func InInt64s(slice []int64, elem int64) bool {
	for _, x := range slice {
		if x == elem {
			return true
		}
	}
	return false
}

// InStrings tells whether the string value elem is in the slice.
func InStrings(slice []string, elem string) bool {
	for _, x := range slice {
		if elem == x {
			return true
		}
	}
	return false
}

// Index is reserved for a generic implementation.

// IndexFunc iterates the given slice, it calls predicate(i) for i in
// range [0, n) where n is the length of the slice.
// When predicate(i) returns true, it stops and returns the index i.
//
// The parameter predicate must not be nil, otherwise it panics.
func IndexFunc(slice interface{}, predicate func(i int) bool) int {
	if slice == nil {
		return -1
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("IndexFunc: " + errNotSliceType)
	}
	_, header := reflectx.UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// IndexInt32s returns the index of the first instance of elem slice,
// or -1 if elem is not present in slice.
func IndexInt32s(slice []int32, elem int32) int {
	for i := 0; i < len(slice); i++ {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

// IndexInt64s returns the index of the first instance of elem in slice,
// or -1 if elem is not present in slice.
func IndexInt64s(slice []int64, elem int64) int {
	for i := 0; i < len(slice); i++ {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

// IndexStrings returns the index of the first instance of elem in slice,
// or -1 if elem is not present in slice.
func IndexStrings(slice []string, elem string) int {
	for i := 0; i < len(slice); i++ {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

// LastIndex is reserved for a generic implementation.

// LastIndexFunc iterates the given slice, it calls predicate(i) for i in
// range [0, n) in descending order, where n is the length of the slice.
// When predicate(i) returns true, it stops and returns the index i.
//
// The parameter predicate must not be nil, otherwise it panics.
func LastIndexFunc(slice interface{}, predicate func(i int) bool) int {
	if slice == nil {
		return -1
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("LastIndexFunc: " + errNotSliceType)
	}
	_, header := reflectx.UnpackSlice(slice)
	for i := header.Len - 1; i >= 0; i-- {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// LastIndexInt32s returns the index of the last instance of elem in slice,
// or -1 if elem is not present in slice.
func LastIndexInt32s(slice []int32, elem int32) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

// LastIndexInt64s returns the index of the last instance of elem in slice,
// or -1 if elem is not present in slice.
func LastIndexInt64s(slice []int64, elem int64) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

// LastIndexStrings returns the index of the last instance of elem in slice,
// or -1 if elem is not present in slice.
func LastIndexStrings(slice []string, elem string) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

// InsertSlice is reserved for a generic implementation.

// InsertInt32s inserts the given int32 elem into the slice at index position.
// If index is equal or greater than the length of slice, the elem will be
// appended to the end of the slice. In case the slice is full of it's
// capacity, a new slice will be created and returned.
func InsertInt32s(slice []int32, index int, elem int32) (out []int32) {
	if index >= len(slice) {
		return append(slice, elem)
	}
	oldLen := len(slice)
	if len(slice) == cap(slice) {
		// capacity not enough, grow the slice
		newCap := oldLen + min(max(1, oldLen), maxInsertGrowth)
		out = make([]int32, oldLen+1, newCap)
		copy(out, slice[:index])
	} else {
		out = slice[:oldLen+1]
	}
	copy(out[index+1:], slice[index:])
	out[index] = elem
	return
}

// InsertInt64s inserts the given int64 elem into the slice at index position.
// If index is equal or greater than the length of slice, the elem will be
// appended to the end of the slice. In case the slice is full of it's
// capacity, a new slice will be created and returned.
func InsertInt64s(slice []int64, index int, elem int64) (out []int64) {
	if index >= len(slice) {
		return append(slice, elem)
	}
	oldLen := len(slice)
	if len(slice) == cap(slice) {
		// capacity not enough, grow the slice
		newCap := oldLen + min(max(1, oldLen), maxInsertGrowth)
		out = make([]int64, oldLen+1, newCap)
		copy(out, slice[:index])
	} else {
		out = slice[:oldLen+1]
	}
	copy(out[index+1:], slice[index:])
	out[index] = elem
	return
}

// InsertStrings inserts the given string elem into the slice at index position.
// If index is equal or greater than the length of slice, the elem will be
// appended to the end of the slice. In case the slice is full of it's
// capacity, a new slice will be created and returned.
func InsertStrings(slice []string, index int, elem string) (out []string) {
	if index >= len(slice) {
		return append(slice, elem)
	}
	oldLen := len(slice)
	if len(slice) == cap(slice) {
		// capacity not enough, grow the slice
		newCap := oldLen + min(max(1, oldLen), maxInsertGrowth)
		out = make([]string, oldLen+1, newCap)
		copy(out, slice[:index])
	} else {
		out = slice[:oldLen+1]
	}
	copy(out[index+1:], slice[index:])
	out[index] = elem
	return
}

// Reverse is reserved for a generic implementation.

// ReverseSlice returns a new slice containing the elements of the given
// slice in reversed order.
//
// When inplace is true, the slice is reversed in place, it does not create
// a new slice, but returns the original slice with reversed order.
//
// The given slice must be not nil, otherwise it panics.
func ReverseSlice(slice interface{}, inplace bool) interface{} {
	if slice == nil {
		panicNilParams("ReverseSlice", "slice", slice)
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("ReverseSlice: " + errNotSliceType)
	}

	_, srcHeader := reflectx.UnpackSlice(slice)
	length := srcHeader.Len
	elemTyp := sliceTyp.Elem()
	elemRType := reflectx.ToRType(elemTyp)
	elemSize := elemRType.Size()

	outSlice, outHeader := slice, srcHeader
	if !inplace {
		outSlice, outHeader = reflectx.MakeSlice(elemTyp, length, length)
		reflectx.TypedSliceCopy(elemRType, *outHeader, *srcHeader)
	}
	tmp := reflect.New(elemTyp).Elem().Interface()
	swap := reflectx.EfaceOf(&tmp).Word
	for i, mid := 0, length/2; i < mid; i++ {
		j := length - i - 1
		pi := reflectx.ArrayAt(outHeader.Data, i, elemSize)
		pj := reflectx.ArrayAt(outHeader.Data, j, elemSize)
		reflectx.TypedMemMove(elemRType, swap, pi)
		reflectx.TypedMemMove(elemRType, pi, pj)
		reflectx.TypedMemMove(elemRType, pj, swap)
	}
	return outSlice
}

// ReverseInt32s returns a new slice of the elements in reversed order.
//
// When inplace is true, the slice is reversed in place, it does not create
// a new slice, but returns the original slice with reversed order.
func ReverseInt32s(slice []int32, inplace bool) []int32 {
	length := len(slice)
	out := slice
	if !inplace {
		out = make([]int32, length)
		copy(out, slice)
	}
	for i, mid := 0, length/2; i < mid; i++ {
		j := length - i - 1
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// ReverseInt64s returns a new slice of the elements in reversed order.
//
// When inplace is true, the slice is reversed in place, it does not create
// a new slice, but returns the original slice with reversed order.
func ReverseInt64s(slice []int64, inplace bool) []int64 {
	length := len(slice)
	out := slice
	if !inplace {
		out = make([]int64, length)
		copy(out, slice)
	}
	for i, mid := 0, length/2; i < mid; i++ {
		j := length - i - 1
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// ReverseStrings returns a new slice of the elements in reversed order.
//
// When inplace is true, the slice is reversed in place, it does not create
// a new slice, but returns the original slice with reversed order.
func ReverseStrings(slice []string, inplace bool) []string {
	length := len(slice)
	out := slice
	if !inplace {
		out = make([]string, length)
		copy(out, slice)
	}
	for i, mid := 0, length/2; i < mid; i++ {
		j := length - i - 1
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// UniqueSlice is reserved for a generic implementation.

// UniqueInt32s returns a new slice containing the elements of the given
// slice in same order, but filter out duplicate values.
//
// When inplace is true, it does not create a new slice, the unique values
// will be written to the input slice from the beginning.
func UniqueInt32s(slice []int32, inplace bool) []int32 {
	seen := make(map[int32]struct{})
	out := slice[:0]
	if !inplace {
		out = make([]int32, 0)
	}
	for _, x := range slice {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

// UniqueInt64s returns a new slice containing the elements of the given
// slice in same order, but filter out duplicate values.
//
// When inplace is true, it does not create a new slice, the unique values
// will be written to the input slice from the beginning.
func UniqueInt64s(slice []int64, inplace bool) []int64 {
	seen := make(map[int64]struct{})
	out := slice[:0]
	if !inplace {
		out = make([]int64, 0)
	}
	for _, x := range slice {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

// UniqueStrings returns a new slice containing the elements of the given
// slice in same order, but filter out duplicate values.
//
// When inplace is true, it does not create a new slice, the unique values
// will be written to the input slice from the beginning.
func UniqueStrings(slice []string, inplace bool) []string {
	seen := make(map[string]struct{})
	out := slice[:0]
	if !inplace {
		out = make([]string, 0)
	}
	for _, x := range slice {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

// DiffSlice is reserved for a generic implementation.

// DiffInt32s returns a new int32 slice containing the values which present
// in slice a but not present in slice b.
func DiffInt32s(a []int32, b []int32) []int32 {
	bset := make(map[int32]struct{}, len(b))
	for _, x := range b {
		bset[x] = struct{}{}
	}
	out := make([]int32, 0)
	for _, x := range a {
		if _, ok := bset[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// DiffInt64s returns a new int64 slice containing the values which present
// in slice a but not present in slice b.
func DiffInt64s(a []int64, b []int64) []int64 {
	bset := make(map[int64]struct{}, len(b))
	for _, x := range b {
		bset[x] = struct{}{}
	}
	out := make([]int64, 0)
	for _, x := range a {
		if _, ok := bset[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// DiffStrings returns a new string slice containing the values which
// present in slice a but not present in slice b.
func DiffStrings(a []string, b []string) []string {
	bset := make(map[string]struct{}, len(b))
	for _, x := range b {
		bset[x] = struct{}{}
	}
	out := make([]string, 0)
	for _, x := range a {
		if _, ok := bset[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// ToMap converts the given slice of struct (or pointer to struct) to a map,
// with the field specified by keyField as key and the slice element as value.
//
// If slice is nil, keyField does not exists or the element of slice is not
// struct or pointer to struct, it panics.
func ToMap(slice interface{}, keyField string) interface{} {
	if slice == nil {
		panicNilParams("ToMap", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("ToMap", sliceTyp, keyField)
	keyTyp := fieldInfo.Type
	if keyTyp.Kind() == reflect.Ptr {
		keyTyp = keyTyp.Elem()
	}

	elemTyp := sliceTyp.Elem()
	outVal := reflect.MakeMapWithSize(reflect.MapOf(keyTyp, elemTyp), sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		fieldVal := reflect.Indirect(reflect.Indirect(elem).FieldByName(keyField))
		outVal.SetMapIndex(fieldVal, elem)
	}
	return outVal.Interface()
}

// ToSliceMap converts the given slice of struct (or pointer to struct) to a map,
// with the field specified by keyField as key and a slice of elements which have
// same key as value.
//
// If slice is nil, keyField does not exists or the element of slice is not
// struct or pointer to struct, it panics.
func ToSliceMap(slice interface{}, keyField string) interface{} {
	if slice == nil {
		panicNilParams("ToSliceMap", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("ToSliceMap", sliceTyp, keyField)
	keyTyp := fieldInfo.Type
	if keyTyp.Kind() == reflect.Ptr {
		keyTyp = keyTyp.Elem()
	}

	elemTyp := sliceTyp.Elem()
	elemSliceTyp := reflect.SliceOf(elemTyp)
	outVal := reflect.MakeMap(reflect.MapOf(keyTyp, elemSliceTyp))
	for i := sliceVal.Len() - 1; i >= 0; i-- {
		elem := sliceVal.Index(i)
		fieldVal := reflect.Indirect(reflect.Indirect(elem).FieldByName(keyField))
		elemSlice := outVal.MapIndex(fieldVal)
		if !elemSlice.IsValid() {
			elemSlice = reflect.MakeSlice(elemSliceTyp, 0, 1)
		}
		elemSlice = reflect.Append(elemSlice, elem)
		outVal.SetMapIndex(fieldVal, elemSlice)
	}
	return outVal.Interface()
}

// ToMapMap converts the given slice of struct (or pointer to struct) to a map,
// with the field specified by keyField as key.
// The returned map's value is another map with the field specified by
// subKeyField as key and thee slice element as value.
//
// If slice is nil, keyField or subKeyField does not exists or the element of
// slice is not struct or pointer to struct, it panics.
func ToMapMap(slice interface{}, keyField, subKeyField string) interface{} {
	if slice == nil {
		panicNilParams("ToMapMap", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
	fieldInfo1 := assertSliceElemStructAndField("ToMapMap", sliceTyp, keyField)
	fieldInfo2 := assertSliceElemStructAndField("ToMapMap", sliceTyp, subKeyField)
	keyTyp1 := fieldInfo1.Type
	if keyTyp1.Kind() == reflect.Ptr {
		keyTyp1 = keyTyp1.Elem()
	}
	keyTyp2 := fieldInfo2.Type
	if keyTyp2.Kind() == reflect.Ptr {
		keyTyp2 = keyTyp2.Elem()
	}

	elemTyp := sliceTyp.Elem()
	elemMapTyp := reflect.MapOf(keyTyp2, elemTyp)
	outVal := reflect.MakeMap(reflect.MapOf(keyTyp1, elemMapTyp))
	for i := sliceVal.Len() - 1; i >= 0; i-- {
		elem := sliceVal.Index(i)
		fieldVal1 := reflect.Indirect(reflect.Indirect(elem).FieldByName(keyField))
		fieldVal2 := reflect.Indirect(reflect.Indirect(elem).FieldByName(subKeyField))
		elemMap := outVal.MapIndex(fieldVal1)
		if !elemMap.IsValid() {
			elemMap = reflect.MakeMap(elemMapTyp)
			outVal.SetMapIndex(fieldVal1, elemMap)
		}
		elemMap.SetMapIndex(fieldVal2, elem)
	}
	return outVal.Interface()
}

// ToInterfaceSlice returns a []interface{} containing elements from slice.
func ToInterfaceSlice(slice interface{}) []interface{} {
	if slice == nil {
		return nil
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("ToInterfaceSlice: " + errNotSliceType)
	}
	sliceVal := reflect.ValueOf(slice)
	out := make([]interface{}, 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i).Interface()
		out = append(out, elem)
	}
	return out
}

// Find is reserved for a generic implementation.

// FindFunc returns the first element in the slice for which predicate returns true.
func FindFunc(slice interface{}, predicate func(i int) bool) interface{} {
	if slice == nil {
		return nil
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("FindFunc: " + errNotSliceType)
	}
	_, header := reflectx.UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		if predicate(i) {
			return reflect.ValueOf(slice).Index(i).Interface()
		}
	}
	return nil
}

// Filter is reserved for a generic implementation.

// FilterFunc iterates the given slice, it calls predicate(i) for i in
// range [0, n), where n is the length of the slice.
// It returns a new slice of elements for which predicate(i) returns true.
//
// The parameter slice and predicate must not be nil, otherwise it panics.
func FilterFunc(slice interface{}, predicate func(i int) bool) interface{} {
	if slice == nil {
		panicNilParams("FilterFunc", "slice", slice)
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("FilterFunc: " + errNotSliceType)
	}

	sliceVal := reflect.ValueOf(slice)
	length := sliceVal.Len()
	outVal := reflect.MakeSlice(sliceVal.Type(), 0, max(length/4+1, 4))
	for i := 0; i < length; i++ {
		if predicate(i) {
			elem := sliceVal.Index(i)
			outVal = reflect.Append(outVal, elem)
		}
	}
	return outVal.Interface()
}

// FilterInt32s iterates the given slice, it calls predicate(i) for i in
// range [0, n), where n is the length of the slice.
// It returns a new slice of elements for which predicate(i) returns true.
func FilterInt32s(slice []int32, predicate func(i int) bool) []int32 {
	length := len(slice)
	out := make([]int32, 0, max(length/4+1, 4))
	for i := 0; i < length; i++ {
		if predicate(i) {
			out = append(out, slice[i])
		}
	}
	return out
}

// FilterInt64s iterates the given slice, it calls predicate(i) for i in
// range [0, n), where n is the length of the slice.
// It returns a new slice of elements for which predicate(i) returns true.
func FilterInt64s(slice []int64, predicate func(i int) bool) []int64 {
	length := len(slice)
	out := make([]int64, 0, max(length/4+1, 4))
	for i := 0; i < length; i++ {
		if predicate(i) {
			out = append(out, slice[i])
		}
	}
	return out
}

// FilterStrings iterates the given slice, it calls predicate(i) for i in
// range [0, n), where n is the length of the slice.
// It returns a new slice of elements for which predicate(i) returns true.
func FilterStrings(slice []string, predicate func(i int) bool) []string {
	length := len(slice)
	out := make([]string, 0, max(length/4+1, 4))
	for i := 0; i < length; i++ {
		if predicate(i) {
			out = append(out, slice[i])
		}
	}
	return out
}

// SumSlice returns the sum value of the elements in the given slice.
// If slice is nil or it's elements are not integers, it panics.
func SumSlice(slice interface{}) int64 {
	if slice == nil {
		panicNilParams("SumSlice", "slice", slice)
	}
	sliceTyp := reflect.TypeOf(slice)
	assertSliceOfIntegers("SumSlice", sliceTyp)

	var sum int64
	elemTyp := sliceTyp.Elem()
	elemKind := elemTyp.Kind()
	elemSize := elemTyp.Size()
	_, header := reflectx.UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		ptr := reflectx.ArrayAt(header.Data, i, elemSize)
		sum += reflectx.CastIntPointer(elemKind, ptr)
	}
	return sum
}

// ParseInt64s parses a number string separated by sep into a []int64 slice.
// If there is invalid number value, it reports malformed = true as the
// second return value.
func ParseInt64s(values, sep string, ignoreZero bool) (slice []int64, malformed bool) {
	values = strings.TrimSpace(values)
	values = strings.Trim(values, sep)
	segments := strings.Split(values, sep)
	slice = make([]int64, 0, len(segments))
	for _, x := range segments {
		id, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			malformed = true
			continue
		}
		if id == 0 && ignoreZero {
			continue
		}
		slice = append(slice, id)
	}
	return
}

// JoinInt64s returns a string consisting of slice elements separated by sep.
func JoinInt64s(slice []int64, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return strconv.FormatInt(slice[0], 10)
	}
	var buf []byte
	buf = strconv.AppendInt(buf, slice[0], 10)
	for _, x := range slice[1:] {
		buf = append(buf, sep...)
		buf = strconv.AppendInt(buf, x, 10)
	}
	return unsafeheader.BytesToString(buf)
}

// IJ represents a slice index of I, J.
type IJ struct{ I, J int }

// SplitBatch splits a large number to batches, it's mainly designed to
// help operations with large slice, such as inserting lots of records
// into database, or logging lots of identifiers, etc.
func SplitBatch(total, batch int) []IJ {
	if total <= 0 {
		return nil
	}
	if batch <= 0 {
		return []IJ{{0, total}}
	}
	n := total/batch + 1
	ret := make([]IJ, n)
	idx := 0
	for i, j := 0, batch; idx < n && i < total; i, j = i+batch, j+batch {
		if j > total {
			j = total
		}
		ret[idx] = IJ{i, j}
		idx++
	}
	return ret[:idx]
}

// Split is reserved for a generic implementation.

// SplitSlice splits a large slice []T to batches, it returns a slice
// of slice of type [][]T.
//
// The given slice must not be nil, otherwise it panics.
func SplitSlice(slice interface{}, batch int) interface{} {
	if slice == nil {
		panicNilParams("SplitSlice", "slice", slice)
	}
	sliceTyp := reflectx.RTypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("SplitSlice: " + errNotSliceType)
	}

	_, sliceHeader := reflectx.UnpackSlice(slice)
	indexes := SplitBatch(sliceHeader.Len, batch)
	elemTyp := sliceTyp.Elem()
	elemSize := elemTyp.Size()
	out := make([]reflectx.SliceHeader, len(indexes))
	for i, idx := range indexes {
		subSlice := _takeSlice(sliceHeader.Data, elemSize, idx.I, idx.J)
		out[i] = subSlice
	}

	outTyp := reflectx.SliceOf(sliceTyp)
	return outTyp.PackInterface(unsafe.Pointer(&out))
}

func _takeSlice(base unsafe.Pointer, elemSize uintptr, i, j int) (slice reflectx.SliceHeader) {
	if length := j - i; length > 0 {
		slice.Data = reflectx.ArrayAt(base, i, elemSize)
		slice.Len = length
		slice.Cap = length
	}
	return
}

// SplitInt64s splits a large int64 slice to batches.
func SplitInt64s(slice []int64, batch int) [][]int64 {
	indexes := SplitBatch(len(slice), batch)
	out := make([][]int64, len(indexes))
	for i, idx := range indexes {
		out[i] = slice[idx.I:idx.J]
	}
	return out
}

// SplitInt32s splits a large int32 slice to batches.
func SplitInt32s(slice []int32, batch int) [][]int32 {
	indexes := SplitBatch(len(slice), batch)
	out := make([][]int32, len(indexes))
	for i, idx := range indexes {
		out[i] = slice[idx.I:idx.J]
	}
	return out
}

// SplitStrings splits a large string slice to batches.
func SplitStrings(slice []string, batch int) [][]string {
	indexes := SplitBatch(len(slice), batch)
	out := make([][]string, len(indexes))
	for i, idx := range indexes {
		out[i] = slice[idx.I:idx.J]
	}
	return out
}

func isIntTypeOrPtr(typ reflect.Type) bool {
	if reflectx.IsIntType(typ.Kind()) ||
		(typ.Kind() == reflect.Ptr && reflectx.IsIntType(typ.Elem().Kind())) {
		return true
	}
	return false
}

func isStringTypeOrPtr(typ reflect.Type) bool {
	return typ.Kind() == reflect.String ||
		(typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.String)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
