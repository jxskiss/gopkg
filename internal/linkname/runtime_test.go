package linkname

import "testing"

func compileRuntimeFunctions() {
	call(Runtime_memclrNoHeapPointers)
	call(Runtime_fastrand)
	call(Runtime_fastrandn)
	call(Runtime_fastrand64)
	call(Runtime_procPin)
	call(Runtime_procUnpin)
	call(Pid)
	call(Runtime_stopTheWorld)
	call(Runtime_startTheWorld)
	call(Runtime_memhash8)
	call(Runtime_memhash16)
	call(Runtime_stringHash)
	call(Runtime_bytesHash)
	call(Runtime_int32Hash)
	call(Runtime_int64Hash)
	call(Runtime_f32hash)
	call(Runtime_f64hash)
	call(Runtime_c64hash)
	call(Runtime_c128hash)
	call(Runtime_efaceHash)
	call(Runtime_typehash)
	call(Runtime_activeModules)
	call(Runtime_readUnaligned32)
	call(Runtime_readUnaligned64)
	call(Runtime_sysAlloc)
	call(Runtime_sysFree)
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

func TestRuntime_fastrandn(t *testing.T) {
	var n uint32 = 100
	var sum uint32
	for i := 0; i < 10; i++ {
		got := Runtime_fastrandn(n)
		sum += got
		if got >= n {
			t.Errorf("fastrandn got value %d > n", got)
		}
	}
	if sum == 0 {
		t.Errorf("fastrand got all zero values")
	}
}
