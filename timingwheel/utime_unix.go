// +build darwin netbsd freebsd openbsd dragonfly no_cgo_utime

package wheel

import "syscall"

func Usleep(usec uint) {
	spec := syscall.Timespec{
		Nsec: int64(usec * 1000),
	}
	syscall.Nanosleep(&spec, &spec)
}

func Utick(usec uint, f func() bool) {
	spec := syscall.Timespec{
		Nsec: int64(usec * 1000),
	}
	for {
		syscall.Nanosleep(&spec, &spec)
		if done := f(); done {
			break
		}
	}
}
