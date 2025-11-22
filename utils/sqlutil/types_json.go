package sqlutil

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/json"
)

//nolint:unused
var (
	null        = []byte("null")
	emptyObject = []byte("{}")
)

// JSONMap holds a map[string]any value, it implements
// sql/driver.Valuer, sql.Scanner, json.Marshaler, json.Unmarshaler.
// It uses JSON to do serialization.
//
// JSONMap embeds an ezmap.Map, thus all methods defined on ezmap.Map is also
// available from a JSONMap instance.
type JSONMap struct {
	ezmap.Map
}

// Value implements driver.Valuer interface.
func (p JSONMap) Value() (driver.Value, error) {
	if p.Map == nil {
		return emptyObject, nil
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(p.Map)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Scan implements sql.Scanner interface.
func (p *JSONMap) Scan(src any) error {
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = unsafeheader.StringToBytes(v)
	default:
		return fmt.Errorf("sqlutil.JSONMap.Scan: want []byte/string but got %T", src)
	}
	if bytes.Equal(data, null) {
		p.Map = nil
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(&p.Map)
}

func (p JSONMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Map)
}

func (p *JSONMap) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.Map)
}

// LazyJSON is a lazy wrapper around a JSON value, where the underlying
// object will be unmarshalled the first time it is needed and cached.
// Type T should be a struct type.
// It implements sql/driver.Valuer, sql.Scanner, json.Marshaler, json.Unmarshaler.
//
// LazyJSON provides same concurrency safety as []byte, it's safe for
// concurrent read, but not safe for concurrent write or read/write.
//
// See types_test.go for example usage.
type LazyJSON[T any] struct {
	lb LazyBinary
}

// Value implements driver.Valuer interface.
func (p LazyJSON[T]) Value() (driver.Value, error) {
	return p.lb.Value()
}

// Scan implements sql.Scanner interface.
func (p *LazyJSON[T]) Scan(src any) error {
	return p.lb.Scan(src)
}

// Get returns the cached object, or unmarshal it from the raw bytes
// if it is not cached yet.
func (p *LazyJSON[T]) Get() (*T, error) {
	if bytes.Equal(p.lb.GetBytes(), null) {
		return nil, nil
	}
	obj, err := p.lb.Get(func(buf []byte) (any, error) {
		var result T
		err := json.Unmarshal(buf, &result)
		return &result, err
	})
	if err != nil {
		return nil, err
	}
	return obj.(*T), nil
}

// Set marshals the object to JSON and save the JSON bytes
// together with the object.
func (p *LazyJSON[T]) Set(obj *T) error {
	if obj == nil {
		p.lb.Set(null, nil)
		return nil
	}
	buf, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	p.lb.Set(buf, obj)
	return nil
}

// GetBytes returns the raw JSON bytes.
func (p *LazyJSON[T]) GetBytes() []byte {
	return p.lb.GetBytes()
}

func (p *LazyJSON[T]) SetBytes(b []byte) {
	p.lb.SetBytes(b)
}

func (p LazyJSON[T]) MarshalJSON() ([]byte, error) {
	return p.GetBytes(), nil
}

func (p *LazyJSON[T]) UnmarshalJSON(data []byte) error {
	p.SetBytes(data)
	return nil
}
