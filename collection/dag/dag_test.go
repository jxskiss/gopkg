package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDAG(t *testing.T) {
	d := NewDAG[int]()
	d.AddVertex(1)
	d.AddVertex(2)
	d.AddEdge(1, 2)
	assert.False(t, d.IsCyclic(1, 2))
	assert.True(t, d.AddEdge(2, 1))

	topoOrder := d.TopoSort()
	assert.Equal(t, []int{1, 2}, topoOrder)
}
