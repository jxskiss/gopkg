package linkname

func compileUtilFunctions() {
	call(GetPid)
	call(Runtime_readUnaligned32)
	call(Runtime_readUnaligned64)
}
