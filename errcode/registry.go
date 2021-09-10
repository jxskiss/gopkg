package errcode

import (
	"fmt"
	"sort"
	"sync/atomic"
)

// Registry represents an error code registry.
type Registry struct {
	reserve  func(code int32) bool // check reserved code
	codes    map[int32]struct{}    // registry of all error codes
	messages atomic.Value          // map[int32]string, copy on write
}

// New creates a new error code registry.
func New() *Registry {
	r := &Registry{
		codes: make(map[int32]struct{}),
	}
	r.messages.Store(make(map[int32]string))
	return r
}

// NewWithReserved creates a new error code registry with reserved codes,
// calling Register with a reserved code causes a panic.
// Reserved code can be registered by calling RegisterReserved.
func NewWithReserved(reserveFunc func(code int32) bool) *Registry {
	p := New()
	p.reserve = reserveFunc
	return p
}

// Register register an error code to the registry.
// If the registry is created by NewWithReserved, it checks the code with
// the reserve function and panics if the code is reserved.
func (p *Registry) Register(code int32, msg string) *Code {
	if p.reserve != nil && p.reserve(code) {
		panic(fmt.Sprintf("errcode: code %d is reserved", code))
	}
	return p.add(code, msg)
}

// RegisterReserved register an error code to the registry.
// It does not checks the reserve function, but simply adds the code
// to the register.
func (p *Registry) RegisterReserved(code int32, msg string) *Code {
	return p.add(code, msg)
}

func (p *Registry) add(code int32, msg string) *Code {
	if _, ok := p.codes[code]; ok {
		panic(fmt.Sprintf("errcode: code %d is already registered", code))
	}
	p.codes[code] = struct{}{}
	if msg != "" {
		messages := p.messages.Load().(map[int32]string)
		messages[code] = msg
	}
	return &Code{code: code, reg: p}
}

// UpdateMessages updates error messages to the registry.
// This method copies the underlying message map, it's safe for
// concurrent use.
func (p *Registry) UpdateMessages(messages map[int32]string) {
	oldMsgs, _ := p.messages.Load().(map[int32]string)
	newMsgs := make(map[int32]string, len(oldMsgs))
	for code, msg := range oldMsgs {
		newMsgs[code] = msg
	}
	for code, msg := range messages {
		newMsgs[code] = msg
	}
	p.messages.Store(newMsgs)
}

// Dump returns all error codes registered with the registry.
func (p *Registry) Dump() []*Code {
	out := make([]*Code, 0, len(p.codes))
	for code := range p.codes {
		out = append(out, &Code{code: code, reg: p})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].code < out[j].code
	})
	return out
}

func (p *Registry) getMessage(code int32) string {
	if p == nil {
		return ""
	}
	msgs, _ := p.messages.Load().(map[int32]string)
	return msgs[code]
}
