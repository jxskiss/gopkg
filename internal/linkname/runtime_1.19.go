//go:build gc && go1.19 && !go1.22

package linkname

import "unsafe"

//go:linkname Runtime_fastrand runtime.fastrand
//go:nosplit
func Runtime_fastrand() uint32

//go:linkname Runtime_fastrand64 runtime.fastrand64
//go:nosplit
func Runtime_fastrand64() uint64

// -------- runtime moduledata --------

func Runtime_activeModules() []Runtime_moduledata {
	runtimeModuledataInfo.once.Do(init_runtime_moduledata_info)
	if err := runtimeModuledataInfo.initErr; err != nil {
		panic("runtime moduledata info not initialized: " + err.Error())
	}

	mdptrs := runtime_activeModules()
	out := make([]Runtime_moduledata, len(mdptrs))
	for i, ptr := range mdptrs {
		out[i] = Runtime_moduledata{ptr}
	}
	return out
}

//go:linkname runtime_activeModules runtime.activeModules
func runtime_activeModules() []unsafe.Pointer
