package sqlutil

import (
	"database/sql/driver"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

// NewLazyBinary creates a new lazy binary wrapper, delaying the
// unmarshalling work until it is first needed.
func NewLazyBinary(raw []byte) LazyBinary {
	return LazyBinary{raw: raw}
}

// LazyBinary is a lazy wrapper around a binary value, where the underlying
// object will be unmarshalled the first time it is needed and cached.
// It implements sql/driver.Valuer and sql.Scanner.
//
// LazyBinary provides same concurrency safety as []byte, it's safe for
// concurrent read, but not safe for concurrent write or read/write.
//
// See types_test.go for example usage.
type LazyBinary struct {
	raw []byte
	obj unsafe.Pointer // *lazyobj
}

type lazyobj struct {
	mu   sync.Mutex
	data interface{}
	err  error
}

// Value implements driver.Valuer interface.
func (p LazyBinary) Value() (driver.Value, error) {
	return p.raw, nil
}

// Scan implements sql.Scanner interface.
func (p *LazyBinary) Scan(src interface{}) error {
	if src != nil {
		// NOTE
		// We MUST copy the src byte slice here, database/sql.Scanner says:
		//
		// Reference types such as []byte are only valid until the next call to Scan
		// and should not be retained. Their underlying memory is owned by the driver.
		// If retention is necessary, copy their values before the next call to Scan.

		b, ok := src.([]byte)
		if !ok {
			return fmt.Errorf("sqlutil: wants []byte but got %T", src)
		}
		tmp := make([]byte, len(b))
		copy(tmp, b)
		p.raw = tmp
		atomic.StorePointer(&p.obj, nil)
	}
	return nil
}

// Unmarshaler is a function which unmarshalls data from a byte slice.
type Unmarshaler func([]byte) (interface{}, error)

// GetBytes returns the underlying byte slice.
func (p *LazyBinary) GetBytes() []byte {
	return p.raw
}

// Get returns the underlying data wrapped by the LazyBinary wrapper,
// if the data has not been unmarshalled, it will be unmarshalled using
// the provided unmarshalFunc.
// The unmarshalling work will do only once, the result data and error
// will be cached and reused for further calling.
func (p *LazyBinary) Get(unmarshalFunc Unmarshaler) (interface{}, error) {
	obj, created := p.getobj()
	defer obj.mu.Unlock()
	if created {
		obj.data, obj.err = unmarshalFunc(p.raw)
		return obj.data, obj.err
	}
	obj.mu.Lock()
	return obj.data, obj.err
}

// Set sets the data and marshaled bytes to the LazyBinary wrapper.
// If the param data is nil, the underlying cache will be removed.
func (p *LazyBinary) Set(b []byte, data interface{}) {
	p.raw = b
	if data == nil {
		atomic.StorePointer(&p.obj, nil)
		return
	}
	obj, created := p.getobj()
	if !created {
		obj.mu.Lock()
	}
	obj.data = data
	obj.err = nil
	obj.mu.Unlock()
}

func (p *LazyBinary) getobj() (*lazyobj, bool) {
	ptr := atomic.LoadPointer(&p.obj)
	if ptr != nil {
		return (*lazyobj)(ptr), false
	}
	tmp := &lazyobj{}
	tmp.mu.Lock()
	swapped := atomic.CompareAndSwapPointer(&p.obj, nil, unsafe.Pointer(tmp))
	if swapped {
		return tmp, true
	}
	ptr = atomic.LoadPointer(&p.obj)
	return (*lazyobj)(ptr), false
}
