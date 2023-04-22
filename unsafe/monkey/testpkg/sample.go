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

type testObj struct {
}

func (p *testObj) Value() int {
	return 0
}

func NewTestObj() interface {
	Value() int
} {
	return &testObj{}
}
