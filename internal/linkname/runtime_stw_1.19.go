//go:build gc && go1.19 && !go1.21

package linkname

import _ "unsafe"

// Runtime_stopTheWorld links to runtime.stopTheWorld.
// It stops all P's from executing goroutines, interrupting all goroutines
// at GC safe points and records reason as the reason for the stop.
// On return, only the current goroutine's P is running.
//
//go:nosplit
func Runtime_stopTheWorld() {
	runtime_stopTheWorld("unknown")
}

//go:linkname Runtime_startTheWorld runtime.startTheWorld
//go:nosplit
func Runtime_startTheWorld()

//go:linkname runtime_stopTheWorld runtime.stopTheWorld
//go:nosplit
func runtime_stopTheWorld(reason string)
