package bbp

import (
	"sync"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// NewObjectPool creates an ObjectPool for type T.
func NewObjectPool[T any]() ObjectPool[T] {
	var x T
	idx := indexGet(int(unsafe.Sizeof(x)))
	return ObjectPool[T]{
		cap:  1 << idx,
		pool: &sizedPools[idx],
	}
}

// ObjectPool is an object pool which uses the shared sized
// byte buffer pools to reuse memory.
type ObjectPool[T any] struct {
	cap  int
	pool *sync.Pool
}

// Get returns a new object of type *T from the pool.
func (a ObjectPool[T]) Get() *T {
	buf := a.pool.Get().([]byte)[:a.cap]
	// zero the memory, memclr
	for i := range buf {
		buf[i] = 0
	}
	h := *(*unsafeheader.Slice)(unsafe.Pointer(&buf))
	return (*T)(h.Data)
}

// Put puts back an object to the pool for reusing.
//
// The object mustn't be touched after passing to this method,
// otherwise undefined behavior happens.
func (a ObjectPool[T]) Put(x *T) {
	h := unsafeheader.Slice{
		Data: unsafe.Pointer(x),
		Len:  0,
		Cap:  a.cap,
	}
	buf := *(*[]byte)(unsafe.Pointer(&h))
	a.pool.Put(buf)
}
