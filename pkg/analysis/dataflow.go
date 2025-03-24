package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// DataFlowContext tracks taint propagation and data flow
type DataFlowContext struct {
	// Track variables that contain tainted data
	TaintedVars map[string]bool
	// Track intermediate variables that propagate taint
	IntermediateVars map[string]bool
	// Track known sources of user input
	Sources map[string]bool
	// Track sanitization functions
	Sanitizers map[string]bool
	// Track variable definitions and their values
	VarDefinitions map[string]*sitter.Node
	// Source code for content lookup
	SourceCode []byte
}

// NewDataFlowContext creates a new context with default sources and sanitizers
func NewDataFlowContext(sourceCode []byte) *DataFlowContext {
	return &DataFlowContext{
		TaintedVars:      make(map[string]bool),
		IntermediateVars: make(map[string]bool),
		Sources: map[string]bool{
			"req.body":    true,
			"req.query":   true,
			"req.params":  true,
			"req.cookies": true,
			// Add more sources as needed
		},
		Sanitizers: map[string]bool{
			"escape":       true,
			"mysql.escape": true,
			"parameterize": true,
			"sanitize":     true,
		},
		VarDefinitions: make(map[string]*sitter.Node),
		SourceCode:     sourceCode,
	}
}

// TrackAssignment analyzes assignment expressions for taint propagation
func (dc *DataFlowContext) TrackAssignment(node *sitter.Node) {
	if node.Type() != "assignment_expression" && node.Type() != "variable_declarator" {
		return
	}

	var leftNode, rightNode *sitter.Node
	if node.Type() == "assignment_expression" {
		leftNode = node.ChildByFieldName("left")
		rightNode = node.ChildByFieldName("right")
	} else {
		leftNode = node.ChildByFieldName("name")
		rightNode = node.ChildByFieldName("value")
	}

	if leftNode == nil || rightNode == nil {
		return
	}

	leftName := leftNode.Content(dc.SourceCode)

	// Track variable definition
	dc.VarDefinitions[leftName] = rightNode

	// Propagate taint
	if dc.IsNodeTainted(rightNode) {
		dc.TaintedVars[leftName] = true
		dc.IntermediateVars[leftName] = true
	}
}

// IsNodeTainted checks if a node contains or propagates tainted data
func (dc *DataFlowContext) IsNodeTainted(node *sitter.Node) bool {
	if node == nil {
		return false
	}

	switch node.Type() {
	case "identifier":
		varName := node.Content(dc.SourceCode)
		// Check if it's a known source or tainted variable
		if dc.Sources[varName] || dc.TaintedVars[varName] {
			return true
		}
		// Check variable definition if available
		if defNode, exists := dc.VarDefinitions[varName]; exists {
			return dc.IsNodeTainted(defNode)
		}

	case "member_expression":
		// Check for known dangerous patterns like req.body
		fullExpr := node.Content(dc.SourceCode)
		if dc.Sources[fullExpr] {
			return true
		}

		// Check object and property
		object := node.ChildByFieldName("object")
		return dc.IsNodeTainted(object)

	case "template_string":
		// Check for tainted interpolations
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "template_substitution" && dc.IsNodeTainted(child.NamedChild(0)) {
				return true
			}
		}

	case "binary_expression":
		// Check both sides of the expression
		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")
		return dc.IsNodeTainted(left) || dc.IsNodeTainted(right)
	}

	return false
}

// IsSanitized checks if a node has been properly sanitized
func (dc *DataFlowContext) IsSanitized(node *sitter.Node) bool {
	if node == nil {
		return false
	}

	// Check if the node is a call to a sanitizer function
	if node.Type() == "call_expression" {
		funcNode := node.ChildByFieldName("function")
		if funcNode != nil {
			funcName := funcNode.Content(dc.SourceCode)
			return dc.Sanitizers[funcName]
		}
	}

	return false
}
