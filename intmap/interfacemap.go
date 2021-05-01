package intmap

import (
	"math"
	"unsafe"
)

// TODO: doc concurrent safety.

// InterfaceMap is a hash map data structure optimized for int64 keys and
// interface values.
// InterfaceMap should be created by calling NewInterfaceMap, usage of
// uninitialized zero InterfaceMap will cause panic.
type InterfaceMap struct {
	m *interfaceMap

	hasFreeKey bool
	freeVal    interface{}
}

// InterfaceEntry represents a key value pair in a InterfaceMap or TypeMap.
type InterfaceEntry struct {
	K int64
	V interface{}
}

// NewInterfaceMap returns a map initialized with the stated size and
// fill factor. The InterfaceMap will grow as needed.
func NewInterfaceMap(size int, fillFactor float64) *InterfaceMap {
	if fillFactor <= 0 || fillFactor >= 1 {
		panic("fill factor must be in (0, 1)")
	}
	if size <= 0 {
		panic("size must be positive")
	}

	capacity := arraySize(size, fillFactor)
	imap := newInterfaceMap(capacity, fillFactor)
	return &InterfaceMap{m: imap}
}

// Size returns the size of the map.
func (m *InterfaceMap) Size() int {
	size := m.m.size
	if m.hasFreeKey {
		size++
	}
	return size
}

// Get returns the value if teh key is found in the map.
func (m *InterfaceMap) Get(key int64) interface{} {
	if key == FREE_KEY {
		if m.hasFreeKey {
			return m.freeVal
		}
		return nil
	}
	return m.m.Get(key)
}

// Has tells whether a key is found in the map.
func (m *InterfaceMap) Has(key int64) bool {
	if key == FREE_KEY {
		if m.hasFreeKey {
			return true
		}
		return false
	}
	return m.m.Has(key)
}

// Set adds or updates key with value to the map.
func (m *InterfaceMap) Set(key int64, val interface{}) {
	if key == FREE_KEY {
		m.freeVal = val
		return
	}
	m.m.SetRehash(key, val)
}

// Delete deletes a key and it's value from the map.
func (m *InterfaceMap) Delete(key int64) {
	if key == FREE_KEY {
		if m.hasFreeKey {
			m.hasFreeKey = false
			m.freeVal = nil
		}
		return
	}
	m.m.Delete(key)
}

// Keys returns a slice of all keys stored in the map.
func (m *InterfaceMap) Keys() []int64 {
	keys := m.m.Keys()
	if m.hasFreeKey {
		keys = append(keys, FREE_KEY)
	}
	return keys
}

// Items returns a slice of all items stored in the map.
func (m *InterfaceMap) Items() []InterfaceEntry {
	items := m.m.Items()
	if m.hasFreeKey {
		items = append(items, InterfaceEntry{
			K: FREE_KEY,
			V: m.freeVal,
		})
	}
	return items
}

// ------------------------- interfaceMap ------------------------- //

func newInterfaceMap(capacity int, fillFactor float64) *interfaceMap {
	if capacity&(capacity-1) != 0 {
		panic("interfaceMap capacity must be power of two")
	}
	threshold := int(math.Floor(float64(capacity) * fillFactor))
	mask := capacity - 1
	data := make([]InterfaceEntry, capacity)
	return &interfaceMap{
		data:       data,
		dataptr:    unsafe.Pointer(&data[0]),
		fillFactor: fillFactor,
		threshold:  threshold,
		size:       0,
		mask:       uint64(mask),
	}
}

type interfaceMap struct {
	data    []InterfaceEntry
	dataptr unsafe.Pointer

	fillFactor float64
	threshold  int
	size       int
	mask       uint64
}

// getK helps to eliminate slice bounds checking
func (m *interfaceMap) getK(ptr uint64) *int64 {
	return (*int64)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize))
}

// getV helps to eliminate slice bounds checking
func (m *interfaceMap) getV(ptr uint64) *interface{} {
	return (*interface{})(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize + ptrsize))
}

// Get returns the value if the key is found, else it returns nil.
// It will be inlined by the compiler.
func (m *interfaceMap) Get(key int64) interface{} {
	// manually inline phiMix to help inlining
	h := uint64(key) * INT_PHI
	ptr := h ^ (h >> 16)

	for {
		ptr &= m.mask
		// manually inline m.getK and m.getV
		k := *(*int64)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize))
		if k == key {
			return *(*interface{})(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize + ptrsize))
		}
		if k == 0 {
			return nil
		}
		ptr += 1
	}
}

