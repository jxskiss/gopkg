package gemap

import "sync"

// SafeMap wraps a Map with a RWMutex to provide concurrent safety.
type SafeMap struct {
	sync.RWMutex
	Map
}

// NewSafeMap returns a new initialized SafeMap.
func NewSafeMap() *SafeMap {
	return &SafeMap{Map: make(Map)}
}
