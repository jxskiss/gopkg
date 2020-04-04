package sqlutil

import (
	"strings"
	"unsafe"
)

func And(conds ...*Builder) *Builder {
	f := new(Builder)
	for _, c := range conds {
		clause, args := c.Build()
		f.And(clause, args...)
	}
	return f
}

func Or(conds ...*Builder) *Builder {
	f := new(Builder)
	for _, c := range conds {
		clause, args := c.Build()
		f.Or(clause, args...)
	}
	return f
}

func Cond(clause string, args ...interface{}) *Builder {
	return new(Builder).And(clause, args...)
}

type Builder struct {
	builder strings.Builder
	prefix  []byte
	args    []interface{}
}

func (p *Builder) And(clause string, args ...interface{}) *Builder {
	if p.builder.Len() == 0 {
		p.builder.WriteString(clause)
	} else {
		p.builder.WriteString(" AND ")
		p.builder.WriteString(clause)
	}
	p.args = append(p.args, args...)
	return p
}

func (p *Builder) Or(clause string, args ...interface{}) *Builder {
	if p.builder.Len() == 0 {
		p.builder.WriteString(clause)
	} else {
		p.prefix = append(p.prefix, '(')
		p.builder.WriteString(" OR ")
		p.builder.WriteString(clause)
		p.builder.WriteByte(')')
	}
	p.args = append(p.args, args...)
	return p
}

//func (p *Builder) shouldAddBrackets(clause string) bool {
//	return (clause[0] != '(' || clause[len(clause)-1] != ')') &&
//		strings.Contains(strings.ToLower(clause), " or ")
//}

func (p *Builder) IfAnd(cond bool, clause string, args ...interface{}) *Builder {
	if cond {
		return p.And(clause, args...)
	}
	return p
}

func (p *Builder) IfOr(cond bool, clause string, args ...interface{}) *Builder {
	if cond {
		return p.Or(clause, args...)
	}
	return p
}

func (p *Builder) Build() (string, []interface{}) {
	buf := make([]byte, len(p.prefix)+p.builder.Len())
	copy(buf, p.prefix)
	copy(buf[len(p.prefix):], p.builder.String())
	clause := *(*string)(unsafe.Pointer(&buf))
	return clause, p.args
}

func (p *Builder) String() string {
	clause, _ := p.Build()
	return clause
}
