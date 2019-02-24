// +build darwin netbsd freebsd openbsd dragonfly linux

// See: https://github.com/golang/go/issues/27707

package wheel

// #include <unistd.h>
import "C"

func Usleep(usec uint) {
	C.usleep(C.useconds_t(usec))
}
