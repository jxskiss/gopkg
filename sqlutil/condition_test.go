package sqlutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilter(t *testing.T) {
	want1 := "a = ? AND b = ? AND (c = ? AND d = 4 OR (e = ? OR f = 6))"
	builder1 := And(
		Cond("a = ?", 1),
		Cond("b = ?", 2),
		Or(
			Cond("c = ?", 3).IfAnd(true, "d = 4"),
			Cond("e = ?", 5).IfOr(true, "f = 6"),
		),
	)
	clause1, args1 := builder1.Build()
	assert.Equal(t, want1, clause1)
	assert.Equal(t, want1, builder1.String())
	assert.Equal(t, []interface{}{1, 2, 3, 5}, args1)

	want2 := "a = ? AND b = ? AND (c = ? OR d = 4) AND (e = ? OR f = 6)"
	builder2 := And(
		Cond("a = ?", 1),
		Cond("b = ?", 2),
		And(
			Cond("c = ?", 3).Or("d = 4"),
			Cond("e = ?", 5).Or("f = 6"),
		),
	)
	clause2, args2 := builder2.Build()
	assert.Equal(t, want2, clause2)
	assert.Equal(t, want2, builder2.String())
	assert.Equal(t, []interface{}{1, 2, 3, 5}, args2)
}
