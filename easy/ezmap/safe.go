package ezmap

import (
	"sync"
	"time"

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

func (p *SafeMap) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.MarshalJSON()
}

func (p *SafeMap) MarshalJSONPretty() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.MarshalJSONPretty()
}

func (p *SafeMap) UnmarshalJSON(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.map_.UnmarshalJSON(data)
}

func (p *SafeMap) MarshalYAML() (any, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.MarshalYAML()
}

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

func (p *SafeMap) Set(key string, value any) {
	p.mu.Lock()
	p.map_.Set(key, value)
	p.mu.Unlock()
}

func (p *SafeMap) Get(key string) (value any, exists bool) {
	p.mu.RLock()
	value, exists = p.map_.Get(key)
	p.mu.RUnlock()
	return
}

func (p *SafeMap) GetOr(key string, defaultVal any) (value any) {
	p.mu.RLock()
	value = p.map_.GetOr(key, defaultVal)
	p.mu.RUnlock()
	return
}

func (p *SafeMap) MustGet(key string) any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.map_.MustGet(key)
}

func (p *SafeMap) GetString(key string) string {
	return getWithRLock(&p.mu, p.map_.GetString, key)
}

func (p *SafeMap) GetBytes(key string) []byte {
	return getWithRLock(&p.mu, p.map_.GetBytes, key)
}

func (p *SafeMap) GetBool(key string) bool {
	return getWithRLock(&p.mu, p.map_.GetBool, key)
}

func (p *SafeMap) GetInt(key string) int {
	return getWithRLock(&p.mu, p.map_.GetInt, key)
}

func (p *SafeMap) GetInt32(key string) int32 {
	return getWithRLock(&p.mu, p.map_.GetInt32, key)
}

func (p *SafeMap) GetInt64(key string) int64 {
	return getWithRLock(&p.mu, p.map_.GetInt64, key)
}

func (p *SafeMap) GetUint(key string) uint {
	return getWithRLock(&p.mu, p.map_.GetUint, key)
}

func (p *SafeMap) GetUint32(key string) uint32 {
	return getWithRLock(&p.mu, p.map_.GetUint32, key)
}

func (p *SafeMap) GetUint64(key string) uint64 {
	return getWithRLock(&p.mu, p.map_.GetUint64, key)
}

func (p *SafeMap) GetFloat(key string) float64 {
	return getWithRLock(&p.mu, p.map_.GetFloat, key)
}

func (p *SafeMap) GetTime(key string) time.Time {
	return getWithRLock(&p.mu, p.map_.GetTime, key)
}

func (p *SafeMap) GetDuration(key string) time.Duration {
	return getWithRLock(&p.mu, p.map_.GetDuration, key)
}

func (p *SafeMap) GetInt64s(key string) []int64 {
	return getWithRLock(&p.mu, p.map_.GetInt64s, key)
}

func (p *SafeMap) GetInt32s(key string) []int32 {
	return getWithRLock(&p.mu, p.map_.GetInt32s, key)
}

func (p *SafeMap) GetStrings(key string) []string {
	return getWithRLock(&p.mu, p.map_.GetStrings, key)
}

func (p *SafeMap) GetSlice(key string) any {
	return getWithRLock(&p.mu, p.map_.GetSlice, key)
}

func (p *SafeMap) GetMap(key string) Map {
	return getWithRLock(&p.mu, p.map_.GetMap, key)
}

func (p *SafeMap) GetStringMap(key string) map[string]string {
	return getWithRLock(&p.mu, p.map_.GetStringMap, key)
}

func (p *SafeMap) Iterate(fn func(k string, v any) int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	p.map_.Iterate(fn)
}

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
