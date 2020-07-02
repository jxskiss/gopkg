package easy

import (
	"fmt"
	"github.com/jxskiss/gopkg/reflectx"
	"reflect"
)

const (
	errNotSliceType           = "not slice type"
	errNotSliceOrPointer      = "not a slice or pointer to slice"
	errNotSliceOfInt          = "not a slice of integers"
	errNotMapOfSlice          = "not a map of slice"
	errNotMapOfIntSlice       = "not a map of integer slice"
	errNotSameTypeOrNotMap    = "not same type or not map"
	errElemTypeNotMatchSlice  = "elem type does not match slice"
	errElemNotStructOrPointer = "elem is not struct or pointer to struct"
	errStructFieldNotProvided = "struct field is not provided"
	errStructFieldNotExists   = "struct field not exists"
	errStructFieldIsNotInt    = "struct field is not integer or pointer"
	errStructFieldIsNotStr    = "struct field is not string or pointer"
	errPredicateFuncSig       = "predicate func signature not match"
)

func panicNilParams(where string, params ...interface{}) {
	const nilInterface = "%s: param %s is nil interface"

	for i := 0; i < len(params); i += 2 {
		arg := params[i].(string)
		val := params[i+1]
		if val == nil {
			panic(fmt.Sprintf(nilInterface, where, arg))
		}
	}
}

func invalidType(where string, want string, got interface{}) string {
	const invalidType = "%s: invalid type, want %s, got %T"

	return fmt.Sprintf(invalidType, where, want, got)
}

func assertSliceOfIntegers(where string, sliceTyp reflect.Type) {
	if sliceTyp.Kind() != reflect.Slice || !reflectx.IsIntType(sliceTyp.Elem().Kind()) {
		panic(where + ":" + errNotSliceOfInt)
	}
}

func assertSliceAndElemType(where string, sliceVal reflect.Value, elemTyp reflect.Type) (reflect.Value, bool) {
	sliceVal = indirect(sliceVal)
	if sliceVal.Kind() != reflect.Slice {
		panic(where + ": " + errNotSliceOrPointer)
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
		panic(where + ": " + errNotSliceOrPointer)
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

func assertSliceAndPredicateFunc(where string, sliceVal reflect.Value, fnTyp reflect.Type) reflect.Value {
	sliceVal = indirect(sliceVal)
	if sliceVal.Kind() != reflect.Slice {
		panic(where + ": " + errNotSliceOrPointer)
	}
	elemTyp := sliceVal.Type().Elem()
	if !(fnTyp.Kind() == reflect.Func &&
		fnTyp.NumIn() == 1 && fnTyp.NumOut() == 1 &&
		(fnTyp.In(0).Kind() == reflect.Interface || fnTyp.In(0) == elemTyp) &&
		fnTyp.Out(0).Kind() == reflect.Bool) {
		panic(where + ": " + errPredicateFuncSig)
	}
	return sliceVal
}
