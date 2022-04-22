package validator

import (
	"strings"
)

type treeLog struct {
	unique map[validator]bool
	tree   map[validator][]validator // map[parent][]child
}

func newTreeLog(t Tree) treeLog {
	d := treeLog{
		unique: make(map[validator]bool),
		tree:   make(map[validator][]validator),
	}
	for _, leaf := range t.leaves {
		d.addChild(leaf)
	}
	return d
}

func (d treeLog) addChild(v validator) {
	if _, ok := d.unique[v]; !ok {
		parent := v.parent()

		_, ok := d.tree[parent]
		if !ok {
			d.tree[parent] = make([]validator, 0, 1)
		}
		d.tree[parent] = append(d.tree[parent], v)

		d.unique[v] = true

		if parent != nil {
			d.addChild(parent)
		}
	}
}

func (d treeLog) string() string {
	return d.nodeToString(nil, 0)
}

func (d treeLog) nodeToString(v validator, depth int) string {
	var str strings.Builder
	indent := strings.Repeat("\t", depth)
	for _, child := range d.tree[v] { // children
		str.WriteString(indent)
		str.WriteString(child.log())
		str.WriteString("\n")
		str.WriteString(d.nodeToString(child, depth+1))
	}
	return str.String()
}
