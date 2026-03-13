package workflow

import "sync"

// RunContext is a concurrent-safe context shared by all tasks
// in a single workflow running.
type RunContext struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewRunContext creates an empty RunContext.
func NewRunContext() *RunContext {
	return &RunContext{
		data: make(map[string]any),
	}
}

// Store stores a value for key.
func (rc *RunContext) Store(key string, value any) {
	rc.mu.Lock()
	rc.data[key] = value
	rc.mu.Unlock()
}

// Load returns value and existence for key.
func (rc *RunContext) Load(key string) (value any, ok bool) {
	rc.mu.RLock()
	value, ok = rc.data[key]
	rc.mu.RUnlock()
	return value, ok
}

// Delete removes key from context.
func (rc *RunContext) Delete(key string) {
	rc.mu.Lock()
	delete(rc.data, key)
	rc.mu.Unlock()
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
func (rc *RunContext) LoadOrStore(key string, value any) (actual any, loaded bool) {
	rc.mu.Lock()
	actual, loaded = rc.data[key]
	if !loaded {
		rc.data[key] = value
		actual = value
	}
	rc.mu.Unlock()
	return actual, loaded
}
