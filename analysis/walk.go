package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// Walker is an interface that dictates what to do when
// entering and leaving each node during the pre-order traversal
// of a tree.
// To traverse post-order, use the `OnLeaveNode` callback.
type Walker interface {
	// OnEnterNode is called when the walker enters a node.
	// The boolean return value indicates whether the walker should
	// continue walking the sub-tree of this node.
	OnEnterNode(node *sitter.Node) bool
	// OnLeaveNode is called when the walker leaves a node.
	// This is called after all the children of the node have been visited and explored.
	OnLeaveNode(node *sitter.Node)
}

func WalkTree(node *sitter.Node, walker Walker) {
	goInside := walker.OnEnterNode(node)
	if goInside {
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			WalkTree(child, walker)
		}
	}

	walker.OnLeaveNode(node)
}

// ChildrenWithFieldName returns all the children of a node
// with a specific field name.
// Tree-sitter can have multiple children with the same field name.
func ChildrenWithFieldName(node *sitter.Node, fieldName string) []*sitter.Node {
	var children []*sitter.Node
	for i := 0; i < int(node.ChildCount()); i++ {
		if node.FieldNameForChild(i) == fieldName {
			child := node.Child(i)
			children = append(children, child)
		}
	}

	return children
}

// FindMatchingChild iterates over all children of a node—both named and unnamed—and returns the
// first child that matches the predicate function.
func FindMatchingChild(node *sitter.Node, predicate func(*sitter.Node) bool) *sitter.Node {
	nChildren := int(node.ChildCount())

	for i := 0; i < nChildren; i++ {
		child := node.Child(i)
		if predicate(child) {
			return child
		}
	}

	return nil
}

func ChildrenOfType(node *sitter.Node, nodeType string) []*sitter.Node {
	nChildren := int(node.ChildCount())
	var results []*sitter.Node
	for i := 0; i < nChildren; i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			results = append(results, child)
		}
	}
	return results
}

func ChildWithFieldName(node *sitter.Node, fieldName string) *sitter.Node {
	nChildren := int(node.NamedChildCount())
	for i := 0; i < nChildren; i++ {
		if node.FieldNameForChild(i) == fieldName {
			return node.Child(i)
		}
	}

	return nil
}

func FirstChildOfType(node *sitter.Node, nodeType string) *sitter.Node {
	nChildren := int(node.ChildCount())
	for i := 0; i < nChildren; i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			return child
		}
	}

	return nil
}
