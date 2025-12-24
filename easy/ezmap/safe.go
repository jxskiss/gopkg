package ezmap

import (
	"fmt"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// SafeMap wraps a Map with a RWMutex to provide concurrent safety.
// It's safe for concurrent use by multiple goroutines.
type SafeMap struct {
	mu   sync.RWMutex
	map_ Map //nolint:revive
}

// NewSafeMap returns a new initialized SafeMap.
func NewSafeMap() *SafeMap {
	return &SafeMap{map_: make(Map)}
}

// MarshalJSON implements the json.Marshaler interface.
func (p *SafeMap) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.MarshalJSON()
}

// MarshalJSONPretty returns its marshaled data as `[]byte` with
// indentation using two spaces.
func (p *SafeMap) MarshalJSONPretty() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.MarshalJSONPretty()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *SafeMap) UnmarshalJSON(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.map_.UnmarshalJSON(data)
}

// MarshalYAML implements the [yaml.Marshaler] interface.
func (p *SafeMap) MarshalYAML() (any, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.MarshalYAML()
}

// UnmarshalYAML implements the [yaml.Unmarshaler] interface.
func (p *SafeMap) UnmarshalYAML(value *yaml.Node) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.map_.UnmarshalYAML(value)
}

// Size returns the number of elements in the map.
func (p *SafeMap) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.map_)
}

// Set is used to store a new key/value pair exclusively in the map.
// It also lazily initializes the map if it was not used previously.
func (p *SafeMap) Set(key string, value any) {
	p.mu.Lock()
	p.map_.Set(key, value)
	p.mu.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (p *SafeMap) Get(key string) (value any, exists bool) {
	p.mu.RLock()
	value, exists = p.map_.Get(key)
	p.mu.RUnlock()
	return
}

// GetOr returns the value for the given key if it exists in the map,
// else it returns the default value.
func (p *SafeMap) GetOr(key string, defaultVal any) (value any) {
	p.mu.RLock()
	value = p.map_.GetOr(key, defaultVal)
	p.mu.RUnlock()
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (p *SafeMap) MustGet(key string) any {
	var (
		value  any
		exists bool
	)
	p.mu.RLock()
	value, exists = p.map_.Get(key)
	p.mu.RLock()
	if exists {
		return value
	}
	panic(fmt.Sprintf("key %q not exists", key))
}

// GetString returns the value associated with the key as a string.
func (p *SafeMap) GetString(key string) string {
	return getWithRLock(&p.mu, p.map_.GetString, key)
}

// GetBytes returns the value associated with the key as bytes.
func (p *SafeMap) GetBytes(key string) []byte {
	return getWithRLock(&p.mu, p.map_.GetBytes, key)
}

// GetBool returns the value associated with the key as a boolean value.
func (p *SafeMap) GetBool(key string) bool {
	return getWithRLock(&p.mu, p.map_.GetBool, key)
}

// GetInt returns the value associated with the key as an int.
func (p *SafeMap) GetInt(key string) int {
	return getWithRLock(&p.mu, p.map_.GetInt, key)
}

// GetInt32 returns the value associated with the key as an int32.
func (p *SafeMap) GetInt32(key string) int32 {
	return getWithRLock(&p.mu, p.map_.GetInt32, key)
}

// GetInt64 returns the value associated with the key as an int64.
func (p *SafeMap) GetInt64(key string) int64 {
	return getWithRLock(&p.mu, p.map_.GetInt64, key)
}

// GetUint returns the value associated with the key as an uint.
func (p *SafeMap) GetUint(key string) uint {
	return getWithRLock(&p.mu, p.map_.GetUint, key)
}

// GetUint32 returns the value associated with the key as an uint32.
func (p *SafeMap) GetUint32(key string) uint32 {
	return getWithRLock(&p.mu, p.map_.GetUint32, key)
}

// GetUint64 returns the value associated with the key as an uint64.
func (p *SafeMap) GetUint64(key string) uint64 {
	return getWithRLock(&p.mu, p.map_.GetUint64, key)
}

// GetFloat returns the value associated with the key as a float64.
func (p *SafeMap) GetFloat(key string) float64 {
	return getWithRLock(&p.mu, p.map_.GetFloat, key)
}

// GetTime returns the value associated with the key as time.
func (p *SafeMap) GetTime(key string) time.Time {
	return getWithRLock(&p.mu, p.map_.GetTime, key)
}

// GetDuration returns the value associated with the key as a duration.
func (p *SafeMap) GetDuration(key string) time.Duration {
	return getWithRLock(&p.mu, p.map_.GetDuration, key)
}

// GetInt64s returns the value associated with the key as a slice of int64.
func (p *SafeMap) GetInt64s(key string) []int64 {
	return getWithRLock(&p.mu, p.map_.GetInt64s, key)
}

// GetInt32s returns the value associated with the key as a slice of int32.
func (p *SafeMap) GetInt32s(key string) []int32 {
	return getWithRLock(&p.mu, p.map_.GetInt32s, key)
}

// GetStrings returns the value associated with the key as a slice of strings.
func (p *SafeMap) GetStrings(key string) []string {
	return getWithRLock(&p.mu, p.map_.GetStrings, key)
}

// GetSlice returns the value associated with the key as a slice.
// It returns nil if key does not present in Map or the value's type
// is not a slice.
func (p *SafeMap) GetSlice(key string) []any {
	return getWithRLock(&p.mu, p.map_.GetSlice, key)
}

// GetSliceElem returns the ith element of a slice associated with key.
// It returns nil if key does not present in Map or the value's type
// is not a slice, or i exceeds the slice's length.
func (p *SafeMap) GetSliceElem(key string, i int) any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.GetSliceElem(key, i)
}

// GetSliceElemMap returns the ith element of a slice associated
// with key as a Map (map[string]any).
// It returns nil if key does not present in Map or the value's type
// is not a slice, or i exceeds the slice's length,
// or the slice element is not a map.
func (p *SafeMap) GetSliceElemMap(key string, i int) Map {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.GetSliceElemMap(key, i)
}

// GetMap returns the value associated with the key as a Map (map[string]any).
func (p *SafeMap) GetMap(key string) Map {
	return getWithRLock(&p.mu, p.map_.GetMap, key)
}

// GetStringMap returns the value associated with the key as a map of (map[string]string).
func (p *SafeMap) GetStringMap(key string) map[string]string {
	return getWithRLock(&p.mu, p.map_.GetStringMap, key)
}

// Iterate iterates the map in unspecified order, the given function fn
// will be called for each key value pair.
// The iteration can be aborted by returning a non-zero value from fn.
func (p *SafeMap) Iterate(fn func(k string, v any) int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	p.map_.Iterate(fn)
}

// Merge merges key values from another map.
func (p *SafeMap) Merge(other map[string]any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.map_.Merge(other)
}

func getWithRLock[T any](mu *sync.RWMutex, f func(key string) T, key string) (ret T) {
	mu.RLock()
	ret = f(key)
	mu.RUnlock()
	return
}

// DecodeToStruct decodes the map to a struct using mapstructure.
// output must be a pointer to a struct.
// config is optional, if nil, the default decoder config will be used.
func (p *SafeMap) DecodeToStruct(output any, config *mapstructure.DecoderConfig) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.DecodeToStruct(output, config)
}
