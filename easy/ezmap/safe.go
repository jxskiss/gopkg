package ezmap

import "sync"

// SafeMap wraps a Map with a RWMutex to provide concurrent safety.
//
// Note that user must explicitly acquire and release the lock to be
// concurrent safe (i.e. to avoid data race).
type SafeMap struct {
	sync.RWMutex
	Map
}

// NewSafeMap returns a new initialized SafeMap.
func NewSafeMap() *SafeMap {
	return &SafeMap{Map: make(Map)}
}
