package easy

import (
	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// DiffMaps returns a new map which contains elements which present in m,
// but not present in others.
//
// If length of m is zero, it returns nil.
func DiffMaps[M ~map[K]V, K comparable, V any](m M, others ...M) M {
	out := make(M)
	for k, v := range m {
		found := false
		for _, b := range others {
			if _, ok := b[k]; ok {
				found = true
				break
			}
		}
		if !found {
			out[k] = v
		}
	}
	return out
}

// DiffMapsInplace removes elements that present in others from m.
func DiffMapsInplace[M ~map[K]V, K comparable, V any](m M, others ...M) M {
	for k := range m {
		for _, m1 := range others {
			if _, ok := m1[k]; ok {
				delete(m, k)
				break
			}
		}
	}
	return m
}

// FilterMaps iterates the given maps, it calls predicate(k, v) for each
// key value in the maps and returns a new map of key value pairs for
// which predicate(k, v) returns true.
func FilterMaps[M ~map[K]V, K comparable, V any](predicate func(k K, v V) bool, maps ...M) M {
	if len(maps) == 0 {
		return nil
	}
	out := make(M, len(maps[0]))
	for _, x := range maps {
		for k, v := range x {
			if predicate(k, v) {
				out[k] = v
			}
		}
	}
	return out
}

// MergeMaps returns a new map containing all key values present in given maps.
func MergeMaps[M ~map[K]V, K comparable, V any](maps ...M) M {
	var length int
	for _, m := range maps {
		length += len(m)
	}
	dst := make(M, length)
	for _, m := range maps {
		for k, v := range m {
			dst[k] = v
		}
	}
	return dst
}

// MergeMapsTo adds key values present in others to the dst map.
// If dst is a nil map, it creates a new map and returns it.
func MergeMapsTo[M ~map[K]V, K comparable, V any](dst M, others ...M) M {
	if len(others) == 0 {
		return dst
	}
	if dst == nil {
		dst = make(M, len(others[0]))
	}
	for _, m := range others {
		for k, v := range m {
			dst[k] = v
		}
	}
	return dst
}

// MergeMapsToPtr is similar to MergeMapsTo, but it accepts a pointer as dst,
// if dst points to a nil map, it creates a new map and assigns it to dst.
// If dst is a nil pointer, it panics.
func MergeMapsToPtr[M ~map[K]V, K comparable, V any](dst *M, others ...M) {
	if dst == nil {
		panic("easy.MergeMapsToPtr: dst must not be nil")
	}
	*dst = MergeMapsTo(*dst, others...)
}

// Keys returns the keys of the map m.
// The keys will be in an indeterminate order.
//
// Optionally, a filter function can be given to make it returning
// only keys for which filter(k, v) returns true.
func Keys[M ~map[K]V, K comparable, V any](m M, filter ...func(K, V) bool) []K {
	var f func(K, V) bool
	if len(filter) > 0 {
		f = filter[0]
	}
	keys := make([]K, 0, len(m))
	if f == nil {
		for k := range m {
			keys = append(keys, k)
		}
	} else {
		for k, v := range m {
			if f(k, v) {
				keys = append(keys, k)
			}
		}
	}
	return keys
}

// Values returns the values of the map m.
// The values will be in an indeterminate order.
//
// Optionally, a filter function can be given to make it returning
// only values for which filter(k, v) returns true.
func Values[M ~map[K]V, K comparable, V any](m M, filter ...func(K, V) bool) []V {
	var f func(K, V) bool
	if len(filter) > 0 {
		f = filter[0]
	}
	values := make([]V, 0, len(m))
	if f == nil {
		for _, v := range m {
			values = append(values, v)
		}
	} else {
		for k, v := range m {
			if f(k, v) {
				values = append(values, v)
			}
		}
	}
	return values
}

// CopyMap copies a map to be a new one.
// optionalSize optionally specifies the size of the new map.
func CopyMap[M ~map[K]V, K comparable, V any](m M, optionalSize ...int) M {
	copySize := len(m)
	if len(optionalSize) > 0 && optionalSize[0] > copySize {
		copySize = optionalSize[0]
	}
	out := make(M, copySize)
	for k, v := range m {
		out[k] = v
	}
	return out
}

// SplitMap splits a large map to batches, it returns a slice
// of type []M whose elements are subset of the given map.
func SplitMap[M ~map[K]V, K comparable, V any](m M, batchSize int) []M {
	if len(m) == 0 {
		return nil
	}
	if len(m) <= batchSize {
		return []M{m}
	}

	cnt := (len(m) + batchSize - 1) / batchSize
	out := make([]M, cnt)
	for i := range out {
		out[i] = make(M, batchSize)
	}
	i := 0
	for k, v := range m {
		out[i/batchSize][k] = v
		i++
	}
	return out
}

// SplitMapStable splits a large map to batches, it returns a slice
// of type []M whose elements are subset of the given map.
// The elements in each sub map are stable between multiple calls
// with same input.
func SplitMapStable[M ~map[K]V, K constraints.Ordered, V any](m M, batchSize int) []M {
	if len(m) == 0 {
		return nil
	}
	if len(m) <= batchSize {
		return []M{m}
	}

	keys := Keys(m)
	Sort(keys)

	cnt := (len(m) + batchSize - 1) / batchSize
	out := make([]M, cnt)
	for i := range out {
		out[i] = make(M, batchSize)
	}
	total, step := len(m), batchSize
	idx := 0
	for i, j := 0, step; idx < cnt && i < total; i, j = i+step, j+step {
		if j > total {
			j = total
		}
		for _, k := range keys[i:j] {
			out[idx][k] = m[k]
		}
		idx++
	}
	return out
}
