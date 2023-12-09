package yamlx

type dag struct {
	nodes        map[int]bool
	edges        map[int]map[int]bool
	reverseEdges map[int]map[int]bool
}

func (d *dag) addVertex(n int) {
	if d.nodes == nil {
		d.nodes = make(map[int]bool)
	}
	d.nodes[n] = true
}

func (d *dag) addEdge(from, to int) (isCyclic bool) {
	if d.isCyclic(from, to) {
		return true
	}
	if d.nodes == nil {
		d.nodes = make(map[int]bool)
	}
	d.nodes[from] = true
	d.nodes[to] = true
	d.edges = d._addToEdges(d.edges, from, to)
	d.reverseEdges = d._addToEdges(d.reverseEdges, to, from)
	return false
}

func (d *dag) isCyclic(from, to int) bool {
	stack := []int{from}
	pop := func() (n int) {
		n = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return
	}

	seen := make(map[int]bool)
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

func (d *dag) _addToEdges(edges map[int]map[int]bool, from, to int) map[int]map[int]bool {
	if edges == nil {
		edges = make(map[int]map[int]bool)
	}
	m, ok := edges[from]
	if !ok {
		m = make(map[int]bool)
		edges[from] = m
	}
	m[to] = true
	return edges
}

func (d *dag) visitVertex(f func(n int)) {
	for n := range d.nodes {
		f(n)
	}
}

func (d *dag) visitNeighbors(from int, f func(to int)) {
	for n := range d.edges[from] {
		f(n)
	}
}

// Kahn's algorithm
func (d *dag) topoSort() []int {
	indegree := make(map[int]int)
	d.visitVertex(func(n int) {
		d.visitNeighbors(n, func(to int) {
			indegree[to]++
		})
	})

	// queue holds all vertices with indegree 0.
	var queue []int
	d.visitVertex(func(n int) {
		if indegree[n] == 0 {
			queue = append(queue, n)
		}
	})
	pop := func() (n int) {
		n = queue[0]
		queue = queue[1:]
		return
	}

	order := make([]int, 0, len(d.nodes))
	count := 0
	for len(queue) > 0 {
		n := pop()
		order = append(order, n)
		count++
		d.visitNeighbors(n, func(to int) {
			indegree[to]--
			if indegree[to] == 0 {
				queue = append(queue, to)
			}
		})
	}

	if count != len(d.nodes) { // unreachable
		panic("yamlx: dag is in invalid state")
	}

	return order
}
