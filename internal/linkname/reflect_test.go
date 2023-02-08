package linkname

import (
	"sort"
	"testing"
)

func compileReflectFunctions() {
	call(Reflect_typelinks)
	call(Reflect_resolveTypeOff)
	call(Reflect_rtype_Align)
	call(Reflect_rtype_FieldAlign)
	call(Reflect_rtype_Method)
	call(Reflect_rtype_MethodByName)
	call(Reflect_rtype_NumMethod)
	call(Reflect_rtype_Name)
	call(Reflect_rtype_PkgPath)
	call(Reflect_rtype_Size)
	call(Reflect_rtype_String)
	call(Reflect_rtype_Kind)
	call(Reflect_rtype_Implements)
	call(Reflect_rtype_AssignableTo)
	call(Reflect_rtype_ConvertibleTo)
	call(Reflect_rtype_Comparable)
	call(Reflect_rtype_Bits)
	call(Reflect_rtype_ChanDir)
	call(Reflect_rtype_IsVariadic)
	call(Reflect_rtype_Elem)
	call(Reflect_rtype_Field)
	call(Reflect_rtype_FieldByIndex)
	call(Reflect_rtype_FieldByName)
	call(Reflect_rtype_FieldByNameFunc)
	call(Reflect_rtype_In)
	call(Reflect_rtype_Key)
	call(Reflect_rtype_Len)
	call(Reflect_rtype_NumField)
	call(Reflect_rtype_NumIn)
	call(Reflect_rtype_NumOut)
	call(Reflect_rtype_Out)
	call(Reflect_rtype_ptrTo)
	call(Reflect_ifaceIndir)
	call(Reflect_toType)
	call(Reflect_unsafe_New)
	call(Reflect_unsafe_NewArray)
	call(Reflect_typedmemmove)
	call(Reflect_typedslicecopy)
	call(Reflect_maplen)
	call(Reflect_mapiterinit)
	call(Reflect_mapiterkey)
	call(Reflect_mapiterelem)
	call(Reflect_mapiternext)
}

func TestReflect_mapiterinit(t *testing.T) {
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	var val any = m
	ef := unpackEface(&val)
	it := Reflect_mapiterinit(ef.rtype, ef.data)

	var keys []string
	var values []int
	mlen := Reflect_maplen(ef.data)
	for i := 0; i < mlen; i++ {
		k := Reflect_mapiterkey(it)
		v := Reflect_mapiterelem(it)
		keys = append(keys, *(*string)(k))
		values = append(values, *(*int)(v))
		Reflect_mapiternext(it)
	}

	sort.Strings(keys)
	sort.Ints(values)
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("got unexpected keys: %v", keys)
	}
	if values[0] != 1 || values[1] != 2 || values[2] != 3 {
		t.Errorf("got unexpected values: %v", values)
	}
}
