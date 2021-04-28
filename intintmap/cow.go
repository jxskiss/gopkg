package intintmap

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// COWMap provides a lockless copy-on-write Map to optimize read-heavy
// workload, while write requests can be very little.
// When Set, Delete are called, the underlying Map will be copied.
//
// COWMap also embeds a sync.Mutex, which can be used optionally to lock
// the map to prevent unnecessary concurrent calculation. When lock is
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
// The underlying Map will grow as needed.
func NewCOWMap(fillFactor float64) *COWMap {
	m := New(8, fillFactor)
	return &COWMap{m: unsafe.Pointer(m)}
}

// SetMap stores the given map as the underlying map.
// Since each write operation will copy the map, write operations are
// considerably expensive, if there are many write operations, you may
// prepare a Map in batch mode, then use this method to set it to a COWMap.
func (m *COWMap) SetMap(map_ *Map) {
	atomic.StorePointer(&m.m, unsafe.Pointer(map_))
}

func (m *COWMap) getMap() *Map {
	return (*Map)(atomic.LoadPointer(&m.m))
}

// Size returns size of the map.
func (m *COWMap) Size() int {
	return m.getMap().size
}

// Get returns the value if the key is found.
func (m *COWMap) Get(key int64) (int64, bool) {
	return m.getMap().Get(key)
}

// Has tells whether a key is found in the COWMap.
func (m *COWMap) Has(key int64) bool {
	return m.getMap().Has(key)
}

// Set adds or updates key with value to the COWMap, if the key value
// is not present in the underlying Map, it will copy the Map and
// set the key value to the copy, then swap to the new Map using atomic
// operation.
func (m *COWMap) Set(key, val int64) {
	mm := m.getMap()
	if v, ok := mm.Get(key); ok && v == val {
		return
	}
	newMap := mm.Clone()
	newMap.Set(key, val)
	atomic.StorePointer(&m.m, unsafe.Pointer(newMap))
}

// Delete deletes a key and it's value from the COWMap, if the key
// presents in the underlying Map, it will copy the Map and
// delete the key value from the copy, then swap the new Map using
// atomic operation.
func (m *COWMap) Delete(key int64) {
	mm := m.getMap()
	if mm.Has(key) {
		newMap := mm.Clone()
		newMap.Delete(key)
		atomic.StorePointer(&m.m, unsafe.Pointer(newMap))
	}
}

// Keys returns the keys presented in the COWMap.
func (m *COWMap) Keys() []int64 {
	return m.getMap().Keys()
}

// Items returns all items stored in the COWMap.
func (m *COWMap) Items() []Entry {
	return m.getMap().Items()
}
