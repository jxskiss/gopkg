package ecode

import (
	"fmt"
	"sort"
	"sync/atomic"
)

type Registry struct {
	reserved *int32
	codeSet  map[int32]struct{} // registry of all error codes
	messages atomic.Value       // map[int32]string, copy on write
}

func New() *Registry {
	return &Registry{
		codeSet: make(map[int32]struct{}),
	}
}

func NewWithReserved(maxReserved int32) *Registry {
	p := New()
	p.reserved = &maxReserved
	return p
}

func (p *Registry) Register(n int32, msg string) Code {
	if p.reserved != nil && n <= *p.reserved {
		panic(fmt.Sprintf("ecode: code <= %d is reserved", *p.reserved))
	}
	return p.add(n, msg)
}

func (p *Registry) RegisterReserved(n int32, msg string) Code {
	return p.add(n, msg)
}

func (p *Registry) add(n int32, msg string) Code {
	if _, ok := p.codeSet[n]; ok {
		panic(fmt.Sprintf("ecode: %d already registered", n))
	}
	p.codeSet[n] = struct{}{}
	if len(msg) > 0 {
		p.AddMessages(map[int32]string{n: msg})
	}
	return Code{code: n, reg: p}
}

func (p *Registry) SetMessages(messages map[int32]string) {
	p.messages.Store(messages)
}

func (p *Registry) AddMessages(messages map[int32]string) {
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

func (p *Registry) Dump() []Code {
	out := make([]Code, 0, len(p.codeSet))
	for code := range p.codeSet {
		out = append(out, Code{code: code, reg: p})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].code < out[j].code
	})
	return out
}

func (p *Registry) getMessage(code int32) string {
	msgs, _ := p.messages.Load().(map[int32]string)
	return msgs[code]
}
