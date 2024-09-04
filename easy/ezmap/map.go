package ezmap

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"
)

// Map is a map of string key and any value.
// It provides many useful methods to work with map[string]any.
type Map map[string]any

// NewMap returns a new initialized Map.
func NewMap() Map {
	return make(Map)
}

// MarshalJSON implements the json.Marshaler interface.
func (p Map) MarshalJSON() ([]byte, error) {
	x := map[string]any(p)
	return json.Marshal(x)
}

// MarshalJSONPretty returns its marshaled data as `[]byte` with
// indentation using two spaces.
func (p Map) MarshalJSONPretty() ([]byte, error) {
	x := map[string]any(p)
	return json.MarshalIndent(x, "", "  ")
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *Map) UnmarshalJSON(data []byte) error {
	x := (*map[string]any)(p)
	return json.Unmarshal(data, x)
}

// MarshalYAML implements the [yaml.Marshaler] interface.
func (p Map) MarshalYAML() (any, error) {
	return (map[string]any)(p), nil
}

// UnmarshalYAML implements the [yaml.Unmarshaler] interface.
func (p *Map) UnmarshalYAML(value *yaml.Node) error {
	x := (*map[string]any)(p)
	return value.Decode(x)
}

// Size returns the number of elements in the map.
func (p Map) Size() int {
	return len(p)
}

// Set is used to store a new key/value pair exclusively in the map.
// It also lazily initializes the map if it was not used previously.
func (p *Map) Set(key string, value any) {
	if *p == nil {
		*p = make(Map)
	}
	(*p)[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (p Map) Get(key string) (value any, exists bool) {
	value, exists = p[key]
	return
}

// GetOr returns the value for the given key if it exists in the map,
// else it returns the default value.
func (p Map) GetOr(key string, defaultVal any) (value any) {
	value, exists := p[key]
	if exists {
		return value
	}
	return defaultVal
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (p Map) MustGet(key string) any {
	if val, ok := p[key]; ok {
		return val
	}
	panic(fmt.Sprintf("key %q not exists", key))
}

// GetString returns the value associated with the key as a string.
func (p Map) GetString(key string) string {
	val, ok := p[key]
	if ok {
		return cast.ToString(val)
	}
	return ""
}

// GetBytes returns the value associated with the key as bytes.
func (p Map) GetBytes(key string) []byte {
	v := p[key]
	if val, ok := v.([]byte); ok {
		return val
	}
	if val, ok := v.(string); ok {
		return []byte(val)
	}
	return nil
}

// GetBool returns the value associated with the key as a boolean value.
func (p Map) GetBool(key string) bool {
	val, ok := p[key]
	if ok {
		return cast.ToBool(val)
	}
	return false
}

// GetInt returns the value associated with the key as an int.
func (p Map) GetInt(key string) int {
	val, ok := p[key]
	if ok {
		return cast.ToInt(val)
	}
	return 0
}

// GetInt32 returns the value associated with the key as an int32.
func (p Map) GetInt32(key string) int32 {
	val, ok := p[key]
	if ok {
		return cast.ToInt32(val)
	}
	return 0
}

// GetInt64 returns the value associated with the key as an int64.
func (p Map) GetInt64(key string) int64 {
	val, ok := p[key]
	if ok {
		return cast.ToInt64(val)
	}
	return 0
}

// GetUint returns the value associated with the key as an uint.
func (p Map) GetUint(key string) uint {
	val, ok := p[key]
	if ok {
		return cast.ToUint(val)
	}
	return 0
}

// GetUint32 returns the value associated with the key as an uint32.
func (p Map) GetUint32(key string) uint32 {
	val, ok := p[key]
	if ok {
		return cast.ToUint32(val)
	}
	return 0
}

// GetUint64 returns the value associated with the key as an uint64.
func (p Map) GetUint64(key string) uint64 {
	val, ok := p[key]
	if ok {
		return cast.ToUint64(val)
	}
	return 0
}

// GetFloat returns the value associated with the key as a float64.
func (p Map) GetFloat(key string) float64 {
	val, ok := p[key]
	if ok {
		switch v := val.(type) {
		case float64:
			return v
		case json.Number:
			num, _ := v.Float64()
			return num
		case string:
			num, _ := strconv.ParseFloat(v, 64)
			return num
		}
		typ := reflect.TypeOf(val)
		switch typ.Kind() {
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(val).Float()
		}
		if intVal := p.GetInt(key); intVal != 0 {
			return float64(intVal)
		}
	}
	return 0
}

// GetTime returns the value associated with the key as time.
func (p Map) GetTime(key string) time.Time {
	val, _ := p[key].(time.Time)
	return val
}

// GetDuration returns the value associated with the key as a duration.
func (p Map) GetDuration(key string) time.Duration {
	val, ok := p[key]
	if ok {
		switch v := val.(type) {
		case time.Duration:
			return v
		case int64:
			return time.Duration(v)
		case string:
			d, err := time.ParseDuration(v)
			if err == nil {
				return d
			}
		}
	}
	return 0
}

// GetInt64s returns the value associated with the key as a slice of int64.
func (p Map) GetInt64s(key string) []int64 {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case []int64:
			return val
		}
	}
	return nil
}

// GetInt32s returns the value associated with the key as a slice of int32.
func (p Map) GetInt32s(key string) []int32 {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case []int32:
			return val
		}
	}
	return nil
}

// GetStrings returns the value associated with the key as a slice of strings.
func (p Map) GetStrings(key string) []string {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case []string:
			return val
		default:
			return cast.ToStringSlice(val)
		}
	}
	return nil
}

// GetSlice returns the value associated with the key as a slice.
// It returns nil if key does not present in Map or the value's type
// is not a slice.
func (p Map) GetSlice(key string) []any {
	val, ok := p[key]
	if !ok || reflect.TypeOf(val).Kind() != reflect.Slice {
		return nil
	}
	rv := reflect.ValueOf(val)
	out := make([]any, rv.Len())
	for i, n := 0, rv.Len(); i < n; i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out
}

// GetSliceElem returns the ith element of a slice associated with key.
// It returns nil if key does not present in Map or the value's type
// is not a slice, or i exceeds the slice's length.
func (p Map) GetSliceElem(key string, i int) any {
	val, ok := p[key]
	if !ok || reflect.TypeOf(val).Kind() != reflect.Slice {
		return nil
	}
	rv := reflect.ValueOf(val)
	if i < rv.Len() {
		return rv.Index(i).Interface()
	}
	return nil
}

// GetMap returns the value associated with the key as a Map (map[string]any).
func (p Map) GetMap(key string) Map {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Map:
			return val
		case map[string]any:
			return val
		}
	}
	return nil
}

// GetStringMap returns the value associated with the key as a map of (map[string]string).
func (p Map) GetStringMap(key string) map[string]string {
	if val, ok := p[key].(map[string]string); ok {
		return val
	}
	return nil
}

// Iterate iterates the map in unspecified order, the given function fn
// will be called for each key value pair.
// The iteration can be aborted by returning a non-zero value from fn.
func (p Map) Iterate(fn func(k string, v any) int) {
	for k, v := range p {
		if fn(k, v) != 0 {
			return
		}
	}
}

// Merge merges key values from another map.
func (p *Map) Merge(other map[string]any) {
	if *p == nil {
		*p = make(Map)
	}
	for k, v := range other {
		(*p)[k] = v
	}
}
