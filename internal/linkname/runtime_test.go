package linkname

func compileRuntimeFunctions() {
	call(Runtime_memclrNoHeapPointers)
	call(Runtime_fastrand)
	call(Runtime_fastrandn)
	call(Runtime_procPin)
	call(Runtime_procUnpin)
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
}
