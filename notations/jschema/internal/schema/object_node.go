package schema

import (
	"context"
	"fmt"
	jschema "j/schema"
	"j/schema/internal/errors"
	"j/schema/internal/json"
	"j/schema/internal/lexeme"
	"strings"
	"sync"
)

type ObjectNode struct {
	baseNode

	// children node list.
	children []Node

	// keys stores the index of the node on the map for quick search.
	keys *objectNodeKeys

	// waitingForChild indicates that the Grow method will create a child node by
	// getting the next lexeme.
	waitingForChild bool
}

// gen:OrderedMap
type objectNodeKeys struct {
	data  map[string]objectNodeKey
	order []string
	mx    sync.RWMutex
}

type objectNodeKey struct {
	Index      int
	IsShortcut bool
	Lex        lexeme.LexEvent
}

var _ Node = &ObjectNode{}

func newObjectNode(lex lexeme.LexEvent) *ObjectNode {
	n := ObjectNode{
		baseNode: newBaseNode(lex),
		children: make([]Node, 0, 10),
		keys:     &objectNodeKeys{},
	}
	n.setJsonType(json.TypeObject)
	return &n
}

func (ObjectNode) Type() json.Type {
	return json.TypeObject
}

func (n *ObjectNode) Grow(lex lexeme.LexEvent) (Node, bool) {
	if n.waitingForChild {
		n.waitingForChild = false
		child := NewNode(lex)
		n.addChild(child)
		return child, true
	}

	switch lex.Type() {
	case lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectValueEnd:

	case lexeme.KeyShortcutEnd:
		key := lex.Value().Unquote().String()
		n.addKey(key, lex.Value().IsUserTypeName(), lex) // can panic

	case lexeme.ObjectKeyEnd:
		key := lex.Value().Unquote().String()
		n.addKey(key, lex.Value().IsUserTypeName(), lex) // can panic

	case lexeme.ObjectValueBegin:
		n.waitingForChild = true

	case lexeme.ObjectEnd:
		return n.parent, false

	default:
		panic(`Unexpected lexical event "` + lex.Type().String() + `" in object node`)
	}

	return n, false
}

func (n ObjectNode) Children() []Node {
	return n.children
}

func (n ObjectNode) Len() int {
	return len(n.children)
}

func (n ObjectNode) Child(key string) (Node, bool) {
	i, ok := n.keys.Get(key)
	if ok {
		return n.children[i.Index], true
	}
	return nil, false
}

func (n *ObjectNode) addKey(key string, isShortcut bool, lex lexeme.LexEvent) {
	if n.keys.Has(key) {
		panic(errors.Format(errors.ErrDuplicateKeysInSchema, key))
	}

	// Save child node index into map for faster search.
	n.keys.Set(key, objectNodeKey{
		Index:      len(n.children),
		IsShortcut: isShortcut,
		Lex:        lex,
	})
}

func (n *ObjectNode) addChild(child Node) {
	child.SetParent(n)
	n.children = append(n.children, child)
}

func (n *ObjectNode) AddChild(key ObjectNodeKey, child Node) {
	n.addKey(key.Name, key.IsShortcut, key.Lex) // can panic
	n.addChild(child)
}

type ObjectNodeKey struct {
	Name       string
	IsShortcut bool
	Lex        lexeme.LexEvent
}

func (n ObjectNode) Key(index int) ObjectNodeKey {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for kv := range n.keys.IterateContext(ctx) {
		if kv.Value.Index == index {
			return ObjectNodeKey{
				Name:       kv.Key,
				IsShortcut: kv.Value.IsShortcut,
				Lex:        kv.Value.Lex,
			}
		}
	}
	panic(fmt.Sprintf(`Schema key not found in index %d`, index))
}

type ObjectNodeKeysIterator interface {
	Iterate() <-chan objectNodeKeysItem
}

func (n ObjectNode) Keys() ObjectNodeKeysIterator {
	return n.keys
}

func (n ObjectNode) IndentedTreeString(depth int) string {
	indent := strings.Repeat("\t", depth)

	var str strings.Builder
	str.WriteString(n.IndentedNodeString(depth))

	for index, childNode := range n.children {
		key := n.Key(index) // can panic: Index not found in array
		str.WriteString(indent + "\t\"" + key.Name + "\":\n")
		str.WriteString(childNode.IndentedTreeString(depth + 2))
	}

	return str.String()
}

func (n ObjectNode) IndentedNodeString(depth int) string {
	indent := strings.Repeat("\t", depth)

	var str strings.Builder
	str.WriteString(indent + "* " + n.Type().String() + "\n")

	for kv := range n.constraints.Iterate() {
		str.WriteString(indent + "* " + kv.Value.String() + "\n")
	}

	return str.String()
}

func (n *ObjectNode) ASTNode() (jschema.ASTNode, error) {
	an := astNodeFromNode(n)

	var err error
	an.Properties, err = n.collectASTProperties()
	if err != nil {
		return jschema.ASTNode{}, err
	}

	return an, nil
}

func (n *ObjectNode) collectASTProperties() (*jschema.ASTNodes, error) {
	pp := &jschema.ASTNodes{}
	for kv := range n.keys.Iterate() {
		c := n.children[kv.Value.Index]
		cn, err := c.ASTNode()
		if err != nil {
			return nil, err
		}

		cn.IsKeyShortcut = kv.Value.IsShortcut

		pp.Set(kv.Key, cn)
	}
	return pp, nil
}
