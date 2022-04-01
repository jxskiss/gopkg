package easy

// DiffMaps return a map which contains elements which present in m, but
// not present in others.
//
// If inplace = true, it does not allocate new memory, instead it modifies
// m in-place and returns m.
// Otherwise, it allocates a new map and m won't be modified.
//
// If length of m is zero, it returns nil.
func DiffMaps[M ~map[K]V, K comparable, V any](inplace bool, m M, others ...M) M {
	if len(m) == 0 {
		return nil
	}

	if inplace {
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

	// allocate a new map
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
func MergeMapsTo[M ~map[K]V, K comparable, V any](dst M, others ...M) M {
	if dst == nil {
		dst = make(M)
	}
	for _, m := range others {
		for k, v := range m {
			dst[k] = v
		}
	}
	return dst
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