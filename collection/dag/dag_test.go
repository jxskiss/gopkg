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

// https://en.wikipedia.org/wiki/Topological_sorting
func TestDAG_TopoSort(t *testing.T) {
	/*
		5 -> 11
		7 -> 11
		7 -> 8
		3 -> 8
		3 -> 10
		11 -> 2
		11 -> 9
		11 -> 10
		8 -> 9
		2, 9, 10
	*/
	d := NewDAG[int]()
	d.AddEdge(5, 11)
	d.AddEdge(7, 11)
	d.AddEdge(7, 8)
	d.AddEdge(3, 8)
	d.AddEdge(3, 10)
	d.AddEdge(11, 2)
	d.AddEdge(11, 9)
	d.AddEdge(11, 10)
	d.AddEdge(8, 9)
	d.AddVertex(2)
	d.AddVertex(9)
	d.AddVertex(10)
	t.Logf("topo sort result: %v", d.TopoSort())

	var got [][]int
	for i := 0; i < 100; i++ {
		topoOrder := d.TopoSort()
		got = append(got, topoOrder)
	}
	assert.Equal(t, 100, len(got))
	for i := 1; i < len(got); i++ {
		assert.Equal(t, got[0], got[i])
	}
}
