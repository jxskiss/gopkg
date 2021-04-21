package easy

import (
	"github.com/jxskiss/gopkg/reflectx"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

const (
	maxInsertGrowth = 1024
)

// InSlice iterates the given slice, it calls predicate(i) for i in range [0, n)
// where n is the length of the slice.
// When predicate(i) returns true, it stops and returns true.
//
// The parameter predicate must be not nil, otherwise it panics.
func InSlice(slice interface{}, predicate func(i int) bool) bool {
	if slice == nil {
		return false
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("Find: " + errNotSliceType)
	}
	header := reflectx.UnpackSlice(slice)
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

// Int64s tells whether the int64 value elem is in the slice.
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

// Index iterates the given slice, it calls predicate(i) for i in range [0, n)
// where n is the length of the slice.
// When predicate(i) returns true, it stops and returns the index i.
//
// The parameter predicate must be not nil, otherwise it panics.
func Index(slice interface{}, predicate func(i int) bool) int {
	if slice == nil {
		return -1
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("Index: " + errNotSliceType)
	}
	header := reflectx.UnpackSlice(slice)
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

// LastIndex iterates the given slice, it calls predicate(i) for i in range [0, n)
// in descending order, where n is the length of the slice.
// When predicate(i) returns true, it stops and returns the index i.
//
// The parameter predicate must be not nil, otherwise it panics.
func LastIndex(slice interface{}, predicate func(i int) bool) int {
	if slice == nil {
		return -1
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("LastIndex: " + errNotSliceType)
	}
	header := reflectx.UnpackSlice(slice)
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

// InsertSlice inserts the given elem into the slice at index position.
// If index is equal or greater than the length of slice, the elem will be
// appended to the end of the slice. In case the slice is full of it's
// capacity, a new slice will be created and returned.
//
// The parameter slice must be a slice, index must be a positive number
// and elem must be same type of the slice element, otherwise it panics.
func InsertSlice(slice interface{}, index int, elem interface{}) (out interface{}) {
	if slice == nil || elem == nil {
		panicNilParams("InsertSlice", "slice", slice, "elem", elem)
	}
	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	elemKind := elemTyp.Kind()
	if sliceTyp.Kind() == reflect.Slice && sliceTyp.Elem().Kind() == elemKind {
		if reflectx.Is64bitInt(elemKind) {
			return InsertInt64s(ToInt64s_(slice), index, _int64(elem)).castType(sliceTyp)
		}
		if reflectx.Is32bitInt(elemKind) {
			return InsertInt32s(ToInt32s_(slice), index, _int32(elem)).castType(sliceTyp)
		}
		if elemKind == reflect.String {
			return InsertStrings(ToStrings_(slice), index, _string(elem)).castType(sliceTyp)
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType("InsertSlice", reflect.ValueOf(slice), elemTyp)

	var outVal reflect.Value
	oldLen := sliceVal.Len()
	if index >= oldLen {
		index = oldLen
	}
	if sliceVal.Cap() == oldLen {
		// capacity not enough, grow the slice
		newCap := oldLen + min(max(1, oldLen), maxInsertGrowth)
		outVal = reflect.MakeSlice(sliceVal.Type(), oldLen+1, newCap)
		reflect.Copy(outVal, sliceVal.Slice(0, index))
	} else {
		outVal = sliceVal.Slice(0, oldLen+1)
	}
	if index < oldLen {
		reflect.Copy(outVal.Slice(index+1, oldLen+1), sliceVal.Slice(index, oldLen))
	}
	if intTypeNotMatch {
		_elemInt := _int64(elem)
		outVal.Index(index).SetInt(_elemInt)
	} else {
		outVal.Index(index).Set(reflect.ValueOf(elem))
	}
	return outVal.Interface()
}

// InsertInt32s inserts the given int32 elem into the slice at index position.
// If index is equal or greater than the length of slice, the elem will be
// appended to the end of the slice. In case the slice is full of it's
// capacity, a new slice will be created and returned.
func InsertInt32s(slice []int32, index int, elem int32) (out Int32s) {
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
func InsertInt64s(slice []int64, index int, elem int64) (out Int64s) {
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
func InsertStrings(slice []string, index int, elem string) (out Strings) {
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

// ReverseSlice returns a new slice containing the elements of the given
// slice in reversed order.
//
// The given slice must be not nil, otherwise it panics.
func ReverseSlice(slice interface{}) interface{} {
	if slice == nil {
		panicNilParams("ReverseSlice", "slice", slice)
	}
	sliceTyp := reflect.TypeOf(slice)
	switch slice := slice.(type) {
	case Int64s, []int64, []uint64:
		return ReverseInt64s(ToInt64s_(slice)).castType(sliceTyp)
	case Int32s, []int32, []uint32:
		return ReverseInt32s(ToInt32s_(slice)).castType(sliceTyp)
	case Strings, []string:
		return ReverseStrings(ToStrings_(slice)).castType(sliceTyp)
	}

	if sliceTyp.Kind() != reflect.Slice {
		panic("ReverseSlice: " + errNotSliceType)
	}

	srcHeader := reflectx.UnpackSlice(slice)
	length := srcHeader.Len
	elemTyp := sliceTyp.Elem()
	elemRType := reflectx.ToRType(elemTyp)
	elemSize := elemRType.Size()
	outSlice, outHeader := reflectx.MakeSlice(elemTyp, length, length)
	reflectx.TypedSliceCopy(elemRType, *outHeader, srcHeader)

	tmp := reflect.New(elemTyp).Elem().Interface()
	swap := reflectx.EFaceOf(&tmp).Word
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

// ReverseSliceInplace reverse the given slice inplace.
// The given slice must be not nil, otherwise it panics.
func ReverseSliceInplace(slice interface{}) {
	if slice == nil {
		panicNilParams("ReverseSliceInplace", "slice", slice)
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("ReverseSliceInplace: " + errNotSliceType)
	}
	elemTyp := sliceTyp.Elem()
	elemRType := reflectx.ToRType(elemTyp)

	switch elemRType.Kind() {
	case reflect.Int64, reflect.Uint64:
		values := *(*[]int64)(reflectx.EFaceOf(&slice).Word)
		ReverseInt64sInplace(values)
		return
	case reflect.Int32, reflect.Uint32:
		values := *(*[]int32)(reflectx.EFaceOf(&slice).Word)
		ReverseInt32sInplace(values)
		return
	case reflect.String:
		values := *(*[]string)(reflectx.EFaceOf(&slice).Word)
		ReverseStringsInplace(values)
		return
	}

	header := reflectx.UnpackSlice(slice)
	length := header.Len
	elemSize := elemRType.Size()

	tmp := reflect.New(elemTyp).Elem().Interface()
	swap := reflectx.EFaceOf(&tmp).Word
	for i, mid := 0, length/2; i < mid; i++ {
		j := length - 1 - i
		pi := reflectx.ArrayAt(header.Data, i, elemSize)
		pj := reflectx.ArrayAt(header.Data, j, elemSize)
		reflectx.TypedMemMove(elemRType, swap, pi)
		reflectx.TypedMemMove(elemRType, pi, pj)
		reflectx.TypedMemMove(elemRType, pj, swap)
	}
}

// ReverseInt32s returns a new slice of the elements in reversed order.
func ReverseInt32s(slice []int32) Int32s {
	length := len(slice)
	out := make([]int32, length)
	for i, x := range slice {
		out[length-1-i] = x
	}
	return out
}

// ReverseInt32sInplace reverse the in32 slice inplace.
func ReverseInt32sInplace(slice []int32) {
	length := len(slice)
	for i, mid := 0, length/2; i < mid; i++ {
		j := length - 1 - i
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// ReverseInt64s returns a new slice of the elements in reversed order.
func ReverseInt64s(slice []int64) Int64s {
	length := len(slice)
	out := make([]int64, length)
	for i, x := range slice {
		out[length-1-i] = x
	}
	return out
}

// ReverseInt64sInplace reverse the int64 slice inplace.
func ReverseInt64sInplace(slice []int64) {
	length := len(slice)
	for i, mid := 0, length/2; i < mid; i++ {
		j := length - 1 - i
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// ReverseStrings returns a new slice of the elements in reversed order.
func ReverseStrings(slice []string) Strings {
	length := len(slice)
	out := make([]string, length)
	for i, x := range slice {
		out[length-1-i] = x
	}
	return out
}

// ReverseStringsInplace reverse the string slice inplace.
func ReverseStringsInplace(slice []string) {
	length := len(slice)
	for i, mid := 0, length/2; i < mid; i++ {
		j := length - 1 - i
		slice[i], slice[j] = slice[j], slice[i]
	}
}

var (
	emptyStructVal = reflect.ValueOf(struct{}{})
	emptyStructTyp = reflect.TypeOf(struct{}{})
)

// UniqueSlice returns a new slice containing the elements of the given
// slice in same order, but filter out duplicate values.
//
// The given slice must be not nil and element must be hashable (can be
// used as map key), otherwise it panics.
func UniqueSlice(slice interface{}) interface{} {
	if slice == nil {
		panicNilParams("UniqueSlice", "slice", slice)
	}
	sliceTyp := reflect.TypeOf(slice)
	switch slice := slice.(type) {
	case Int64s, []int64, []uint64:
		return UniqueInt64s(ToInt64s_(slice)).castType(sliceTyp)
	case Int32s, []int32, []uint32:
		return UniqueInt32s(ToInt32s_(slice)).castType(sliceTyp)
	case Strings, []string:
		return UniqueStrings(ToStrings_(slice)).castType(sliceTyp)
	}

	if sliceTyp.Kind() != reflect.Slice {
		panicNilParams("UniqueSlice: " + errNotSliceType)
	}
	setTyp := reflect.MapOf(sliceTyp.Elem(), emptyStructTyp)
	seen := reflect.MakeMap(setTyp)
	sliceVal := reflect.ValueOf(slice)
	outVal := reflect.MakeSlice(sliceTyp, 0, 8)
	sliceLen := sliceVal.Len()
	for i := 0; i < sliceLen; i++ {
		elem := sliceVal.Index(i)
		if seen.MapIndex(elem).IsValid() {
			continue
		}
		seen.SetMapIndex(elem, emptyStructVal)
		outVal = reflect.Append(outVal, elem)
	}
	return outVal.Interface()
}

// UniqueInt32s returns a new slice containing the elements of the given
// slice in same order, but filter out duplicate values.
func UniqueInt32s(slice []int32) Int32s {
	seen := make(map[int32]struct{})
	out := make([]int32, 0)
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
func UniqueInt64s(slice []int64) Int64s {
	seen := make(map[int64]struct{})
	out := make([]int64, 0)
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
func UniqueStrings(slice []string) Strings {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, x := range slice {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

// DiffInt32s returns a new int32 slice containing the values which present
// in slice a but not present in slice b.
func DiffInt32s(a []int32, b []int32) Int32s {
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
func DiffInt64s(a []int64, b []int64) Int64s {
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
func DiffStrings(a []string, b []string) Strings {
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

// Pluck accepts a slice of struct or pointer to struct, and returns a
// new slice with field values specified by field of the struct elements.
func Pluck(slice interface{}, field string) interface{} {
	if slice == nil {
		panicNilParams("Pluck", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("Pluck", sliceTyp, field)

	var outVal = reflect.MakeSlice(reflect.SliceOf(fieldInfo.Type), 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := reflect.Indirect(sliceVal.Index(i))
		fieldVal := elem.FieldByName(field)
		outVal = reflect.Append(outVal, fieldVal)
	}
	return outVal.Interface()
}

// PluckInt32s accepts a slice of struct or pointer to struct, and returns a
// new int32 slice with field values specified by field of the struct elements.
func PluckInt32s(slice interface{}, field string) Int32s {
	if slice == nil {
		panicNilParams("PluckInt32s", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("PluckInt32s", sliceTyp, field)
	if !reflectx.IsIntTypeOrPtr(fieldInfo.Type) {
		panic("PluckInt32s: " + errStructFieldIsNotInt)
	}

	out := make([]int32, 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := reflect.Indirect(sliceVal.Index(i))
		fieldVal := reflect.Indirect(elem.FieldByName(field))
		if fieldVal.IsValid() {
			out = append(out, int32(reflectInt(fieldVal)))
		}
	}
	return out
}

// PluckInt64s accepts a slice of struct or pointer to struct, and returns a
// new int64 slice with field values specified by field of the struct elements.
func PluckInt64s(slice interface{}, field string) Int64s {
	if slice == nil {
		panicNilParams("PluckInt64s", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("PluckInt64s", sliceTyp, field)
	if !reflectx.IsIntTypeOrPtr(fieldInfo.Type) {
		panic("PluckInt64s: " + errStructFieldIsNotInt)
	}

	out := make([]int64, 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := reflect.Indirect(sliceVal.Index(i))
		fieldVal := reflect.Indirect(elem.FieldByName(field))
		if fieldVal.IsValid() {
			out = append(out, reflectInt(fieldVal))
		}
	}
	return out
}

// PluckStrings accepts a slice of struct or pointer to struct, and returns a
// new string slice with field values specified by field of the struct elements.
func PluckStrings(slice interface{}, field string) Strings {
	if slice == nil {
		panicNilParams("PluckStrings", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
	fieldInfo := assertSliceElemStructAndField("PluckStrings", sliceTyp, field)
	if !reflectx.IsStringTypeOrPtr(fieldInfo.Type) {
		panic("PluckStrings: " + errStructFieldIsNotStr)
	}

	out := make([]string, 0, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		elem := reflect.Indirect(sliceVal.Index(i))
		fieldVal := reflect.Indirect(elem.FieldByName(field))
		if fieldVal.IsValid() {
			out = append(out, fieldVal.String())
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

// Find returns the first element in the slice for which predicate returns true.
func Find(slice interface{}, predicate func(i int) bool) interface{} {
	if slice == nil {
		return nil
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("Find: " + errNotSliceType)
	}
	header := reflectx.UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		if predicate(i) {
			return reflect.ValueOf(slice).Index(i).Interface()
		}
	}
	return nil
}

// Filter iterates the given slice, it calls predicate(i) for i in range [0, n)
// where n is the length of the slice.
// It returns a new slice of elements for which predicate(i) returns true.
//
// The parameter slice and predicate must not be nil, otherwise it panics.
func Filter(slice interface{}, predicate func(i int) bool) interface{} {
	if slice == nil {
		panicNilParams("Filter", "slice", slice)
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("Filter: " + errNotSliceType)
	}
	elemKind := sliceTyp.Elem().Kind()
	if reflectx.Is64bitInt(elemKind) {
		return FilterInt64s(ToInt64s_(slice), predicate).castType(sliceTyp)
	}
	if reflectx.Is32bitInt(elemKind) {
		return FilterInt32s(ToInt32s_(slice), predicate).castType(sliceTyp)
	}
	if elemKind == reflect.String {
		return FilterStrings(ToStrings_(slice), predicate).castType(sliceTyp)
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

func FilterInt32s(slice []int32, predicate func(i int) bool) Int32s {
	length := len(slice)
	out := make([]int32, 0, max(length/4+1, 4))
	for i := 0; i < length; i++ {
		if predicate(i) {
			out = append(out, slice[i])
		}
	}
	return out
}

func FilterInt64s(slice []int64, predicate func(i int) bool) Int64s {
	length := len(slice)
	out := make([]int64, 0, max(length/4+1, 4))
	for i := 0; i < length; i++ {
		if predicate(i) {
			out = append(out, slice[i])
		}
	}
	return out
}

func FilterStrings(slice []string, predicate func(i int) bool) Strings {
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
	info := reflectx.GetIntCaster(sliceTyp.Elem().Kind())
	header := reflectx.UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		ptr := reflectx.ArrayAt(header.Data, i, info.Size)
		sum += info.Cast(ptr)
	}
	return sum
}

// SumMapSlice returns the sum value of the slice elements in the given map.
//
// The given map must not be nil and the map's value must be slice of integers,
// otherwise it panics.
func SumMapSlice(mapOfSlice interface{}) int64 {
	if mapOfSlice == nil {
		panicNilParams("SumMapSlice", "mapOfSlice", mapOfSlice)
	}
	mTyp := reflect.TypeOf(mapOfSlice)
	if mTyp.Kind() != reflect.Map || mTyp.Elem().Kind() != reflect.Slice ||
		!reflectx.IsIntType(mTyp.Elem().Elem().Kind()) {
		panic("SumMapSlice: " + errNotMapOfIntSlice)
	}

	var sum int64
	elemTyp := mTyp.Elem().Elem()
	info := reflectx.GetIntCaster(elemTyp.Kind())
	reflectx.MapIterPointer(mapOfSlice, func(_, v unsafe.Pointer) int {
		header := *(*reflectx.SliceHeader)(v)
		for i := 0; i < header.Len; i++ {
			ptr := reflectx.ArrayAt(header.Data, i, info.Size)
			sum += info.Cast(ptr)
		}
		return 0
	})
	return sum
}

// SumMapSliceLength returns the sum length of the slice values in the give map.
//
// The given map must not be nil and the map's value must be slice (of any type),
// otherwise it panics.
func SumMapSliceLength(mapOfSlice interface{}) int {
	if mapOfSlice == nil {
		panicNilParams("SumMapSliceLength", "mapOfSlice", mapOfSlice)
	}
	mTyp := reflect.TypeOf(mapOfSlice)
	if mTyp.Kind() != reflect.Map || mTyp.Elem().Kind() != reflect.Slice {
		panic("SumMapSliceLength: " + errNotMapOfSlice)
	}

	var sumLen int
	reflectx.MapIterPointer(mapOfSlice, func(_, v unsafe.Pointer) int {
		header := *(*reflectx.SliceHeader)(v)
		sumLen += header.Len
		return 0
	})
	return sumLen
}

func ParseInt64s(values, sep string, ignoreZero bool) (slice Int64s, malformed bool) {
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

func JoinInt64s(slice []int64, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return strconv.FormatInt(slice[0], 10)
	}
	var buf Bytes
	buf = strconv.AppendInt(buf, slice[0], 10)
	for _, x := range slice[1:] {
		buf = append(buf, sep...)
		buf = strconv.AppendInt(buf, x, 10)
	}
	return buf.String_()
}

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

// SplitSlice splits a large slice []T to batches, it returns a slice
// of slice of type [][]T.
//
// The given slice must not be nil, otherwise it panics.
func SplitSlice(slice interface{}, batch int) interface{} {
	if slice == nil {
		panicNilParams("SplitSlice", "slice", slice)
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("SplitSlice: " + errNotSliceType)
	}

	sliceHeader := reflectx.UnpackSlice(slice)
	indexes := SplitBatch(sliceHeader.Len, batch)
	elemTyp := sliceTyp.Elem()
	elemSize := elemTyp.Size()
	out := make([]reflectx.SliceHeader, len(indexes))
	for i, idx := range indexes {
		subSlice := _takeSlice(sliceHeader.Data, elemSize, idx.I, idx.J)
		out[i] = subSlice
	}

	outType := reflect.SliceOf(sliceTyp)
	return reflectx.CastSlice(out, outType)
}

func _takeSlice(base unsafe.Pointer, elemSize uintptr, i, j int) (slice reflectx.SliceHeader) {
	if length := j - i; length > 0 {
		slice.Data = reflectx.ArrayAt(base, i, elemSize)
		slice.Len = length
		slice.Cap = length
	}
	return
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
