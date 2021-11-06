package zlog

import (
	"fmt"
	"sort"
	"strings"
)

type perLoggerLevelFunc func(name string) (Level, bool)

func buildPerLoggerLevelFunc(levelRules []string) (*levelTree, perLoggerLevelFunc, error) {
	if len(levelRules) == 0 {
		return nil, nil, nil
	}
	tree := &levelTree{}
	for _, rule := range levelRules {
		tmp := strings.Split(rule, "=")
		if len(tmp) != 2 {
			return nil, nil, fmt.Errorf("invalid per logger level rule: %s", rule)
		}
		name, levelName := tmp[0], tmp[1]
		var level Level
		if !level.unmarshalText([]byte(levelName)) {
			return nil, nil, fmt.Errorf("unrecognized level: %s", levelName)
		}
		tree.root.insert(name, level)
	}
	return tree, tree.search, nil
}

type levelTree struct {
	root radixNode
}

func (p *levelTree) search(name string) (Level, bool) {
	level, found := p.root.search(name)
	return level, found
}

type radixNode struct {
	prefix     string
	level      *Level
	childNodes childNodes
}

type childNodes struct {
	labels []byte
	nodes  []*radixNode
}

func (nodes *childNodes) Len() int { return len(nodes.labels) }

func (nodes *childNodes) Less(i, j int) bool { return nodes.labels[i] < nodes.labels[j] }

func (nodes *childNodes) Swap(i, j int) {
	nodes.labels[i], nodes.labels[j] = nodes.labels[j], nodes.labels[i]
	nodes.nodes[i], nodes.nodes[j] = nodes.nodes[j], nodes.nodes[i]
}

func (n *radixNode) getLevel() (Level, bool) {
	if n.level == nil {
		return 0, false
	}
	return *n.level, true
}

func (n *radixNode) insert(name string, level Level) {
	if name == "" {
		n.level = &level
		return
	}

	firstChar := name[0]
	for i, label := range n.childNodes.labels {
		if firstChar == label {
			// Split based on the common prefix of the existing node and the new one.
			child, prefixSplit := n.splitCommonPrefix(i, name)
			child.insert(name[prefixSplit:], level)
			return
		}
	}

	// No existing node starting with this letter, so create it.
	child := &radixNode{prefix: name, level: &level}
	n.childNodes.labels = append(n.childNodes.labels, firstChar)
	n.childNodes.nodes = append(n.childNodes.nodes, child)
	sort.Sort(&n.childNodes)
	return
}

func (n *radixNode) splitCommonPrefix(existingChildIndex int, name string) (*radixNode, int) {
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
	newNode := &radixNode{
		prefix: commonPrefix,
		childNodes: childNodes{
			labels: []byte{child.prefix[0]},
			nodes:  []*radixNode{child},
		},
	}
	n.childNodes.nodes[existingChildIndex] = newNode
	return newNode, i
}

func (n *radixNode) search(name string) (level Level, found bool) {
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
	return node.getLevel()
}

func (n *radixNode) getChild(label byte) *radixNode {
	num := len(n.childNodes.labels)
	i := sort.Search(num, func(i int) bool {
		return n.childNodes.labels[i] >= label
	})
	if i < num && n.childNodes.labels[i] == label {
		return n.childNodes.nodes[i]
	}
	return nil
}

func (n *radixNode) dumpTree(prefix string) string {
	var out string
	if n.level != nil {
		out += fmt.Sprintf("%s=%s\n", prefix+n.prefix, n.level.String())
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
