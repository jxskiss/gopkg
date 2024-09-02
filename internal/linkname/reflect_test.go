package linkname

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
	call(Reflect_ifaceIndir)
	call(Reflect_unsafe_New)
	call(Reflect_unsafe_NewArray)
	call(Reflect_typedmemmove)
	call(Reflect_typedslicecopy)
	call(Reflect_maplen)
}

var reflectSourceCode = []SourceCodeTestCase{
	{
		MinVer:   newVer(1, 19, 0),
		MaxVer:   newVer(1, 20, 999),
		FileName: "reflect/type.go",
		Lines: []string{
			"func (t *rtype) common() *rtype { return t }",
			"func typelinks() (sections []unsafe.Pointer, offset [][]int32)",
			"func resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer",
			"func ifaceIndir(t *rtype) bool",
		},
	},
	{
		MinVer:   newVer(1, 21, 0),
		FileName: "reflect/type.go",
		Lines: []string{
			"func (t *rtype) common() *abi.Type",
			"func typelinks() (sections []unsafe.Pointer, offset [][]int32)",
			"func resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer",
		},
	},
	{
		MinVer:   newVer(1, 21, 0),
		MaxVer:   newVer(1, 22, 999),
		FileName: "reflect/type.go",
		Lines: []string{
			"func ifaceIndir(t *abi.Type) bool",
		},
	},
	{
		MinVer:   newVer(1, 23, 0),
		FileName: "reflect/badlinkname.go",
		Lines: []string{
			"//go:linkname unusedIfaceIndir reflect.ifaceIndir",
		},
	},
	{
		MaxVer:   newVer(1, 20, 999),
		FileName: "reflect/value.go",
		Lines: []string{
			"func unsafe_New(*rtype) unsafe.Pointer",
			"func unsafe_NewArray(*rtype, int) unsafe.Pointer",
			"func typedmemmove(t *rtype, dst, src unsafe.Pointer)",
			"func typedslicecopy(elemType *rtype, dst, src unsafeheader.Slice) int",
			"func maplen(m unsafe.Pointer) int",
		},
	},
	{
		MinVer:   newVer(1, 21, 0),
		FileName: "reflect/value.go",
		Lines: []string{
			"func unsafe_New(*abi.Type) unsafe.Pointer",
			"func unsafe_NewArray(*abi.Type, int) unsafe.Pointer",
			"func typedmemmove(t *abi.Type, dst, src unsafe.Pointer)",
			"func typedslicecopy(t *abi.Type, dst, src unsafeheader.Slice) int",
			"func maplen(m unsafe.Pointer) int",
		},
	},
}
