package gemap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Map is a map of string key and interface{} value.
// It provides many useful methods to work with map[string]interface{}.
type Map map[string]interface{}

// SafeMap wraps a Map with a RWMutex to provide concurrent safety.
type SafeMap struct {
	sync.RWMutex
	Map
}

// NewMap returns a new initialized Map.
func NewMap() Map {
	return make(Map)
}

// NewSafeMap returns a new initialized SafeMap.
func NewSafeMap() *SafeMap {
	return &SafeMap{Map: make(Map)}
}

// MarshalJSON implements the json.Marshaler interface.
func (p Map) MarshalJSON() ([]byte, error) {
	x := map[string]interface{}(p)
	return json.Marshal(x)
}

// MarshalJSONPretty returns its marshaled data as `[]byte` with
// indentation using two spaces.
func (p Map) MarshalJSONPretty() ([]byte, error) {
	x := map[string]interface{}(p)
	return json.MarshalIndent(x, "", "  ")
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *Map) UnmarshalJSON(data []byte) error {
	x := (*map[string]interface{})(p)
	return json.Unmarshal(data, x)
}

// Set is used to store a new key/value pair exclusively in the map.
// It also lazily initializes the map if it was not used previously.
func (p *Map) Set(key string, value interface{}) {
	if *p == nil {
		*p = make(Map)
	}
	(*p)[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (p Map) Get(key string) (value interface{}, exists bool) {
	value, exists = p[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (p Map) MustGet(key string) interface{} {
	if val, ok := p[key]; ok {
		return val
	}
	panic(fmt.Sprintf("key %q not exists", key))
}

// GetString returns the value associated with the key as a string.
func (p Map) GetString(key string) string {
	v := p[key]
	if val, ok := v.(string); ok {
		return val
	}
	if val, ok := v.([]byte); ok {
		return string(val)
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
	val, _ := p[key].(bool)
	return val
}

// GetInt returns the value associated with the key as an integer.
func (p Map) GetInt(key string) int {
	val, ok := p[key]
	if ok {
		switch v := val.(type) {
		case int:
			return v
		case json.Number:
			num, _ := v.Int64()
			return int(num)
		case string:
			num, _ := strconv.ParseInt(v, 10, 64)
			return int(num)
		}
	}
	return 0
}

// GetInt64 returns the value associated with the key as an int64.
func (p Map) GetInt64(key string) int64 {
	val, ok := p[key]
	if ok {
		switch v := val.(type) {
		case int64:
			return v
		case json.Number:
			num, _ := v.Int64()
			return num
		case string:
			num, _ := strconv.ParseInt(v, 10, 64)
			return num
		}
	}
	return 0
}

// GetInt32 returns the value associated with the key as an int32.
func (p Map) GetInt32(key string) int32 {
	val, ok := p[key]
	if ok {
		switch v := val.(type) {
		case int32:
			return v
		case json.Number:
			num, _ := v.Int64()
			return int32(num)
		case string:
			num, _ := strconv.ParseInt(v, 10, 64)
			return int32(num)
		}
	}
	return 0
}

// GetFloat64 returns the value associated with the key as a float64.
func (p Map) GetFloat64(key string) float64 {
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
	val, _ := p[key].(time.Duration)
	return val
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
		}
	}
	return nil
}

// GetSlice returns the value associated with the key as []interface{}.
// It returns nil if key does not present in Map or the value's type
// is not []interface{}.
func (p Map) GetSlice(key string) []interface{} {
	val, _ := p[key].([]interface{})
	return val
}

// GetMap returns the value associated with the key as a Map (map[string]interface{}).
func (p Map) GetMap(key string) Map {
	val, ok := p[key]
	if ok {
		switch val := val.(type) {
		case Map:
			return val
		case map[string]interface{}:
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
func (p Map) Iterate(fn func(k string, v interface{}) int) {
	for k, v := range p {
		if fn(k, v) != 0 {
			return
		}
	}
}
