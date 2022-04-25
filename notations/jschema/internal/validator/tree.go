package validator

import (
	"fmt"
	"j/schema/errors"
	"j/schema/internal/lexeme"
	"j/schema/internal/logger"
)

// A tree of validators.
//
// The tree consists of nodes (validators).
// Each node contains a pointer to the parent and validator from the schema package.
// Leaf, is the node without children.
// Root, is the top node in a tree. The parent of the root node is nil
//
// The "tree" structure contains the leaves in which to pass the LexEvent for
// validation.

type Tree struct {
	// nextIndex the index for next leaf.
	nextIndex int

	// leaves a list of all the leaves of the tree.
	leaves map[int]validator

	// log a logger for logging and debugging tool.
	log logger.Logger

	// leavesIndexes the list of indexes for the leaves. Defined in struct (not
	// in method) to optimize memory allocation.
	leavesIndexes []int
}

func NewTree(list []validator, log logger.Logger) Tree {
	t := Tree{
		nextIndex:     0,
		leaves:        make(map[int]validator, 5),
		log:           log,
		leavesIndexes: make([]int, 0, 5),
	}
	for _, v := range list {
		t.addLeaf(v)
	}

	t.log.Notice("Tree of validators:")
	t.log.Default(t.string())

	return t
}

// FeedLeaves returns true if the validation of the entire tree is completed. There
// are no more validators left in the tree.
func (t *Tree) FeedLeaves(jsonLex lexeme.LexEvent) bool {
	// A new array with the indexes of the leaves of the tree, to iterate on it.
	// The tree will change during the iteration.
	t.setLeavesIndexes()
	errorsCount := 0

	t.log.Notice("\tPut json lex:")
	lexType := jsonLex.Type()
	switch lexType {
	case lexeme.LiteralEnd, lexeme.ObjectKeyEnd:
		t.log.Warning("\t" + lexType.String() + "=" + jsonLex.Value().String())
	default:
		t.log.Warning("\t" + lexType.String())
	}
	t.log.Default("")

	var err error

	for _, indexOfLeaf := range t.leavesIndexes {
		if leaf, ok := t.leaves[indexOfLeaf]; ok {
			err = t.feedLeaf(leaf, jsonLex, indexOfLeaf) // can panic
			if err != nil {
				errorsCount++
			}
		}
	}

	if errorsCount == len(t.leavesIndexes) {
		if len(t.leavesIndexes) == 1 {
			panic(err)
		} else {
			panic(lexeme.NewLexEventError(jsonLex, errors.ErrOrRuleSetValidation))
		}
	}

	if len(t.leaves) == 0 {
		return true
	}
	return false
}

func (t *Tree) setLeavesIndexes() {
	t.leavesIndexes = t.leavesIndexes[:0]
	for i := range t.leaves {
		t.leavesIndexes = append(t.leavesIndexes, i)
	}
}

// Passes the LexEvent to the validator. Based on the results changes the tree.
// Removes or adds new validators to the tree.
// Returns common.DocumentError if an error is found during node validation.
func (t *Tree) feedLeaf(leaf validator, jsonLex lexeme.LexEvent, indexOfLeaf int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(errors.DocumentError)
			if ok {
				t.log.Error(err.Error() + "\n")
				delete(t.leaves, indexOfLeaf)

				t.log.Notice("Tree of validators:")
				t.log.Default(t.string())
			} else {
				panic(r)
			}
		}
	}()

	t.log.Info("\t\tTo validator of schema node")
	t.log.Warning(fmt.Sprintf("\t\t[%p]:", leaf.node()))
	t.log.Default(leaf.node().IndentedNodeString(2))

	children, done := leaf.feed(jsonLex) // can panic

	if done { // validation of node completed
		t.log.Notice("Trim node")
		t.log.Warning(fmt.Sprintf("[%p]\n", leaf.node()))

		parent := leaf.parent()
		leaf.setParent(nil) // remove the pointer to simplify garbage collection in the future
		if parent == nil {
			delete(t.leaves, indexOfLeaf)
		} else {
			t.leaves[indexOfLeaf] = parent // step back to parent
		}

		t.log.Notice("Tree of validators:")
		t.log.Default(t.string())
	} else if children != nil { // children found
		t.log.Notice("Append child(ren) to node")
		t.log.Warning(fmt.Sprintf("[%p]\n", leaf.node()))

		for j, child := range children {
			if j == 0 {
				// Forget/replace the current leaf. He becomes branch, parent for
				// first child.
				t.leaves[indexOfLeaf] = child
			} else {
				t.addLeaf(child) // append new child leaf to tree
			}
		}

		t.log.Notice("Tree of validators:")
		t.log.Default(t.string())
	}

	return nil
}

func (t *Tree) addLeaf(v validator) {
	t.leaves[t.nextIndex] = v
	t.nextIndex++
}

func (t Tree) string() string {
	return newTreeLog(t).string()
}

// Truncates all leaves to the specified node.
// If the leaves whose parents have some specified node, leaving only one leaf after trimming.
// func (t *tree) trimBranchesTo(target validator) {
// 	targetAreFound := false
// 	for i := range t.leaves {
// 		for {
// 			leaf := t.leaves[i]
// 			if leaf != target {
// 				// trim leaf
// 				parent := leaf.parent()
// 				leaf.setParent(nil) // remove the pointer to simplify garbage collection in the future
// 				if parent == nil {
// 					delete(t.leaves, i) // trim branch
// 					break
// 				} else {
// 					t.leaves[i] = parent // trim leaf
// 				}
// 			} else { // leaf == target
// 				if targetAreFound {
// 					delete(t.leaves, i) // unnecessary leaf is removed completely
// 				} else {
// 					targetAreFound = true // first branch with the target is found
// 				}
// 				break
// 			}
// 		}
// 	}
// }