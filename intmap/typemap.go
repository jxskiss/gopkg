package intmap

import (
	"math"
	"reflect"
	"sync/atomic"
	"unsafe"
)

const (
	typemapFillFactor = 0.6
	ptrsize           = unsafe.Sizeof(uintptr(0))
)

// TypeMap provides a lockless copy-on-write map mainly to use for type
// information cache, such as runtime generated encoders and decoders.
//
// The fill factor used for TypeMap is 0.6. A TypeMap will grow as needed.
type TypeMap struct {
	m unsafe.Pointer // *typemap
}

// NewTypeMap returns a new TypeMap with 8 as initial capacity.
func NewTypeMap() *TypeMap {
	capacity := 8
	tmap := newTypemap(capacity)
	return &TypeMap{m: unsafe.Pointer(tmap)}
}

// Size returns size of the map.
func (m *TypeMap) Size() int {
	return (*typemap)(atomic.LoadPointer(&m.m)).size
}

// GetByUintptr returns value for the the given uintptr key.
// If key is not found in the map, it returns nil.
func (m *TypeMap) GetByUintptr(key uintptr) interface{} {
	return (*typemap)(atomic.LoadPointer(&m.m)).Get(key)
}

// GetByType returns value for the given reflect.Type.
// If key is not found in the map, it returns nil.
func (m *TypeMap) GetByType(key reflect.Type) interface{} {
	// type iface { tab  *itab, data unsafe.Pointer }
	typeptr := (*(*[2]uintptr)(unsafe.Pointer(&key)))[1]
	return (*typemap)(atomic.LoadPointer(&m.m)).Get(typeptr)
}

// SetByUintptr adds or updates value to the map using uintptr key.
// If the key value is not present in the underlying map, it will copy the
// map and add the key value to the copy, then swap to the new map using
// atomic operation.
func (m *TypeMap) SetByUintptr(key uintptr, val interface{}) {
	tmap := (*typemap)(atomic.LoadPointer(&m.m))
	if v := tmap.Get(key); v == val {
		return
	}
	newMap := tmap.Copy()
	newMap.Set(key, val)
	atomic.StorePointer(&m.m, unsafe.Pointer(newMap))
}

// SetByType adds or updates value to the map using reflect.Type key.
// If the key value is not present in the underlying map, it will copy the
// map and add the key value to the copy, then swap to the new map using
// atomic operation.
func (m *TypeMap) SetByType(key reflect.Type, val interface{}) {
	// type iface { tab  *itab, data unsafe.Pointer }
	typeptr := (*(*[2]uintptr)(unsafe.Pointer(&key)))[1]
	m.SetByUintptr(typeptr, val)
}

func newTypemap(capacity int) *typemap {
	if capacity&(capacity-1) != 0 {
		panic("typemap capacity must be power of two")
	}
	threshold := int(math.Floor(float64(capacity) * typemapFillFactor))
	mask := capacity - 1
	data := make([]typeEntry, capacity)
	return &typemap{
		data:      data,
		dataptr:   unsafe.Pointer(&data[0]),
		threshold: threshold,
		size:      0,
		mask:      uint64(mask),
	}
}

type typemap struct {
	data    []typeEntry
	dataptr unsafe.Pointer

	threshold int
	size      int
	mask      uint64
}

type typeEntry struct {
	K uintptr
	V interface{}
}

// getK helps to eliminate slice bounds checking
func (m *typemap) getK(ptr uint64) *uintptr {
	return (*uintptr)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize))
}

// getV helps to eliminate slice bounds checking
func (m *typemap) getV(ptr uint64) *interface{} {
	return (*interface{})(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize + ptrsize))
}

// Get returns the value if the key is found, else it returns nil.
// It will be inlined by the compiler.
func (m *typemap) Get(key uintptr) interface{} {
	// manually inline phiMix to help inlining
	h := uint64(key) * INT_PHI
	ptr := h ^ (h >> 16)

	for {
		ptr &= m.mask
		// manually inline m.getK and m.getV
		k := *(*uintptr)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize))
		if k == key {
			return *(*interface{})(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize + ptrsize))
		}
		if k == 0 {
			return nil
		}
		ptr += 1
	}
}

// Set adds or updates key with value to the typemap.
func (m *typemap) Set(key uintptr, val interface{}) {
	ptr := phiMix(int64(key))
	for {
		ptr &= m.mask
		k := *m.getK(ptr)
		if k == 0 {
			*m.getK(ptr) = key
			*m.getV(ptr) = val
			m.size++
			return
		}
		if k == key {
			*m.getV(ptr) = val
			return
		}
		ptr += 1
	}
}

// Copy returns a copy of a typemap, if the map's size triggers it's
// threshold, the new map's capacity will be twice of the old.
func (m *typemap) Copy() *typemap {
	capacity := cap(m.data)
	if m.size >= m.threshold {
		capacity *= 2
	}
	newMap := newTypemap(capacity)
	for _, e := range m.data {
		if e.K == 0 {
			continue
		}
		newMap.Set(e.K, e.V)
	}
	return newMap
}
