// +build !darwin,!netbsd,!freebsd,!openbsd,!dragonfly,!linux no_cgo_usleep

package wheel

import "time"

func Usleep(usec uint) {
	time.Sleep(time.Duration(usec) * time.Microsecond)
}
