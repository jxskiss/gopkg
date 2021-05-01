package intintmap

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// COWMap provides a lockless copy-on-write map to optimize read-heavy
// workload, while write requests can be very little.
// When Set, Delete are called, the underlying map will be copied.
//
// COWMap also embeds a sync.Mutex, which can be used optionally to lock
// the map to prevent unnecessary concurrent copying. When lock is
// held, you may check the map again to see whether the target element has
// already been set or changed.
//
// COWMap should be created by calling NewCOWMap, usage of uninitialized
// zero COWMap will case panic.
type COWMap struct {
	sync.Mutex
	m unsafe.Pointer // *Map
}

// NewCOWMap creates a new COWMap using the stated fill factor.
// The underlying map will grow as needed.
func NewCOWMap(fillFactor float64) *COWMap {
	m := New(8, fillFactor)
	return &COWMap{m: unsafe.Pointer(m)}
}

// UseMap stores the given map as the underlying map.
// Since each write operation will copy the map, write operations are
// considerably expensive, if there are many write operations, you may
// prepare a Map in batch mode and tells COWMap to use it.
func (m *COWMap) UseMap(map_ *Map) {
	atomic.StorePointer(&m.m, unsafe.Pointer(map_))
}

// Underlying returns the current underlying Map of the COWMap.
// The returned Map is not safe for concurrent read and write.
func (m *COWMap) Underlying() *Map {
	return (*Map)(atomic.LoadPointer(&m.m))
}

// Size returns size of the map.
func (m *COWMap) Size() int {
	return (*Map)(atomic.LoadPointer(&m.m)).Size()
}

// Get returns the value if the key is found.
func (m *COWMap) Get(key int64) (int64, bool) {
	return (*Map)(atomic.LoadPointer(&m.m)).Get(key)
}

// Has tells whether a key is found in the map.
func (m *COWMap) Has(key int64) bool {
	return (*Map)(atomic.LoadPointer(&m.m)).Has(key)
}

// Set adds or updates key with value to the map, if the key value
// is not present in the underlying map, it will copy the map and
// add the key value to the copy, then swap to the new map using atomic
// operation.
func (m *COWMap) Set(key, val int64) {
	mm := (*Map)(atomic.LoadPointer(&m.m))
	if v, ok := mm.Get(key); ok && v == val {
		return
	}
	newMap := mm.Clone()
	newMap.Set(key, val)
	atomic.StorePointer(&m.m, unsafe.Pointer(newMap))
}

// Delete deletes a key and it's value from the map, if the key presents
// in the underlying map, it will copy the map and delete the key value
// from the copy, then swap the new map using atomic operation.
func (m *COWMap) Delete(key int64) {
	mm := (*Map)(atomic.LoadPointer(&m.m))
	if mm.Has(key) {
		newMap := mm.Clone()
		newMap.Delete(key)
		atomic.StorePointer(&m.m, unsafe.Pointer(newMap))
	}
}

// Keys returns the keys presented in the map.
func (m *COWMap) Keys() []int64 {
	return (*Map)(atomic.LoadPointer(&m.m)).Keys()
}

// Items returns all items stored in the map.
func (m *COWMap) Items() []Entry {
	return (*Map)(atomic.LoadPointer(&m.m)).Items()
}
