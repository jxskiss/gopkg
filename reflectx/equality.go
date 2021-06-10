package reflectx

import (
	"fmt"
	"reflect"
)

// IsIdenticalType checks whether the given two object types have same
// struct fields and memory layout (same order, same name and same type).
// It's useful to check generated types are exactly same in different
// packages, e.g. Thrift, Protobuf, Msgpack, etc.
//
// If two types are identical, it is expected that unsafe pointer casting
// between the two types won't crash the program.
// If the given two types are not identical, the returned diff message
// contains the detail difference.
func IsIdenticalType(a, b interface{}) (equal bool, diff string) {
	typ1 := reflect.TypeOf(a)
	typ2 := reflect.TypeOf(b)
	return newTypecmp().isEqualType(typ1, typ2)
}

type typ1typ2 struct {
	typ1, typ2 reflect.Type
}

const (
	notequal = 1
	isequal  = 2
	checking = 3
)

type cmpresult struct {
	result int
	diff   string
}

type typecmp struct {
	seen map[typ1typ2]*cmpresult
}

func newTypecmp() *typecmp {
	return &typecmp{seen: make(map[typ1typ2]*cmpresult)}
}

func (p *typecmp) isEqualType(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != typ2.Kind() {
		return false, fmt.Sprintf("type kind not equal: %s, %s", _t(typ1), _t(typ2))
	}
	if typ1.Kind() == reflect.Ptr {
		if typ1.Elem().Kind() != typ2.Elem().Kind() {
			return false, fmt.Sprintf("pointer type not equal: %s, %s", _t(typ1), _t(typ2))
		}
		typ1 = typ1.Elem()
		typ2 = typ2.Elem()
		return p.isEqualType(typ1, typ2)
	}
	switch typ1.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return true, ""
	case reflect.Struct:
		return p.isEqualStruct(typ1, typ2)
	case reflect.Slice:
		return p.isEqualSlice(typ1, typ2)
	case reflect.Map:
		return p.isEqualMap(typ1, typ2)
	}
	return false, fmt.Sprintf("unsupported types: %s, %s", _t(typ1), _t(typ2))
}

func (p *typecmp) isEqualStruct(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != reflect.Struct || typ1.Kind() != reflect.Struct {
		return false, fmt.Sprintf("type is not struct: %s, %s", _t(typ1), _t(typ2))
	}
	typidx := typ1typ2{typ1, typ2}
	if cmpr := p.seen[typidx]; cmpr != nil {

		// In case of recursive type, cmpr.result will be checking here,
		// we treat it as equal, the final result will be updated below.
		return cmpr.result != notequal, cmpr.diff
	}

	p.seen[typidx] = &cmpresult{checking, ""}
	if typ1.NumField() != typ2.NumField() {
		diff := fmt.Sprintf("struct field num not match: %s, %s", _t(typ1), _t(typ2))
		p.seen[typidx] = &cmpresult{notequal, diff}
		return false, diff
	}
	fnum := typ1.NumField()
	for i := 0; i < fnum; i++ {
		f1 := typ1.Field(i)
		f2 := typ2.Field(i)
		if equal, diff := p.isEqualField(typ1, f1, f2); !equal {
			diff = fmt.Sprintf("struct field not equal: %s", diff)
			p.seen[typidx] = &cmpresult{notequal, diff}
			return false, diff
		}
	}
	p.seen[typidx] = &cmpresult{isequal, ""}
	return true, ""
}

func (p *typecmp) isEqualField(typ reflect.Type, field1, field2 reflect.StructField) (bool, string) {
	if field1.Name != field2.Name {
		return false, fmt.Sprintf("field name not equal: %s, %s", _f(typ, field1), _f(typ, field2))
	}
	if field1.Offset != field2.Offset {
		return false, fmt.Sprintf("field offset not equal: %s, %s", _f(typ, field1), _f(typ, field2))
	}

	typ1 := field1.Type
	typ2 := field2.Type
	if typ1.Kind() != typ2.Kind() {
		return false, fmt.Sprintf("field type not equal: %s, %s", _f(typ, field1), _f(typ, field2))
	}
	if typ1.Kind() == reflect.Ptr {
		typ1 = typ1.Elem()
		typ2 = typ2.Elem()
	}
	equal, diff := p.isEqualType(typ1, typ2)
	if equal {
		return true, ""
	}
	return false, fmt.Sprintf("field type not euqal: %s", diff)
}

func (p *typecmp) isEqualSlice(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != reflect.Slice || typ1.Kind() != reflect.Slice {
		return false, fmt.Sprintf("type is not slice: %s, %s", _t(typ1), _t(typ2))
	}
	typ1 = typ1.Elem()
	typ2 = typ2.Elem()
	return p.isEqualType(typ1, typ2)
}

func (p *typecmp) isEqualMap(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != reflect.Map || typ2.Kind() != reflect.Map {
		return false, fmt.Sprintf("type is not map: %s, %s", _t(typ1), _t(typ2))
	}

	keyTyp1 := typ1.Key()
	keyTyp2 := typ2.Key()
	if equal, diff := p.isEqualType(keyTyp1, keyTyp2); !equal {
		return false, fmt.Sprintf("map key: %s", diff)
	}

	elemTyp1 := typ1.Elem()
	elemTyp2 := typ2.Elem()
	if equal, diff := p.isEqualType(elemTyp1, elemTyp2); !equal {
		return false, fmt.Sprintf("map value: %s", diff)
	}

	return true, ""
}

func _t(typ reflect.Type) string {
	return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
}

func _f(typ reflect.Type, field reflect.StructField) string {
	return fmt.Sprintf("%s.%s", _t(typ), field.Name)
}
