package sqlutil

import (
	"fmt"
	"strings"
	"unsafe"
)

// And returns a new *Condition which is combination of given conditions
// using the "AND" operator.
func And(conds ...*Condition) *Condition {
	f := new(Condition)
	for _, c := range conds {
		clause, args := c.Build()
		f.And(clause, args...)
	}
	return f
}

// Or returns a new *Condition which is combination of given conditions
// using the "OR" operator.
func Or(conds ...*Condition) *Condition {
	f := new(Condition)
	for _, c := range conds {
		clause, args := c.Build()
		f.Or(clause, args...)
	}
	return f
}

// Cond creates a new *Condition from the given params.
func Cond(clause string, args ...interface{}) *Condition {
	return new(Condition).And(clause, args...)
}

// Condition represents a query filter to work with SQL query.
type Condition struct {
	builder strings.Builder
	prefix  []byte
	args    []interface{}
}

// And combines the given query filter to Condition using "AND" operator.
func (p *Condition) And(clause string, args ...interface{}) *Condition {
	if clause == "" {
		return p
	}

	// encapsulate with brackets to avoid misuse
	clause = strings.TrimSpace(clause)
	if containsOr(clause) &&
		!(clause[0] == '(' && clause[len(clause)-1] == ')') {
		clause = "(" + clause + ")"
	}

	if p.builder.Len() == 0 {
		p.builder.WriteString(clause)
	} else {
		p.builder.WriteString(" AND ")
		p.builder.WriteString(clause)
	}
	p.args = append(p.args, args...)
	return p
}

// AndCond combines the given Condition using "AND" operator.
func (p *Condition) AndCond(c *Condition) *Condition {
	clause, args := c.Build()
	return p.And(clause, args...)
}

// Or combines the given query filter to Condition using "OR" operator.
func (p *Condition) Or(clause string, args ...interface{}) *Condition {
	if clause == "" {
		return p
	}

	// encapsulate with brackets to avoid misuse
	clause = strings.TrimSpace(clause)
	if containsAnd(clause) &&
		!(clause[0] == '(' && clause[len(clause)-1] == ')') {
		clause = "(" + clause + ")"
	}

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

// OrCond combines the given Condition using "OR" operator.
func (p *Condition) OrCond(c *Condition) *Condition {
	clause, args := c.Build()
	return p.Or(clause, args...)
}

// IfAnd checks cond, if cond is true, it combines the query filter
// to Condition using "AND" operator.
func (p *Condition) IfAnd(cond bool, clause string, args ...interface{}) *Condition {
	if cond {
		return p.And(clause, args...)
	}
	return p
}

// IfAndCond checks cond, if cond is true, it combines the given Condition
// using "AND" operator.
func (p *Condition) IfAndCond(cond bool, c *Condition) *Condition {
	if cond {
		return p.AndCond(c)
	}
	return p
}

// IfOr checks cond, it cond is true, it combines the query filter
// to Condition using "OR" operator.
func (p *Condition) IfOr(cond bool, clause string, args ...interface{}) *Condition {
	if cond {
		return p.Or(clause, args...)
	}
	return p
}

func (p *Condition) IfOrCond(cond bool, c *Condition) *Condition {
	if cond {
		return p.OrCond(c)
	}
	return p
}

// Build returns the query filter clause and parameters of the Condition.
func (p *Condition) Build() (string, []interface{}) {
	buf := make([]byte, len(p.prefix)+p.builder.Len())
	copy(buf, p.prefix)
	copy(buf[len(p.prefix):], p.builder.String())
	clause := *(*string)(unsafe.Pointer(&buf))
	return clause, p.args
}

// String returns the string representation of the Condition.
func (p *Condition) String() string {
	clause, args := p.Build()
	format := strings.Replace(clause, "?", "%v", -1)
	return fmt.Sprintf(format, args...)
}

func containsAnd(clause string) bool {
	lower := strings.ToLower(clause)
	idx := strings.Index(lower, "and")
	if idx > 0 && len(clause) > idx+3 {
		if isWhitespace(clause[idx-1]) && isWhitespace(clause[idx+3]) {
			return true
		}
	}
	return false
}

func containsOr(clause string) bool {
	lower := strings.ToLower(clause)
	idx := strings.Index(lower, "or")
	if idx > 0 && len(clause) > idx+3 {
		if isWhitespace(clause[idx-1]) && isWhitespace(clause[idx+2]) {
			return true
		}
	}
	return false
}

func isWhitespace(b byte) bool {
	switch b {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}
