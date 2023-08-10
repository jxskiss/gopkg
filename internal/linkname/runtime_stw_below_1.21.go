//go:build gc && !go1.21

package linkname

import _ "unsafe"

// Runtime_stopTheWorld links to runtime.stopTheWorld.
// It stops all P's from executing goroutines, interrupting all goroutines
// at GC safe points and records reason as the reason for the stop.
// On return, only the current goroutine's P is running.
//
//go:linkname Runtime_stopTheWorld runtime.stopTheWorld
func Runtime_stopTheWorld()

//go:linkname Runtime_startTheWorld runtime.startTheWorld
func Runtime_startTheWorld()
