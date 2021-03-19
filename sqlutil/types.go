package sqlutil

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/jxskiss/gopkg/serialize"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Bitmap represents a bitmap value, it implements sql/driver.Valuer and sql.Scanner.
// Bitmap provides Get, Set and Clear methods to manipulate the bitmap value.
type Bitmap int

func (b Bitmap) Value() (driver.Value, error) {
	return int64(b), nil
}

func (b *Bitmap) Scan(src interface{}) error {
	var tmp sql.NullInt64
	err := tmp.Scan(src)
	if err == nil {
		*b = Bitmap(tmp.Int64)
	}
	return err
}

func (b Bitmap) Get(mask int) bool {
	return int(b)&mask != 0
}

func (b *Bitmap) Set(mask int) {
	*b |= Bitmap(mask)
}

func (b *Bitmap) Clear(mask int) {
	*b &= ^Bitmap(mask)
}

var (
	emptyArray  = []byte{'[', ']'}
	emptyObject = []byte{'{', '}'}
	zeroBytes   = []byte{}
)

// JSONInt32s represents an integer array, it implements sql/driver.Valuer
// and sql.Scanner. It uses JSON to do serialization.
type JSONInt32s []int32

func (p JSONInt32s) Value() (driver.Value, error) {
	if len(p) == 0 {
		return emptyArray, nil
	}
	return json.Marshal(p)
}

func (p *JSONInt32s) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// JSONInt64s represents an integer array, it implements sql/driver.Valuer
// and sql.Scanner. It uses JSON to do serialization.
type JSONInt64s []int64

func (p JSONInt64s) Value() (driver.Value, error) {
	if len(p) == 0 {
		return emptyArray, nil
	}
	return json.Marshal(p)
}

func (p *JSONInt64s) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// JSONStrings represents a string array, it implements sql/driver.Valuer
// and sql.Scanner. It uses JSON to do serialization.
type JSONStrings []string

func (p JSONStrings) Value() (driver.Value, error) {
	if len(p) == 0 {
		return emptyArray, nil
	}
	return json.Marshal(p)
}

func (p *JSONStrings) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// JSONStringMap represents a map[string]string value, it implements
// sql/driver.Valuer and sql.Scanner. It uses JSON to do serialization.
type JSONStringMap map[string]string

func (p JSONStringMap) Value() (driver.Value, error) {
	if p == nil {
		return emptyObject, nil
	}
	return json.Marshal(p)
}

func (p *JSONStringMap) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// JSONDict represents a map[string]interface{} value, it implements
// sql/driver.Valuer and sql.Scanner. It uses JSON to do serialization.
type JSONDict map[string]interface{}

func (p JSONDict) Value() (driver.Value, error) {
	if p == nil {
		return emptyObject, nil
	}
	return json.Marshal(p)
}

func (p *JSONDict) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return json.Unmarshal(data, p)
	}
}

// PBInt32s represents an integer array, it implements sql/driver.Valuer
// and sql.Scanner. It uses binary serialization format (protobuf).
type PBInt32s []int32

func (p PBInt32s) Value() (driver.Value, error) {
	if len(p) == 0 {
		return zeroBytes, nil
	}
	return serialize.Int32List(p).MarshalProto()
}

func (p *PBInt32s) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return (*serialize.Int32List)(p).UnmarshalProto(data)
	}
}

// PBInt64s represents an integer array, it implements sql/driver.Valuer
// and sql.Scanner. It uses binary serialization format (protobuf).
type PBInt64s []int64

func (p PBInt64s) Value() (driver.Value, error) {
	if len(p) == 0 {
		return zeroBytes, nil
	}
	return serialize.Int64List(p).MarshalProto()
}

func (p *PBInt64s) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return (*serialize.Int64List)(p).UnmarshalProto(data)
	}
}

// PBStrings represents a string array, it implements sql/driver.Valuer
// and sql.Scanner. It uses binary serialization format (protobuf).
type PBStrings []string

func (p PBStrings) Value() (driver.Value, error) {
	if len(p) == 0 {
		return zeroBytes, nil
	}
	return serialize.StringList(p).MarshalProto()
}

func (p *PBStrings) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return (*serialize.StringList)(p).UnmarshalProto(data)
	}
}

// PBStringMap represents a map[string]string value, it implements
// sql/driver.Valuer and sql.Scanner. It uses binary serialization format (protobuf).
type PBStringMap map[string]string

func (p PBStringMap) Value() (driver.Value, error) {
	if len(p) == 0 {
		return zeroBytes, nil
	}
	return serialize.StringMap(p).MarshalProto()
}

func (p *PBStringMap) Scan(src interface{}) error {
	if data, ok := src.([]byte); !ok {
		return nil
	} else {
		return (*serialize.StringMap)(p).UnmarshalProto(data)
	}
}

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

func (p LazyBinary) Value() (driver.Value, error) {
	return p.raw, nil
}

func (p *LazyBinary) Scan(src interface{}) error {
	if src != nil {
		b, ok := src.([]byte)
		if !ok {
			return fmt.Errorf("sqlutil: wants []byte but got %T", src)
		}
		p.raw = b
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
