// +build !linux no_cgo_utime

package wheel

import (
	_ "unsafe"
)

//go:linkname Nanotime runtime.nanotime
func Nanotime() int64

//go:linkname Usleep runtime.usleep
func Usleep(usec uint32)

func Utick(usec uint32, f func() bool) {
	for {
		Usleep(usec)
		if done := f(); done {
			break
		}
	}
}
