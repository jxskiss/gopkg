//go:build gc && go1.22

package linkname

import (
	"reflect"
	"runtime"
	"unsafe"
)

func Runtime_fastrand() uint32 {
	return uint32(Runtime_fastrand64())
}

//go:linkname Runtime_fastrand64 runtime.rand
//go:nosplit
func Runtime_fastrand64() uint64

// -------- runtime moduledata --------

func Runtime_activeModules() []Runtime_moduledata {
	runtimeModuledataInfo.once.Do(init_runtime_moduledata_info)
	if err := runtimeModuledataInfo.initErr; err != nil {
		panic("runtime moduledata info not initialized: " + err.Error())
	}

	pc := reflect.ValueOf(runtime.GOMAXPROCS).Pointer()
	fi := runtime_findfunc(pc)
	if fi.datap == nil {
		panic("cannot find the runtime moduledata")
	}

	var result []Runtime_moduledata
	md := Runtime_moduledata{p: fi.datap}
	seenPtrs := make(map[uintptr]bool)
	for ; md.p != nil; md = md.Field_next() {
		if seenPtrs[uintptr(md.p)] {
			break
		}
		if md.Field_bad() {
			// module failed to load and should be ignored
			continue
		}
		result = append(result, md)
	}
	return result
}

type funcInfo struct {
	_func unsafe.Pointer // *_func
	datap unsafe.Pointer // datap *moduledata
}

//go:linkname runtime_findfunc runtime.findfunc
func runtime_findfunc(pc uintptr) funcInfo
