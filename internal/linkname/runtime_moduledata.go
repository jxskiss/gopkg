package linkname

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

var runtimeModuledataInfo struct {
	once    sync.Once
	initErr error

	// field offsets
	f_pclntable_offset uintptr
	f_ftab_offset      uintptr
	f_bad_offset       uintptr
	f_next_offset      uintptr
}

func init_runtime_moduledata_info() {
	info := &runtimeModuledataInfo
	mdType, err := GetReflectTypeByName("runtime.moduledata")
	if err != nil {
		info.initErr = err
		return
	}
	for _, x := range []struct {
		name string
		typ  string
		dst  *uintptr
	}{
		{"pclntable", "[]uint8", &runtimeModuledataInfo.f_pclntable_offset},
		{"ftab", "[]runtime.functab", &runtimeModuledataInfo.f_ftab_offset},
		{"bad", "bool", &runtimeModuledataInfo.f_bad_offset},
		{"next", "*runtime.moduledata", &runtimeModuledataInfo.f_next_offset},
	} {
		offset, err := getFieldOffset(mdType, x.name, x.typ)
		if err != nil {
			info.initErr = fmt.Errorf("get field %s offset: %w", x.name, err)
			return
		}
		*x.dst = offset
	}

	functabType, err := GetReflectTypeByName("runtime.functab")
	if err != nil {
		info.initErr = err
		return
	}
	if functabType.NumField() != 2 {
		info.initErr = fmt.Errorf("type runtime.functab not match")
		return
	}
	for _, fieldname := range []string{
		"entryoff",
		"funcoff",
	} {
		f1, _ := reflect.TypeOf(functab{}).FieldByName(fieldname)
		f2, ok := functabType.FieldByName(fieldname)
		if !ok {
			info.initErr = fmt.Errorf("runtime.functab field %s not found", fieldname)
			return
		}
		if f1.Type != f2.Type || f1.Offset != f2.Offset {
			info.initErr = fmt.Errorf("runtime.functab field %s not match", fieldname)
			return
		}
	}
}

// Runtime_moduledata is an opaque proxy type to runtime.moduledata.
type Runtime_moduledata struct {
	p unsafe.Pointer
}

func (p *Runtime_moduledata) Field_pclntable() []byte {
	return *(*[]byte)(unsafe.Pointer(uintptr(p.p) + runtimeModuledataInfo.f_pclntable_offset))
}

func (p *Runtime_moduledata) Field_ftab() []functab {
	return *(*[]functab)(unsafe.Pointer(uintptr(p.p) + runtimeModuledataInfo.f_ftab_offset))
}

func (p *Runtime_moduledata) Field_bad() bool {
	return *(*bool)(unsafe.Pointer(uintptr(p.p) + runtimeModuledataInfo.f_bad_offset))
}

func (p *Runtime_moduledata) Field_next() Runtime_moduledata {
	next := *(*unsafe.Pointer)(unsafe.Pointer(uintptr(p.p) + runtimeModuledataInfo.f_next_offset))
	return Runtime_moduledata{p: next}
}

// functab is a copy type of runtime.functab.
type functab struct {
	entryoff uint32 // relative to runtime.text
	funcoff  uint32
}

func (ft functab) Field_entryoff() uint32 { return ft.entryoff }
func (ft functab) Field_funcoff() uint32  { return ft.funcoff }

func getFieldOffset(t reflect.Type, fieldname, typename string) (uintptr, error) {
	f, ok := t.FieldByName(fieldname)
	if !ok {
		return 0, fmt.Errorf("field not found")
	}
	fieldTypeName := fmt.Sprintf("%v", f.Type)
	if fieldTypeName != typename {
		return 0, fmt.Errorf("field type not match, want %s but got %s", typename, fieldTypeName)
	}
	return f.Offset, nil
}
