package ecode

import (
	"fmt"
	"sort"
	"sync/atomic"
)

type Registry struct {
	reserve  func(code int32) bool // check reserved code
	codeSet  map[int32]struct{}    // registry of all error codes
	messages atomic.Value          // map[int32]string, copy on write
}

func New() *Registry {
	r := &Registry{
		codeSet: make(map[int32]struct{}),
	}
	r.messages.Store(make(map[int32]string))
	return r
}

func NewWithReserved(reserveFunc func(code int32) bool) *Registry {
	p := New()
	p.reserve = reserveFunc
	return p
}

func (p *Registry) Register(code int32, msg string) *Code {
	if p.reserve != nil && p.reserve(code) {
		panic(fmt.Sprintf("ecode: code %d is reserved", code))
	}
	return p.add(code, msg)
}

func (p *Registry) RegisterReserved(code int32, msg string) *Code {
	return p.add(code, msg)
}

func (p *Registry) add(code int32, msg string) *Code {
	if _, ok := p.codeSet[code]; ok {
		panic(fmt.Sprintf("ecode: code %d is already registered", code))
	}
	p.codeSet[code] = struct{}{}
	if msg != "" {
		messages := p.messages.Load().(map[int32]string)
		messages[code] = msg
	}
	return &Code{code: code, reg: p}
}

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

func (p *Registry) Dump() []*Code {
	out := make([]*Code, 0, len(p.codeSet))
	for code := range p.codeSet {
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