// Has tells whether a key is found in the map.
// It will be inlined into the callers.
func (m *interfaceMap) Has(key int64) bool {
	// manually inline phiMix to help inlining
	h := uint64(key) * INT_PHI
	ptr := h ^ (h >> 16)

	for {
		ptr &= m.mask
		// manually inline m.getK and m.getV
		k := *(*int64)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr)*3*ptrsize))
		if k == key {
			return true
		}
		if k == 0 {
			return false
		}
		ptr += 1
	}
}

// Set adds or updates key with value to the interfaceMap.
func (m *interfaceMap) Set(key int64, val interface{}) {
	ptr := phiMix(key)
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

func (m *interfaceMap) SetRehash(key int64, val interface{}) {
	ptr := phiMix(key)
	for {
		ptr &= m.mask
		k := *m.getK(ptr)
		if k == FREE_KEY {
			*m.getK(ptr) = key
			*m.getV(ptr) = val
			if m.size >= m.threshold {
				m.rehash()
			} else {
				m.size++
			}
			return
		}
		if k == key {
			*m.getV(ptr) = val
			return
		}
		ptr += 1
	}
}

func (m *interfaceMap) rehash() {
	newCapacity := len(m.data) * 2
	m.threshold = int(math.Floor(float64(newCapacity/2) * m.fillFactor))
	m.mask = uint64(newCapacity - 1)

	data := m.data
	m.data = make([]InterfaceEntry, newCapacity)
	m.dataptr = unsafe.Pointer(&m.data[0])

	var i int64
COPY:
	for i = 0; i < int64(len(data)); i++ {
		e := data[i]
		if e.K == FREE_KEY {
			continue
		}

		// Manually inline the Set function to avoid unnecessary calculation.
		ptr := phiMix(e.K)
		for {
			ptr &= m.mask
			k := *m.getK(ptr)
			if k == FREE_KEY {
				*m.getK(ptr) = e.K
				*m.getV(ptr) = e.V
				m.size++
				continue COPY
			}
			ptr += 1
		}
	}
}

func (m *interfaceMap) Delete(key int64) {
	ptr := phiMix(key)
	for {
		ptr &= m.mask
		k := *m.getK(ptr)
		if k == key {
			m.shiftKeys(ptr)
			m.size--
			return
		}
		if k == FREE_KEY {
			return
		}
		ptr += 1
	}
}

func (m *interfaceMap) shiftKeys(pos uint64) uint64 {
	var last, slot uint64
	var k int64
	for {
		last = pos
		pos = last + 1
		for {
			pos &= m.mask
			k = *m.getK(pos)
			if k == FREE_KEY {
				*m.getK(last) = FREE_KEY
				*m.getV(last) = nil
				return last
			}

			slot = phiMix(k) & m.mask
			if last <= pos {
				if last >= slot || slot > pos {
					break
				}
			} else {
				if last >= slot && slot > pos {
					break
				}
			}
			pos += 1
		}
		*(m.getK(last)) = *m.getK(pos)
		*(m.getV(last)) = *m.getV(pos)
	}
}

// Copy returns a copy of a interfaceMap, if the map's size triggers it's
// threshold, the new map's capacity will be twice of the old.
func (m *interfaceMap) Copy() *interfaceMap {
	capacity := cap(m.data)
	if m.size >= m.threshold {
		capacity *= 2
	}
	newMap := newInterfaceMap(capacity, m.fillFactor)
	for _, e := range m.data {
		if e.K == 0 {
			continue
		}
		newMap.Set(e.K, e.V)
	}
	return newMap
}

func (m *interfaceMap) Keys() []int64 {
	keys := make([]int64, 0, m.size+1)
	data := m.data
	for i := 0; i < len(data); i++ {
		if data[i].K == FREE_KEY {
			continue
		}
		keys = append(keys, data[i].K)
	}
	return keys
}

func (m *interfaceMap) Items() []InterfaceEntry {
	items := make([]InterfaceEntry, 0, m.size+1)
	data := m.data
	for i := 0; i < len(data); i++ {
		if data[i].K == FREE_KEY {
			continue
		}
		items = append(items, data[i])
	}
	return items
}
