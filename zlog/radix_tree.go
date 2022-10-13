package zlog

import (
	"fmt"
	"sort"
	"strings"
)

type radixTree[T any] struct {
	root radixNode[T]
}

func (p *radixTree[T]) search(name string) (T, bool) {
	value, found := p.root.search(name)
	return value, found
}

type radixNode[T any] struct {
	prefix     string
	value      *T
	childNodes childNodes[T]
}

type childNodes[T any] struct {
	labels []byte
	nodes  []*radixNode[T]
}

func (nodes *childNodes[_]) Len() int { return len(nodes.labels) }

func (nodes *childNodes[_]) Less(i, j int) bool { return nodes.labels[i] < nodes.labels[j] }

func (nodes *childNodes[_]) Swap(i, j int) {
	nodes.labels[i], nodes.labels[j] = nodes.labels[j], nodes.labels[i]
	nodes.nodes[i], nodes.nodes[j] = nodes.nodes[j], nodes.nodes[i]
}

func (n *radixNode[T]) getValue() (T, bool) {
	if n.value == nil {
		var zero T
		return zero, false
	}
	return *n.value, true
}

func (n *radixNode[T]) insert(name string, value T) {
	if name == "" {
		n.value = &value
		return
	}

	firstChar := name[0]
	for i, label := range n.childNodes.labels {
		if firstChar == label {
			// Split based on the common prefix of the existing node and the new one.
			child, prefixSplit := n.splitCommonPrefix(i, name)
			child.insert(name[prefixSplit:], value)
			return
		}
	}

	// No existing node starting with this letter, so create it.
	child := &radixNode[T]{prefix: name, value: &value}
	n.childNodes.labels = append(n.childNodes.labels, firstChar)
	n.childNodes.nodes = append(n.childNodes.nodes, child)
	sort.Sort(&n.childNodes)
	return
}

func (n *radixNode[T]) splitCommonPrefix(existingChildIndex int, name string) (*radixNode[T], int) {
	child := n.childNodes.nodes[existingChildIndex]

	if strings.HasPrefix(name, child.prefix) {
		// No split needs to be done. Rather, the new name shares the entire
		// prefix with the existing node, so the new node is just a child of
		// the existing one. Or the new name is the same as the existing name,
		// which means that we just move on to the next token.
		// Either way, this return accomplishes that.
		return child, len(child.prefix)
	}

	i := longestPrefix(name, child.prefix)
	commonPrefix := name[:i]
	child.prefix = child.prefix[i:]

	// Create a new intermediary node in the place of the existing node, with
	// the existing node as a child.
	newNode := &radixNode[T]{
		prefix: commonPrefix,
		childNodes: childNodes[T]{
			labels: []byte{child.prefix[0]},
			nodes:  []*radixNode[T]{child},
		},
	}
	n.childNodes.nodes[existingChildIndex] = newNode
	return newNode, i
}

func (n *radixNode[T]) search(name string) (value T, found bool) {
	node := n
	for {
		nameLen := len(name)
		if nameLen == 0 {
			break
		}
		firstChar := name[0]
		if child := node.getChild(firstChar); child != nil {
			childPrefixLen := len(child.prefix)
			if nameLen >= childPrefixLen && child.prefix == name[:childPrefixLen] {
				name = name[len(child.prefix):]
				node = child
				continue
			}
		}
		break
	}
	return node.getValue()
}

func (n *radixNode[T]) getChild(label byte) *radixNode[T] {
	num := len(n.childNodes.labels)
	i := sort.Search(num, func(i int) bool {
		return n.childNodes.labels[i] >= label
	})
	if i < num && n.childNodes.labels[i] == label {
		return n.childNodes.nodes[i]
	}
	return nil
}

func (n *radixNode[T]) dumpTree(prefix string) string {
	var out string
	if n.value != nil {
		out += fmt.Sprintf("%s=%v\n", prefix+n.prefix, n.value)
	}
	for _, child := range n.childNodes.nodes {
		out += child.dumpTree(prefix + n.prefix)
	}
	return out
}

// longestPrefix finds the length of the shared prefix of two strings.
func longestPrefix(k1, k2 string) int {
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}
	var i int
	for i = 0; i < max; i++ {
		if k1[i] != k2[i] {
			break
		}
	}
	return i
}
