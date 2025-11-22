package sqlutil

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	"google.golang.org/protobuf/proto"
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
	data any
	err  error
}

// Value implements driver.Valuer interface.
func (p LazyBinary) Value() (driver.Value, error) {
	return p.raw, nil
}

// Scan implements sql.Scanner interface.
func (p *LazyBinary) Scan(src any) error {
	if src != nil {
		// NOTE
		// We MUST copy the src byte slice here, database/sql.Scanner says:
		//
		// Reference types such as []byte are only valid until the next call to Scan
		// and should not be retained. Their underlying memory is owned by the driver.
		// If retention is necessary, copy their values before the next call to Scan.

		var b []byte
		switch tmp := src.(type) {
		case string:
			b = []byte(tmp)
		case []byte:
			b = make([]byte, len(tmp))
			copy(b, tmp)
		default:
			return fmt.Errorf("sqlutil.LazyBinary.Scan: want string/[]byte but got %T", src)
		}
		p.raw = b
		atomic.StorePointer(&p.obj, nil)
	}
	return nil
}

// Unmarshaler is a function which unmarshalls data from a byte slice.
type Unmarshaler func([]byte) (any, error)

// GetBytes returns the underlying byte slice.
func (p *LazyBinary) GetBytes() []byte {
	return p.raw
}

// Get returns the underlying data wrapped by the LazyBinary wrapper,
// if the data has not been unmarshalled, it will be unmarshalled using
// the provided unmarshalFunc.
// The unmarshalling work will do only once, the result data and error
// will be cached and reused for further calling.
func (p *LazyBinary) Get(unmarshalFunc Unmarshaler) (any, error) {
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
func (p *LazyBinary) Set(b []byte, data any) {
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

// SetBytes sets the underlying byte slice and updates the cached object to nil.
func (p *LazyBinary) SetBytes(b []byte) {
	p.raw = b
	atomic.StorePointer(&p.obj, nil)
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

// LazyProtobuf is a lazy wrapper around a protobuf message value,
// where the underlying object will be unmarshalled the first time it is
// needed and cached.
// Type T must be a protobuf message pointer type.
// It implements sql/driver.Valuer and sql.Scanner.
//
// LazyProtobuf provides same concurrency safety as []byte, it's safe for
// concurrent read, but not safe for concurrent write or read/write.
//
// See types_test.go for example usage.
type LazyProtobuf[T proto.Message] struct {
	lb LazyBinary
}

// Value implements driver.Valuer interface.
func (p LazyProtobuf[T]) Value() (driver.Value, error) {
	return p.lb.Value()
}

// Scan implements sql.Scanner interface.
func (p *LazyProtobuf[T]) Scan(src any) error {
	return p.lb.Scan(src)
}

// Get returns the underlying protobuf message value wrapped by the wrapper,
// if the message has not been unmarshalled, it will be unmarshalled.
// The unmarshalling work will do only once, the result object and error
// will be cached and reused for further calling.
func (p *LazyProtobuf[T]) Get() (T, error) {
	obj, err := p.lb.Get(func(buf []byte) (any, error) {
		var zero T
		result := reflect.New(reflect.TypeOf(zero).Elem()).Interface().(T)
		err := proto.Unmarshal(buf, result)
		return result, err
	})
	if err != nil {
		return *new(T), err
	}
	return obj.(T), nil
}

// Set sets the protobuf message value to the LazyProtobuf wrapper.
// If the param obj is nil, the underlying cache will be removed.
func (p *LazyProtobuf[T]) Set(obj T) error {
	buf, err := proto.Marshal(obj)
	if err != nil {
		return err
	}
	p.lb.Set(buf, obj)
	return nil
}

func (p *LazyProtobuf[T]) GetBytes() []byte {
	return p.lb.GetBytes()
}

func (p *LazyProtobuf[T]) SetBytes(b []byte) {
	p.lb.SetBytes(b)
}
