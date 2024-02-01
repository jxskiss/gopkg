package linkname

import (
	"sync"
	"testing"
	"time"
)

func compileRuntimeFunctions() {
	call(Runtime_fastrand)
	call(Runtime_fastrand64)
	call(Runtime_memhash32)
	call(Runtime_memhash64)
	call(Runtime_stringHash)
	call(Runtime_bytesHash)
	call(Runtime_efaceHash)
	call(Runtime_typehash)
	call(Runtime_activeModules)
}

func TestRuntime_fastrand(t *testing.T) {
	var sum uint32
	for i := 0; i < 10; i++ {
		sum += Runtime_fastrand()
	}
	if sum == 0 {
		t.Errorf("fastrand got all zero values")
	}
}

func TestRuntime_fastrand64(t *testing.T) {
	var sum uint64
	for i := 0; i < 10; i++ {
		sum += Runtime_fastrand64()
	}
	if sum == 0 {
		t.Errorf("fastrand64 got all zero values")
	}
}

func TestPid(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			pid := Pid()
			t.Logf("Pid got %d", pid)
		}()
	}
	wg.Wait()
}

var runtimeSourceCode = []SourceCodeTestCase{
	{
		MaxVer:   newVer(1, 21, 999),
		FileName: "runtime/stubs.go",
		Lines: []string{
			"func fastrand() uint32",
			"func fastrand64() uint64",
		},
	},
	{
		MinVer:   newVer(1, 22, 0),
		FileName: "runtime/rand.go",
		Lines: []string{
			"func rand() uint64",
		},
	},
	{
		FileName: "runtime/proc.go",
		Lines: []string{
			"func procPin() int",
			"func procUnpin()",
		},
	},
	{
		FileName: "runtime/alg.go",
		Lines: []string{
			"func memhash32(p unsafe.Pointer, h uintptr) uintptr",
			"func memhash64(p unsafe.Pointer, h uintptr) uintptr",
			"func stringHash(s string, seed uintptr) uintptr",
			"func bytesHash(b []byte, seed uintptr) uintptr",
			"func efaceHash(i any, seed uintptr) uintptr",
			"func typehash(t *_type, p unsafe.Pointer, h uintptr) uintptr",
		},
	},
	{
		FileName: "runtime/symtab.go",
		Lines: []string{
			"func activeModules() []*moduledata",
		},
	},
}
