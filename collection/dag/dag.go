package dag

import "slices"

// DAG is a directed acyclic graph.
// A zero value of DAG is ready to use.
type DAG[T comparable] struct {
	nodes        *dagNodes[T] // nil nodes means that the DAG is not initialized
	edges        map[T]*dagNodes[T]
	reverseEdges map[T]*dagNodes[T]
}

// NewDAG creates a new DAG object.
func NewDAG[T comparable]() *DAG[T] {
	dag := &DAG[T]{}
	dag.initialize()
	return dag
}

func (d *DAG[T]) initialize() {
	if d.nodes == nil {
		d.nodes = newDagNodes[T]()
		d.edges = make(map[T]*dagNodes[T])
		d.reverseEdges = make(map[T]*dagNodes[T])
	}
}

// AddVertex adds a vertex to the DAG.
func (d *DAG[T]) AddVertex(n T) {
	if d.nodes == nil {
		d.initialize()
	}
	d.addVertex(n)
}

// AddEdge adds an edge from 'from' to 'to' in the DAG.
// If 'from' to 'to' forms a cycle, it does not add the edge and returns true,
// otherwise it returns false.
func (d *DAG[T]) AddEdge(from, to T) (isCyclic bool) {
	if d.IsCyclic(from, to) {
		return true
	}
	if d.nodes == nil {
		d.initialize()
	}
	d.addVertex(from)
	d.addVertex(to)
	d.addToEdges(d.edges, from, to)
	d.addToEdges(d.reverseEdges, to, from)
	return false
}

func (d *DAG[T]) addVertex(n T) {
	d.nodes.Add(n)
}

func (d *DAG[T]) addToEdges(edges map[T]*dagNodes[T], from, to T) {
	nodes, ok := edges[from]
	if !ok {
		nodes = newDagNodes[T]()
		edges[from] = nodes
	}
	nodes.Add(to)
}

// IsCyclic reports whether there is a cycle from 'from' to 'to' in the DAG.
func (d *DAG[T]) IsCyclic(from, to T) bool {
	stack := []T{from}
	pop := func() (n T) {
		n = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return
	}

	seen := make(map[T]bool)
	for len(stack) > 0 {
		nodes := d.reverseEdges[pop()]
		if nodes == nil {
			continue
		}
		if nodes.Contains(to) {
			return true
		}
		for _, n := range nodes.list {
			if !seen[n] {
				stack = append(stack, n)
				seen[n] = true
			}
		}
	}
	return false
}

// VisitVertex visits all vertices in the DAG.
func (d *DAG[T]) VisitVertex(f func(n T)) {
	if d.nodes == nil {
		return
	}
	for _, n := range d.nodes.list {
		f(n)
	}
}

// VisitNeighbors visits all neighbors of 'from' in the DAG.
func (d *DAG[T]) VisitNeighbors(from T, f func(to T)) {
	nodes := d.edges[from]
	if nodes == nil {
		return
	}
	for _, n := range nodes.list {
		f(n)
	}
}

// VisitReverseNeighbors visits all reverse neighbors of 'to' in the DAG.
func (d *DAG[T]) VisitReverseNeighbors(to T, f func(from T)) {
	nodes := d.reverseEdges[to]
	if nodes == nil {
		return
	}
	for _, n := range nodes.list {
		f(n)
	}
}

// ListZeroIncomingVertices returns all vertices in the DAG that
// have no incoming edges.
func (d *DAG[T]) ListZeroIncomingVertices() []T {
	if d.nodes == nil {
		return nil
	}
	result := make([]T, 0, len(d.nodes.list))
	for _, n := range d.nodes.list {
		nodes := d.reverseEdges[n]
		if nodes == nil || len(nodes.list) == 0 {
			result = append(result, n)
		}
	}
	return result
}

// TopoSort returns a topological sort result of the DAG.
//
// The sort result is stable, which means that multiple calls
// to TopoSort will return the same result, the order is determined
// by the order of vertices and edges added to the DAG.
func (d *DAG[T]) TopoSort() []T {
	if d.nodes == nil {
		return nil
	}

	indegree := make(map[T]int)
	d.VisitVertex(func(n T) {
		d.VisitNeighbors(n, func(to T) {
			indegree[to]++
		})
	})

	// queue holds all vertices with indegree 0.
	var queue []T
	d.VisitVertex(func(n T) {
		if indegree[n] == 0 {
			queue = append(queue, n)
		}
	})
	pop := func() (n T) {
		n = queue[0]
		queue = queue[1:]
		return
	}

	order := make([]T, 0, len(d.nodes.list))
	count := 0
	for len(queue) > 0 {
		n := pop()
		order = append(order, n)
		count++
		d.VisitNeighbors(n, func(to T) {
			indegree[to]--
			if indegree[to] == 0 {
				queue = append(queue, to)
			}
		})
	}

	if count != len(d.nodes.list) { // unreachable
		panic("DAG is in invalid state")
	}

	return order
}

const fastThreshold = 64

type dagNodes[T comparable] struct {
	list []T
	set  map[T]bool
}

func newDagNodes[T comparable]() *dagNodes[T] {
	return &dagNodes[T]{}
}

// Contains reports whether n is contained in dagNodes.
func (p *dagNodes[T]) Contains(n T) bool {
	if p.set != nil {
		return p.set[n]
	}
	return slices.Contains(p.list, n)
}

// Add adds a new node to dagNodes, if it is not contained in dagNodes.
func (p *dagNodes[T]) Add(n T) {
	if p.Contains(n) {
		return
	}
	p.list = append(p.list, n)
	if len(p.list) <= fastThreshold {
		return
	}
	if p.set == nil {
		p.set = make(map[T]bool, len(p.list))
		for _, x := range p.list {
			p.set[x] = true
		}
	}
	p.set[n] = true
}
