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

func InSlice(slice interface{}, elem interface{}) bool {
	if slice == nil {
		return false
	}
	if elem == nil {
		sliceVal := indirect(reflect.ValueOf(slice))
		sliceTyp := sliceVal.Type()
		if sliceTyp.Kind() == reflect.Slice && isNillableKind(sliceTyp.Elem().Kind()) {
			for i := 0; i < sliceVal.Len(); i++ {
				if sliceVal.Index(i).IsNil() {
					return true
				}
			}
		}
		return false
	}

	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	elemKind := elemTyp.Kind()
	if sliceTyp.Kind() == reflect.Slice && sliceTyp.Elem().Kind() == elemKind {
		if reflectx.Is64bitInt(elemKind) {
			return InInt64s(ToInt64s_(slice), _int64(elem))
		}
		if reflectx.Is32bitInt(elemKind) {
			return InInt32s(ToInt32s_(slice), _int32(elem))
		}
		if elemKind == reflect.String {
			return InStrings(ToStrings_(slice), _string(elem))
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType("InSlice", reflect.ValueOf(slice), elemTyp)

	if intTypeNotMatch {
		_elemInt := _int64(elem)
		for i := 0; i < sliceVal.Len(); i++ {
			_sliceInt := reflectInt(sliceVal.Index(i))
			if _elemInt == _sliceInt {
				return true
			}
		}
		return false
	}

	for i := 0; i < sliceVal.Len(); i++ {
		if elem == sliceVal.Index(i).Interface() {
			return true
		}
	}
	return false
}

func InInt32s(slice []int32, elem int32) bool {
	for _, x := range slice {
		if x == elem {
			return true
		}
	}
	return false
}

func InInt64s(slice []int64, elem int64) bool {
	for _, x := range slice {
		if x == elem {
			return true
		}
	}
	return false
}

func InStrings(slice []string, elem string) bool {
	for _, x := range slice {
		if elem == x {
			return true
		}
	}
	return false
}

func Index(slice interface{}, elem interface{}) int {
	if slice == nil {
		return -1
	}
	if elem == nil {
		sliceVal := indirect(reflect.ValueOf(slice))
		sliceTyp := sliceVal.Type()
		if sliceTyp.Kind() == reflect.Slice && isNillableKind(sliceTyp.Elem().Kind()) {
			for i := 0; i < sliceVal.Len(); i++ {
				if sliceVal.Index(i).IsNil() {
					return i
				}
			}
		}
		return -1
	}

	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	elemKind := elemTyp.Kind()
	if sliceTyp.Kind() == reflect.Slice && sliceTyp.Elem().Kind() == elemKind {
		if reflectx.Is64bitInt(elemKind) {
			return IndexInt64s(ToInt64s_(slice), _int64(elem))
		}
		if reflectx.Is32bitInt(elemKind) {
			return IndexInt32s(ToInt32s_(slice), _int32(elem))
		}
		if elemKind == reflect.String {
			return IndexStrings(ToStrings_(slice), _string(elem))
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType("Index", reflect.ValueOf(slice), elemTyp)

	if intTypeNotMatch {
		_elemInt := _int64(elem)
		for i := 0; i < sliceVal.Len(); i++ {
			_sliceInt := reflectInt(sliceVal.Index(i))
			if _elemInt == _sliceInt {
				return i
			}
		}
		return -1
	}

	for i := 0; i < sliceVal.Len(); i++ {
		if elem == sliceVal.Index(i).Interface() {
			return i
		}
	}
	return -1
}

func IndexInt32s(slice []int32, elem int32) int {
	for i := 0; i < len(slice); i++ {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func IndexInt64s(slice []int64, elem int64) int {
	for i := 0; i < len(slice); i++ {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func IndexStrings(slice []string, elem string) int {
	for i := 0; i < len(slice); i++ {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func LastIndex(slice interface{}, elem interface{}) int {
	if slice == nil {
		return -1
	}
	if elem == nil {
		sliceVal := indirect(reflect.ValueOf(slice))
		sliceTyp := sliceVal.Type()
		if sliceTyp.Kind() == reflect.Slice && isNillableKind(sliceTyp.Elem().Kind()) {
			for i := sliceVal.Len() - 1; i >= 0; i-- {
				if sliceVal.Index(i).IsNil() {
					return i
				}
			}
		}
		return -1
	}

	sliceTyp := reflect.TypeOf(slice)
	elemTyp := reflect.TypeOf(elem)
	elemKind := elemTyp.Kind()
	if sliceTyp.Kind() == reflect.Slice && sliceTyp.Elem().Kind() == elemKind {
		if reflectx.Is64bitInt(elemKind) {
			return LastIndexInt64s(ToInt64s_(slice), _int64(elem))
		}
		if reflectx.Is32bitInt(elemKind) {
			return LastIndexInt32s(ToInt32s_(slice), _int32(elem))
		}
		if elemKind == reflect.String {
			return LastIndexStrings(ToStrings_(slice), _string(elem))
		}
	}

	sliceVal, intTypeNotMatch := assertSliceAndElemType("LastIndex", reflect.ValueOf(slice), elemTyp)

	if intTypeNotMatch {
		_elemInt := _int64(elem)
		for i := sliceVal.Len() - 1; i >= 0; i-- {
			_sliceInt := reflectInt(sliceVal.Index(i))
			if _elemInt == _sliceInt {
				return i
			}
		}
		return -1
	}

	for i := sliceVal.Len() - 1; i >= 0; i-- {
		if elem == sliceVal.Index(i).Interface() {
			return i
		}
	}
	return -1
}

func LastIndexInt32s(slice []int32, elem int32) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func LastIndexInt64s(slice []int64, elem int64) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

func LastIndexStrings(slice []string, elem string) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if elem == slice[i] {
			return i
		}
	}
	return -1
}

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
	sliceVal := reflect.ValueOf(slice)
	outVal := reflect.MakeSlice(sliceTyp, 0, sliceVal.Len())
	for i := sliceVal.Len() - 1; i >= 0; i-- {
		outVal = reflect.Append(outVal, sliceVal.Index(i))
	}
	return outVal.Interface()
}

func ReverseInt32s(slice []int32) Int32s {
	out := make([]int32, 0, len(slice))
	for i := len(slice) - 1; i >= 0; i-- {
		out = append(out, slice[i])
	}
	return out
}

func ReverseInt64s(slice []int64) Int64s {
	out := make([]int64, 0, len(slice))
	for i := len(slice) - 1; i >= 0; i-- {
		out = append(out, slice[i])
	}
	return out
}

func ReverseStrings(slice []string) Strings {
	out := make([]string, 0, len(slice))
	for i := len(slice) - 1; i >= 0; i-- {
		out = append(out, slice[i])
	}
	return out
}

var (
	emptyStructVal = reflect.ValueOf(struct{}{})
	emptyStructTyp = reflect.TypeOf(struct{}{})
)

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

func Find(slice interface{}, predicate func(i int) bool) int {
	if slice == nil || predicate == nil {
		panicNilParams("Find", "slice", slice, "predicate", predicate)
	}
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp.Kind() != reflect.Slice {
		panic("Find: " + errNotSliceType)
	}
	header := reflectx.UnpackSlice(slice)
	for i := 0; i < header.Len; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func Filter(slice interface{}, predicate func(i int) bool) interface{} {
	if slice == nil || predicate == nil {
		panicNilParams("Filter", "slice", slice, "predicate", predicate)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
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
	reflectx.MapIter(mapOfSlice, func(_, v unsafe.Pointer) {
		header := *(*reflectx.SliceHeader)(v)
		for i := 0; i < header.Len; i++ {
			ptr := reflectx.ArrayAt(header.Data, i, info.Size)
			sum += info.Cast(ptr)
		}
	})
	return sum
}

func SumMapSliceLength(mapOfSlice interface{}) int {
	if mapOfSlice == nil {
		panicNilParams("SumMapSliceLength", "mapOfSlice", mapOfSlice)
	}
	mTyp := reflect.TypeOf(mapOfSlice)
	if mTyp.Kind() != reflect.Map || mTyp.Elem().Kind() != reflect.Slice {
		panic("SumMapSliceLength: " + errNotMapOfSlice)
	}

	var sumLen int
	reflectx.MapIter(mapOfSlice, func(_, v unsafe.Pointer) {
		header := *(*reflectx.SliceHeader)(v)
		sumLen += header.Len
	})
	return sumLen
}

func ParseInt64s(values, sep string, ignoreZero bool) (slice Int64s, isMalformed bool) {
	values = strings.TrimSpace(values)
	values = strings.Trim(values, sep)
	segments := strings.Split(values, sep)
	slice = make([]int64, 0, len(segments))
	for _, x := range segments {
		id, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			isMalformed = true
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

func SplitSlice(slice interface{}, batch int) interface{} {
	if slice == nil {
		panicNilParams("SplitSlice", "slice", slice)
	}
	sliceVal := reflect.ValueOf(slice)
	sliceTyp := sliceVal.Type()
	if sliceTyp.Kind() != reflect.Slice {
		panic("SplitSlice: " + errNotSliceType)
	}

	indexes := SplitBatch(sliceVal.Len(), batch)
	out := reflect.MakeSlice(reflect.SliceOf(sliceTyp), 0, len(indexes))
	for _, idx := range indexes {
		out = reflect.Append(out, sliceVal.Slice(idx.I, idx.J))
	}
	return out.Interface()
}

func indirect(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}
	return value
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
