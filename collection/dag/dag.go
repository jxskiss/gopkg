package dag

// DAG is a directed acyclic graph.
// A zero value of DAG is ready to use.
type DAG[T comparable] struct {
	initialized  bool
	nodes        map[T]bool
	edges        map[T]map[T]bool
	reverseEdges map[T]map[T]bool
}

// NewDAG creates a new DAG object.
func NewDAG[T comparable]() *DAG[T] {
	return &DAG[T]{}
}

func (d *DAG[T]) initialize() {
	if d.initialized {
		return
	}
	d.initialized = true
	d.nodes = make(map[T]bool)
	d.edges = make(map[T]map[T]bool)
	d.reverseEdges = make(map[T]map[T]bool)
}

// AddVertex adds a vertex to the DAG.
func (d *DAG[T]) AddVertex(n T) {
	if !d.initialized {
		d.initialize()
	}
	d.nodes[n] = true
}

// AddEdge adds an edge from 'from' to 'to' in the DAG.
// If 'from' to 'to' forms a cycle, it does not add the edge and returns true,
// otherwise it returns false.
func (d *DAG[T]) AddEdge(from, to T) (isCyclic bool) {
	if d.IsCyclic(from, to) {
		return true
	}
	if !d.initialized {
		d.initialize()
	}
	d.nodes[from] = true
	d.nodes[to] = true
	d.addToEdges(d.edges, from, to)
	d.addToEdges(d.reverseEdges, to, from)
	return false
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
		m := d.reverseEdges[pop()]
		if m[to] {
			return true
		}
		for n := range m {
			if !seen[n] {
				stack = append(stack, n)
				seen[n] = true
			}
		}
	}
	return false
}

func (d *DAG[T]) addToEdges(edges map[T]map[T]bool, from, to T) {
	m, ok := edges[from]
	if !ok {
		m = make(map[T]bool)
		edges[from] = m
	}
	m[to] = true
}

// VisitVertex visits all vertices in the DAG.
func (d *DAG[T]) VisitVertex(f func(n T)) {
	for n := range d.nodes {
		f(n)
	}
}

// VisitNeighbors visits all neighbors of 'from' in the DAG.
func (d *DAG[T]) VisitNeighbors(from T, f func(to T)) {
	for n := range d.edges[from] {
		f(n)
	}
}

// TopoSort returns a topological sort of the DAG using Kahn's algorithm.
func (d *DAG[T]) TopoSort() []T {
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

	order := make([]T, 0, len(d.nodes))
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

	if count != len(d.nodes) { // unreachable
		panic("DAG is in invalid state")
	}

	return order
}
