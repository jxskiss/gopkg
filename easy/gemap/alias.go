package gemap

import "github.com/jxskiss/gopkg/v2/easy/ezmap"

// Map is a map of string key and interface{} value.
// It provides many useful methods to work with map[string]interface{}.
//
// Deprecated: this package has been renamed to ezmap.
type Map = ezmap.Map

// SafeMap wraps a Map with a RWMutex to provide concurrent safety.
//
// Deprecated: this package has been renamed to ezmap.
type SafeMap = ezmap.SafeMap

// NewMap returns a new initialized Map.
//
// Deprecated: this package has been renamed to ezmap.
func NewMap() Map {
	return ezmap.NewMap()
}

// NewSafeMap returns a new initialized SafeMap.
//
// Deprecated: this package has been renamed to ezmap.
func NewSafeMap() *SafeMap {
	return ezmap.NewSafeMap()
}
