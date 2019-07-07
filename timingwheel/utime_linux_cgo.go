// +build linux,!no_cgo_utime

package wheel

// #include <unistd.h>
// #include <sys/timerfd.h>
import "C"

import (
	"syscall"
	"unsafe"
)

const (
	_CLOCK_MONOTONIC     = 0x1
	_SYS_TIMERFD_CREATE  = 283
	_SYS_TIMERFD_SETTIME = 286
	_SYS_TIMERFD_GETTIME = 287
)

//go:linkname Nanotime runtime.nanotime
func Nanotime() int64

func Usleep(usec uint32) {
	C.usleep(C.useconds_t(usec))
}

func Utick(usec uint32, f func() bool) {
	fd, _, errno := syscall.RawSyscall(_SYS_TIMERFD_CREATE, _CLOCK_MONOTONIC, 0, 0)
	if errno > 0 {
		panic(errno.Error())
	}
	defer syscall.Close(int(fd))
	timespec := C.struct_timespec{
		tv_nsec: C.long(usec * 1000),
	}
	timerspec := &C.struct_itimerspec{
		it_interval: timespec,
		it_value:    timespec,
	}
	_, _, errno = syscall.RawSyscall(_SYS_TIMERFD_SETTIME, fd, 0, uintptr(unsafe.Pointer(timerspec)))
	if errno > 0 {
		panic(errno.Error())
	}
	var buf [8]byte
	for {
		_, err := syscall.Read(int(fd), buf[:])
		if err != nil {
			panic(err)
		}
		if done := f(); done {
			break
		}
	}
}
