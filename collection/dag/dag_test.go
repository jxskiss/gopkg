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

func TestDAG_VisitNeighbors(t *testing.T) {
	d := NewDAG[int]()
	d.AddEdge(1, 2)
	d.AddEdge(1, 3)
	d.AddEdge(2, 4)
	d.AddEdge(3, 4)

	var got1 [][]int
	for i := 0; i < 100; i++ {
		var got []int
		d.VisitNeighbors(1, func(to int) {
			got = append(got, to)
		})
		got1 = append(got1, got)
	}
	for i := 0; i < len(got1); i++ {
		assert.Equal(t, []int{2, 3}, got1[i])
	}

	var got2 [][]int
	for i := 0; i < 100; i++ {
		var got []int
		d.VisitReverseNeighbors(4, func(from int) {
			got = append(got, from)
		})
		got2 = append(got2, got)
	}
	for i := 0; i < len(got2); i++ {
		assert.Equal(t, []int{2, 3}, got2[i])
	}
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

func TestDAG_uninitialized(t *testing.T) {
	var d DAG[int]
	topoOrder := d.TopoSort()
	assert.Equal(t, 0, len(topoOrder))
	assert.NotPanics(t, func() {
		d.VisitVertex(func(n int) {
			t.Logf("visit vertex: %d", n)
		})
		d.VisitNeighbors(1, func(to int) {
			t.Logf("visit neighbor: %d -> %d", 1, to)
		})
		assert.Equal(t, 0, len(d.TopoSort()))
		d.AddEdge(1, 2)
		assert.Equal(t, []int{1, 2}, d.TopoSort())
	})
}
