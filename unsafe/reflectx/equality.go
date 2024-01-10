package reflectx

import (
	"fmt"
	"reflect"
	"strings"
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
func IsIdenticalType(a, b any) (equal bool, diff string) {
	typ1 := reflect.TypeOf(a)
	typ2 := reflect.TypeOf(b)
	return newStrictTypecmp().isEqualType(typ1, typ2)
}

// IsIdenticalThriftType checks whether the given two object types have same
// struct fields and memory layout, in case that a field's name does not
// match, but the thrift tag's first two fields match, it also considers
// the field matches.
//
// It is almost same with IsIdenticalType, but helps the situation that
// different Thrift generators which generate different field names.
//
// If two types are identical, it is expected that unsafe pointer casting
// between the two types won't crash the program.
// If the given two types are not identical, the returned diff message
// contains the detail difference.
func IsIdenticalThriftType(a, b any) (equal bool, diff string) {
	typ1 := reflect.TypeOf(a)
	typ2 := reflect.TypeOf(b)
	return newThriftTypecmp().isEqualType(typ1, typ2)
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

	fieldCmp func(typ reflect.Type, f1, f2 reflect.StructField) (bool, string)
}

func newStrictTypecmp() *typecmp {
	cmp := &typecmp{
		seen: make(map[typ1typ2]*cmpresult),
	}
	cmp.fieldCmp = cmp.isStrictEqualField
	return cmp
}

func newThriftTypecmp() *typecmp {
	cmp := &typecmp{
		seen: make(map[typ1typ2]*cmpresult),
	}
	cmp.fieldCmp = cmp.isEqualThriftField
	return cmp
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
	if typ1.Kind() != reflect.Struct || typ2.Kind() != reflect.Struct {
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
		if equal, diff := p.fieldCmp(typ1, f1, f2); !equal {
			diff = fmt.Sprintf("struct field not equal: %s", diff)
			p.seen[typidx] = &cmpresult{notequal, diff}
			return false, diff
		}
	}
	p.seen[typidx] = &cmpresult{isequal, ""}
	return true, ""
}

func (p *typecmp) isStrictEqualField(typ reflect.Type, field1, field2 reflect.StructField) (bool, string) {
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

func (p *typecmp) isEqualThriftField(typ reflect.Type, field1, field2 reflect.StructField) (bool, string) {
	if field1.Name != field2.Name {
		if !isEqualThriftTag(field1, field2) {
			return false, fmt.Sprintf("field name and thrift tag both not equal: %s, %s", _f(typ, field1), _f(typ, field2))
		}
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

func isEqualThriftTag(f1, f2 reflect.StructField) bool {
	// Be compatible with standard thrift tag.
	tag1 := f1.Tag.Get("thrift")
	tag2 := f2.Tag.Get("thrift")
	if tag1 != "" && tag2 != "" {
		if tag1 == tag2 {
			return true
		}
		parts1 := strings.Split(tag1, ",")
		parts2 := strings.Split(tag2, ",")
		if len(parts1) >= 2 && len(parts2) >= 2 &&
			parts1[0] == parts2[0] && // parts[0] is the field's name
			parts1[1] == parts2[1] { // parts[1] is the field's id number
			return true
		}
	}
	// Be compatible with github.com/cloudwego/frugal.
	tag1 = f1.Tag.Get("frugal")
	tag2 = f2.Tag.Get("frugal")
	if tag1 != "" && tag2 != "" {
		if tag1 == tag2 {
			return true
		}
		parts1 := strings.Split(tag1, ",")
		parts2 := strings.Split(tag2, ",")
		if len(parts1) >= 3 && len(parts2) >= 3 &&
			parts1[0] == parts2[0] && // parts[0] is the field's id number
			parts1[2] == parts2[2] { // parts[2] is the field's type
			return true
		}
	}
	return false
}

func (p *typecmp) isEqualSlice(typ1, typ2 reflect.Type) (bool, string) {
	if typ1.Kind() != reflect.Slice || typ2.Kind() != reflect.Slice {
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
