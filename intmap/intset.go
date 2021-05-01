package intmap

import "unsafe"

const setFillFactor = 0.6

// Set is a hash set data structure optimized for int64 values.
//
// Set is not safe to use concurrently and it should be created by calling
// NewSet, using uninitialized zero Set will cause panic.
//
// The fill factor used for Set is 0.6. A Set will grow as needed.
type Set struct {
	data    []int64
	dataptr unsafe.Pointer

	threshold int
	size      int
	mask      uint64

	hasFreeKey bool
}

// NewSet returns a Set with 8 as initial capacity.
// The Set will grow as needed.
func NewSet() *Set {
	return newSet(8)
}

func newSet(size int) *Set {
	capacity := arraySize(size, setFillFactor)
	threshold := calcThreshold(capacity, setFillFactor)
	data := make([]int64, capacity)
	return &Set{
		data:      data,
		dataptr:   unsafe.Pointer(&data[0]),
		threshold: threshold,
		mask:      uint64(capacity - 1),
	}
}

// Size returns size of the set.
func (s *Set) Size() int {
	return s.size
}

// get uses pointer arithmetic to eliminate slice bounds checking,
// it will be inlined into the callers.
func (s *Set) get(ptr uint64) *int64 {
	return (*int64)(unsafe.Pointer(uintptr(s.dataptr) + uintptr(ptr*8)))
}

// Has tells whether a value is in the set.
func (s *Set) Has(elem int64) bool {
	if elem == FREE_KEY {
		return s.hasFreeKey
	}

	// manually inline phiMix to help inlining
	h := uint64(elem) * INT_PHI
	ptr := h ^ (h >> 16)

	for {
		ptr &= s.mask
		// manually inline getK and getV to help inlining
		e := *(*int64)(unsafe.Pointer(uintptr(s.dataptr) + uintptr(ptr*8)))
		if e == elem {
			return true
		}
		if e == FREE_KEY {
			return false
		}
		ptr += 1
	}
}

// Add adds a value to the set.
func (s *Set) Add(elem int64) {
	if elem == FREE_KEY {
		if !s.hasFreeKey {
			s.hasFreeKey = true
			s.size++
			return
		}
		return
	}

	ptr := phiMix(elem)
	for {
		ptr &= s.mask
		e := *s.get(ptr)
		if e == FREE_KEY {
			*s.get(ptr) = elem
			if s.size >= s.threshold {
				s.rehash()
			} else {
				s.size++
			}
			return
		}
		if e == elem {
			return
		}
		ptr += 1
	}
}

func (s *Set) rehash() {
	newCapacity := len(s.data) * 2
	s.threshold = calcThreshold(newCapacity, setFillFactor)
	s.mask = uint64(newCapacity - 1)

	data := s.data
	s.data = make([]int64, newCapacity)
	s.dataptr = unsafe.Pointer(&s.data[0])
	if s.hasFreeKey {
		s.size = 1
	} else {
		s.size = 0
	}

	var i int64
COPY:
	for i = 0; i < int64(len(data)); i++ {
		e := data[i]
		if e == FREE_KEY {
			continue
		}

		// Manually inline the Set function to avoid unnecessary calculation.
		ptr := phiMix(e)
		for {
			ptr &= s.mask
			if *s.get(ptr) == FREE_KEY {
				*s.get(ptr) = e
				s.size++
				continue COPY
			}
			ptr += 1
		}
	}
}

// Delete removes a value from the set.
func (s *Set) Delete(elem int64) {
	if elem == FREE_KEY {
		if s.hasFreeKey {
			s.hasFreeKey = false
			s.size--
		}
		return
	}

	ptr := phiMix(elem)
	for {
		ptr &= s.mask
		e := *s.get(ptr)
		if e == elem {
			s.shiftElements(ptr)
			s.size--
		}
		if e == FREE_KEY {
			return
		}
		ptr += 1
	}
}

func (s *Set) shiftElements(pos uint64) uint64 {
	var last, slot uint64
	var e int64
	for {
		last = pos
		pos = last + 1
		for {
			pos &= s.mask
			e = *s.get(pos)
			if e == FREE_KEY {
				*s.get(last) = FREE_KEY
				return last
			}

			slot = phiMix(e) & s.mask
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
		*(s.get(last)) = *s.get(pos)
	}
}

// Slice returns a slice of all values stored in the set.
func (s *Set) Slice() []int64 {
	elems := make([]int64, 0, s.size)
	if s.hasFreeKey {
		elems = append(elems, FREE_KEY)
	}
	for _, e := range s.data {
		if e != FREE_KEY {
			elems = append(elems, e)
		}
	}
	return elems
}

// Diff returns a new Set which contains values in s but not in other.
func (s *Set) Diff(other *Set) *Set {
	res := newSet(s.size)
	if s.hasFreeKey {
		if !other.hasFreeKey {
			res.hasFreeKey = true
		}
	}
	for _, e := range s.data {
		if e != FREE_KEY && !other.Has(e) {
			res.Add(e)
		}
	}
	return res
}

// Intersect returns a new Set which contains values in both s and other.
func (s *Set) Intersect(other *Set) *Set {
	// loop over the smaller set
	small, large := s, other
	if s.size > other.size {
		small, large = other, s
	}

	res := newSet(small.size)
	if s.hasFreeKey && other.hasFreeKey {
		res.hasFreeKey = true
		res.size++
	}
	for _, e := range small.data {
		if e != FREE_KEY && large.Has(e) {
			res.Add(e)
		}
	}
	return res
}

// Union returns a new Set which contains values either in s or other.
func (s *Set) Union(other *Set) *Set {
	size := max(s.size, other.size)
	res := newSet(size)

	for _, e := range s.data {
		if e != FREE_KEY {
			res.Add(e)
		}
	}
	for _, e := range other.data {
		if e != FREE_KEY {
			res.Add(e)
		}
	}
	return res
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
