package testpkg

import "math/rand"

//go:noinline
func A() string {
	_ = rand.Intn(10)
	return a()
}

//go:noinline
func a() string {
	_ = rand.Intn(10)
	return "testpkg.a"
}
