package intmap

// go build -gcflags=-m=2 ./
// go build -gcflags="-d=ssa/check_bce/debug=1" ./
// go tool compile -S ./intintmap.go > ./_intintmap.s

import (
	"math"
	"unsafe"
)

// INT_PHI is for scrambling the keys.
const INT_PHI = 0x9E3779B9

// FREE_KEY is the 'free' key.
const FREE_KEY = 0

func phiMix(x int64) uint64 {
	h := x * INT_PHI
	return uint64(h ^ (h >> 16))
}

func nextPowerOfTwo(x int) int {
	if x == 0 {
		return 1
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16

	return x + 1
}

func arraySize(exp int, fill float64) int {
	s := int(math.Ceil(float64(exp) / fill))
	s = nextPowerOfTwo(s)
	if s < 2 {
		s = 2
	}
	return s
}

// Map is a hash map data structure optimized for int64 key values.
//
// Map should be created by calling New, usage of uninitialized
// zero Map will cause panic.
type Map struct {
	data    []Entry
	dataptr unsafe.Pointer // &data[0], helps to eliminate slice bounds checking

	fillFactor float64
	threshold  int    // we will resize a map once it reaches this threshold
	size       int    // the map's size
	mask       uint64 // capacity - 1

	hasFreeKey bool
	freeVal    int64
}

// Entry represents a key value pair in a Map.
type Entry struct {
	K, V int64
}

// New returns a map initialized with the stated size and fill factor.
// The Map will grow as needed.
func New(size int, fillFactor float64) *Map {
	if fillFactor <= 0 || fillFactor >= 1 {
		panic("fill factor must be in (0, 1)")
	}
	if size <= 0 {
		panic("size must be positive")
	}

	capacity := arraySize(size, fillFactor)
	data := make([]Entry, capacity)
	return &Map{
		data:       data,
		dataptr:    unsafe.Pointer(&data[0]),
		fillFactor: fillFactor,
		threshold:  int(math.Floor(float64(capacity) * fillFactor)),
		mask:       uint64(capacity - 1),
	}
}

// Size returns size of the map.
func (m *Map) Size() int {
	return m.size
}

// getK uses pointer arithmetic to eliminate slice bounds checking,
// it will be inlined into the callers.
func (m *Map) getK(ptr uint64) *int64 {
	return (*int64)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr*16)))
}

// getV uses pointer arithmetic to eliminate slice bounds checking,
// it will be inlined into the callers.
func (m *Map) getV(ptr uint64) *int64 {
	return (*int64)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr*16) + 8))
}

// Get returns the value if the key is found.
// It will be inlined into the callers
func (m *Map) Get(key int64) (int64, bool) {
	if key == FREE_KEY {
		if m.hasFreeKey {
			return m.freeVal, true
		}
		return 0, false
	}

	// manually inline phiMix to help inlining
	h := uint64(key) * INT_PHI
	ptr := h ^ (h >> 16)

	for {
		ptr &= m.mask
		// manually inline getK and getV to help inlining
		k := *(*int64)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr*16)))
		if k == key {
			return *(*int64)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr*16) + 8)), true
		}
		if k == FREE_KEY {
			return 0, false
		}
		ptr += 1
	}
}

// Has tells whether a key is found in the map.
// It will be inlined into the callers.
func (m *Map) Has(key int64) bool {
	if key == FREE_KEY {
		return m.hasFreeKey
	}

	// manually inline phiMix to help inlining
	h := uint64(key) * INT_PHI
	ptr := h ^ (h >> 16)

	for {
		ptr &= m.mask
		// manually inline getK to help inlining
		k := *(*int64)(unsafe.Pointer(uintptr(m.dataptr) + uintptr(ptr*16)))
		if k == FREE_KEY {
			return true
		}
		if k == key {
			return true
		}
		ptr += 1
	}
}

// Set adds or updates key with value to the map.
func (m *Map) Set(key, val int64) {
	if key == FREE_KEY {
		if !m.hasFreeKey {
			m.size++
		}
		m.hasFreeKey = true
		m.freeVal = val
		return
	}

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

func (m *Map) rehash() {
	newCapacity := len(m.data) * 2
	m.threshold = int(math.Floor(float64(newCapacity/2) * m.fillFactor))
	m.mask = uint64(newCapacity - 1)

	data := m.data
	m.data = make([]Entry, newCapacity)
	m.dataptr = unsafe.Pointer(&m.data[0])
	if m.hasFreeKey {
		m.size = 1
	} else {
		m.size = 0
	}

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

// Delete deletes a key and it's value from the map.
func (m *Map) Delete(key int64) {
	if key == FREE_KEY {
		if m.hasFreeKey {
			m.hasFreeKey = false
			m.size--
		}
		return
	}

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

// shiftKeys shifts entries with the same hash.
func (m *Map) shiftKeys(pos uint64) uint64 {
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

// Keys returns a slice of all keys stored in the map.
func (m *Map) Keys() []int64 {
	keys := make([]int64, 0, m.size)
	if m.hasFreeKey {
		keys = append(keys, FREE_KEY)
	}
	data := m.data
	for i := 0; i < len(data); i++ {
		if data[i].K == FREE_KEY {
			continue
		}
		keys = append(keys, data[i].K)
	}
	return keys
}

// Items returns a slice of all items stored in the map.
func (m *Map) Items() []Entry {
	items := make([]Entry, 0, m.size)
	if m.hasFreeKey {
		items = append(items, Entry{FREE_KEY, m.freeVal})
	}
	data := m.data
	for i := 0; i < len(data); i++ {
		if data[i].K == FREE_KEY {
			continue
		}
		items = append(items, data[i])
	}
	return items
}

// Clone returns a deep copy of the the map.
func (m *Map) Clone() *Map {
	data := make([]Entry, m.size)
	copy(data, m.data)
	newMap := &Map{
		data:       data,
		dataptr:    unsafe.Pointer(&data[0]),
		fillFactor: m.fillFactor,
		threshold:  m.threshold,
		size:       m.size,
		mask:       m.mask,
		hasFreeKey: m.hasFreeKey,
		freeVal:    m.freeVal,
	}
	return newMap
}
