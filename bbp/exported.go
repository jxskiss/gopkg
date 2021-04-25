// Package bbp provides efficient byte buffer pools with
// anti-memory-waste protection.
//
// Byte buffers acquired from this package may be put back to the pool,
// but they do not need to; if they are returned, they will be recycled
// and reused, otherwise they will be garbage collected as usual.
//
// Within this package, `Get`, `Set` and all `Pool` instances share the same
// underlying sized byte slice pools. The byte buffers provided by this
// package has minimum and maximum limit (see `MinBufSize` and `MaxBufSize`),
// byte slice with size not in the range will be allocated directly from
// the operating system, and won't be recycled for reuse.
package bbp

// Get returns a byte buffer from the pool with specified length and capacity.
// The returned byte buffer's capacity is of at least 8.
//
// The returned byte buffer can be put back to the pool by calling Put(buf),
// which may be reused later. This reduces memory allocations and GC pressure.
func Get(length int, capacity ...int) *Buffer {
	if len(capacity) > 1 {
		panic("too many arguments to bbp.Get")
	}
	buf := getBuffer()
	buf.B = get(length, capacity...)
	return buf
}

// Grow returns a new byte buffer from the pool which guarantees it's
// at least of specified capacity.
//
// If capacity is not specified, the returned slice is at least twice
// of the given buf slice.
// The returned byte buffer's capacity is always power of two, which
// can be put back to the pool after usage.
//
// The old buf will be put into the pool for reusing, so it mustn't be
// touched after calling this function, otherwise data races will occur.
func Grow(buf []byte, capacity ...int) []byte {
	if len(capacity) > 1 {
		panic("too many arguments to bbp.Grow")
	}
	l, c := len(buf), cap(buf)
	newCap := 2 * l
	if len(capacity) > 0 {
		newCap = capacity[0]
	}
	if c >= newCap {
		return buf
	}
	return grow(buf, newCap)
}

// Put puts back a byte buffer to the pool for reusing.
//
// The buf mustn't be touched after retuning it to the pool.
// Otherwise data races will occur.
func Put(buf *Buffer) {
	if !buf.noReuse {
		put(buf.B)
	}
	buf.B = nil
	buf.noReuse = false
	bpool.Put(buf)
}

// PutSlice puts back a byte slice which is obtained from function Grow.
//
// The byte slice mustn't be touched after returning it to the pool.
// Otherwise data races will occur.
func PutSlice(buf []byte) { put(buf) }
