package intmap

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	typemapFillFactor = 0.6
	ptrsize           = unsafe.Sizeof(uintptr(0))
)

// TypeMap provides a lockless copy-on-write map mainly to use for type
// information cache, such as runtime generated encoders and decoders.
// TypeMap is safe to use concurrently, when SetByUintptr, SetByType are
// called, the underlying map will be copied.
//
// The fill factor used for TypeMap is 0.6. A TypeMap will grow as needed.
type TypeMap struct {
	l sync.Mutex
	m unsafe.Pointer // *interfaceMap
}

// NewTypeMap returns a new TypeMap with 8 as initial capacity.
func NewTypeMap() *TypeMap {
	size := 8
	imap := newInterfaceMap(size, typemapFillFactor)
	return &TypeMap{m: unsafe.Pointer(imap)}
}

// Size returns size of the map.
func (m *TypeMap) Size() int {
	return (*interfaceMap)(atomic.LoadPointer(&m.m)).size
}

// GetByUintptr returns value for the the given uintptr key.
// If key is not found in the map, it returns nil.
func (m *TypeMap) GetByUintptr(key uintptr) interface{} {
	return (*interfaceMap)(atomic.LoadPointer(&m.m)).Get(int64(key))
}

// GetByType returns value for the given reflect.Type.
// If key is not found in the map, it returns nil.
func (m *TypeMap) GetByType(key reflect.Type) interface{} {
	// type iface { tab  *itab, data unsafe.Pointer }
	typeptr := (*(*[2]uintptr)(unsafe.Pointer(&key)))[1]
	return (*interfaceMap)(atomic.LoadPointer(&m.m)).Get(int64(typeptr))
}

// SetByUintptr adds or updates value to the map using uintptr key.
// If the key value is not present in the underlying map, it will copy the
// map and add the key value to the copy, then swap to the new map using
// atomic operation.
func (m *TypeMap) SetByUintptr(key uintptr, val interface{}) {
	m.l.Lock()
	defer m.l.Unlock()
	imap := (*interfaceMap)(atomic.LoadPointer(&m.m))
	if v := imap.Get(int64(key)); v == val {
		return
	}
	newMap := imap.Copy()
	newMap.Set(int64(key), val)
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
