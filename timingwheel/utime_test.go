package wheel

import "testing"

func TestNanotime(t *testing.T) {
	println(Nanotime())
}

func TestUsleep(t *testing.T) {
	Usleep(1000)
	println("after 1000 micro seconds")
}
