package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDAG(t *testing.T) {
	d := New[int]()
	d.AddVertex(1)
	d.AddVertex(2)
	d.AddEdge(1, 2)
	assert.False(t, d.IsCyclic(1, 2))
	assert.True(t, d.AddEdge(2, 1))

	topoOrder := d.TopoSort()
	assert.Equal(t, []int{1, 2}, topoOrder)
}

func TestDAG_SelfLoop(t *testing.T) {
	d := New[int]()
	assert.True(t, d.IsCyclic(1, 1))
	assert.True(t, d.AddEdge(1, 1))
	assert.False(t, d.HasEdge(1, 1))
}

func TestDAG_Remove(t *testing.T) {
	d := New[int]()
	d.AddEdge(1, 2)
	d.AddEdge(2, 3)

	assert.True(t, d.HasEdge(1, 2))
	d.RemoveEdge(1, 2)
	assert.False(t, d.HasEdge(1, 2))

	d.AddEdge(1, 2)
	d.RemoveVertex(2)
	assert.False(t, d.HasEdge(1, 2))
	assert.False(t, d.HasEdge(2, 3))
	assert.False(t, d.nodes.Contains(2))

	// 1 and 3 remain. Both have 0 incoming edges.
	zeroIncoming := d.ListZeroIncomingVertices()
	assert.Contains(t, zeroIncoming, 1)
	assert.Contains(t, zeroIncoming, 3)
	assert.Len(t, zeroIncoming, 2)
}

func TestDAG_LargeData(t *testing.T) {
	d := New[int]()
	// Add > 64 neighbors to trigger set creation in dagNodes
	for i := 0; i < 100; i++ {
		d.AddEdge(0, i+1)
	}
	neighbors := d.GetNeighbors(0)
	assert.Equal(t, 100, len(neighbors))
	for i := 0; i < 100; i++ {
		assert.True(t, d.HasEdge(0, i+1))
	}

	// Test remove from large set
	d.RemoveEdge(0, 50)
	assert.False(t, d.HasEdge(0, 50))
	assert.Equal(t, 99, len(d.GetNeighbors(0)))

	// Add back to ensure set is maintained correctly
	d.AddEdge(0, 50)
	assert.True(t, d.HasEdge(0, 50))
	assert.Equal(t, 100, len(d.GetNeighbors(0)))
}

func TestDAG_Helpers(t *testing.T) {
	d := New[int]()
	d.AddEdge(1, 2)

	assert.Equal(t, []int{2}, d.GetNeighbors(1))
	assert.Equal(t, []int{1}, d.GetReverseNeighbors(2))
	assert.Empty(t, d.GetNeighbors(2))
	assert.Empty(t, d.GetReverseNeighbors(1))
}

func TestDAG_VisitNeighbors(t *testing.T) {
	d := New[int]()
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

func TestDAG_ListZeroIncomingVertices(t *testing.T) {
	d := New[int]()
	d.AddEdge(1, 2)
	d.AddEdge(1, 3)
	d.AddEdge(2, 4)
	d.AddEdge(3, 4)
	d.AddEdge(5, 2)
	d.AddEdge(6, 4)
	assert.Equal(t, []int{1, 5, 6}, d.ListZeroIncomingVertices())
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
	d := New[int]()
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
