package testpkg

//go:noinline
func A() string {
	return a()
}

//go:noinline
func a() string {
	return "testpkg.a"
}
