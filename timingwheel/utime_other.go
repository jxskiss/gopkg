// +build !darwin,!netbsd,!freebsd,!openbsd,!dragonfly,!linux

package wheel

import "time"

func Usleep(usec uint) {
	time.Sleep(time.Duration(usec) * time.Microsecond)
}

func Utick(usec uint, f func() bool) {
	d := time.Duration(usec) * time.Microsecond
	timer := time.NewTimer(d)
	for {
		timer.Reset(d)
		<-timer.C
		if done := f(); done {
			break
		}
	}
}
