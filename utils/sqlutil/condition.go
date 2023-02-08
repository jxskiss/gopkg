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
func Cond(clause string, args ...any) *Condition {
	return new(Condition).And(clause, args...)
}

// Condition represents a query filter to work with SQL query.
type Condition struct {
	builder strings.Builder
	prefix  []byte
	args    []any
}

// And combines the given query filter to Condition using "AND" operator.
func (p *Condition) And(clause string, args ...any) *Condition {
	if clause == "" {
		return p
	}

	// encapsulate with brackets to avoid misuse
	clause = strings.TrimSpace(clause)
	if containsOr(clause) {
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
func (p *Condition) Or(clause string, args ...any) *Condition {
	if clause == "" {
		return p
	}

	// encapsulate with brackets to avoid misuse
	clause = strings.TrimSpace(clause)
	if containsAnd(clause) {
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
func (p *Condition) IfAnd(cond bool, clause string, args ...any) *Condition {
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
func (p *Condition) IfOr(cond bool, clause string, args ...any) *Condition {
	if cond {
		return p.Or(clause, args...)
	}
	return p
}

// IfOrCond checks cond, if cond is true, it combines the given Condition
// using "OR" operator.
func (p *Condition) IfOrCond(cond bool, c *Condition) *Condition {
	if cond {
		return p.OrCond(c)
	}
	return p
}

// Build returns the query filter clause and parameters of the Condition.
func (p *Condition) Build() (string, []any) {
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
	parenCnt := 0
	clause = strings.ToLower(clause)
	for i := 0; i < len(clause)-4; i++ {
		switch clause[i] {
		case '(':
			parenCnt++
		case ')':
			parenCnt--
		case 'a':
			if clause[i:i+3] == "and" &&
				i > 0 && isWhitespace(clause[i-1]) && isWhitespace(clause[i+3]) &&
				parenCnt == 0 {
				return true
			}
		}
	}
	return false
}

func containsOr(clause string) bool {
	parenCnt := 0
	clause = strings.ToLower(clause)
	for i := 0; i < len(clause)-3; i++ {
		switch clause[i] {
		case '(':
			parenCnt++
		case ')':
			parenCnt--
		case 'o':
			if clause[i:i+2] == "or" &&
				i > 0 && isWhitespace(clause[i-1]) && isWhitespace(clause[i+2]) &&
				parenCnt == 0 {
				return true
			}
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
