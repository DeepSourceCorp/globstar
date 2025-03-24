package javascript

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var SQLInjection = &analysis.Analyzer{
	Name:        "sql_injection",
	Language:    analysis.LangJs,
	Description: "Using raw SQL queries with unvalidated input can lead to SQL injection vulnerabilities",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Run:         detectSQLInjection,
}

// DataFlowNode represents a node in our data flow graph
type DataFlowNode struct {
	Node           *sitter.Node
	Tainted        bool
	Sources        []*DataFlowNode
	Sanitized      bool
	SanitizedPaths map[string]bool // Track which paths have been sanitized
}

// Map to track function calls that sanitize inputs
var sanitizerFunctions = map[string]bool{
	"escape":                   true,
	"sanitize":                 true,
	"parameterize":             true,
	"mysql_real_escape_string": true,
}

func detectSQLInjection(pass *analysis.Pass) (interface{}, error) {
	// Map of vulnerable function names to watch for
	vulnerableFunctions := map[string]bool{
		"query":             true,
		"raw":               true,
		"$queryRawUnsafe":   true,
		"$executeRawUnsafe": true,
		"execute":           true,
		"db.query":          true,
		"connection.query":  true,
	}

	// Map to track variable definitions and their data flow nodes
	dataFlowGraph := make(map[string]*DataFlowNode)

	// Identify sources of user input (entry points)
	userInputSources := map[string]bool{
		"req.body":                true,
		"req.params":              true,
		"req.query":               true,
		"request.body":            true,
		"request.params":          true,
		"request.query":           true,
		"document.getElementById": true,
		"$_GET":                   true,
		"$_POST":                  true,
		"process.argv":            true,
	}

	// First pass: build initial data flow graph
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}

		// Track variable declarations and assignments
		if node.Type() == "variable_declarator" || node.Type() == "assignment_expression" {
			var nameNode, valueNode *sitter.Node

			if node.Type() == "variable_declarator" {
				nameNode = node.ChildByFieldName("name")
				valueNode = node.ChildByFieldName("value")
			} else { // assignment_expression
				nameNode = node.ChildByFieldName("left")
				valueNode = node.ChildByFieldName("right")
			}

			if nameNode != nil && nameNode.Type() == "identifier" && valueNode != nil {
				varName := nameNode.Content(pass.FileContext.Source)

				// Create or update data flow node
				dfNode := &DataFlowNode{
					Node:           valueNode,
					Tainted:        false,
					Sources:        []*DataFlowNode{},
					Sanitized:      false,
					SanitizedPaths: make(map[string]bool),
				}

				// Check if the value is a source of user input
				if isUserInput(valueNode, pass.FileContext.Source, userInputSources) {
					dfNode.Tainted = true
				} else {
					switch valueNode.Type() {
					case "identifier":
						// If value is another variable, link to its data flow node
						sourceVarName := valueNode.Content(pass.FileContext.Source)
						if sourceNode, exists := dataFlowGraph[sourceVarName]; exists {
							dfNode.Sources = append(dfNode.Sources, sourceNode)
							dfNode.Tainted = sourceNode.Tainted
							dfNode.Sanitized = sourceNode.Sanitized

							// Copy sanitized paths
							for path, sanitized := range sourceNode.SanitizedPaths {
								dfNode.SanitizedPaths[path] = sanitized
							}
						}

					case "call_expression":
						// Check if this is a sanitizer function
						isSanitized, path := isSanitizerCall(valueNode, pass.FileContext.Source)
						if isSanitized {
							dfNode.Sanitized = true
							dfNode.SanitizedPaths[path] = true
						}

						// Track data flow through function arguments
						handleCallExpressionDataFlow(valueNode, dfNode, dataFlowGraph, pass.FileContext.Source)

					case "binary_expression":
						// Handle string concatenation data flow
						handleBinaryExpressionDataFlow(valueNode, dfNode, dataFlowGraph, pass.FileContext.Source)

					case "template_string":
						// Handle template literals data flow
						handleTemplateStringDataFlow(valueNode, dfNode, dataFlowGraph, pass.FileContext.Source)
					}
				}

				dataFlowGraph[varName] = dfNode
			}
		}
	})

	// Propagate taint through the data flow graph
	propagateTaint(dataFlowGraph)

	// Second pass: detect SQL injection vulnerabilities using the data flow graph
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil || node.Type() != "call_expression" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			return
		}

		// Get the function name
		var funcName string
		if funcNode.Type() == "member_expression" {
			propertyNode := funcNode.ChildByFieldName("property")
			if propertyNode != nil {
				funcName = propertyNode.Content(pass.FileContext.Source)

				// Also check for object.method pattern (e.g., db.query)
				objectNode := funcNode.ChildByFieldName("object")
				if objectNode != nil && objectNode.Type() == "identifier" {
					objectName := objectNode.Content(pass.FileContext.Source)
					compositeFunc := objectName + "." + funcName
					if vulnerableFunctions[compositeFunc] {
						funcName = compositeFunc
					}
				}
			}
		} else if funcNode.Type() == "identifier" {
			funcName = funcNode.Content(pass.FileContext.Source)
		}

		// Check if this is a function that executes raw SQL
		if !vulnerableFunctions[funcName] {
			return
		}

		// Get the arguments of the function
		args := node.ChildByFieldName("arguments")
		if args == nil || args.NamedChildCount() == 0 {
			return
		}

		// Get the first argument (SQL query)
		firstArg := args.NamedChild(0)
		if firstArg == nil {
			return
		}

		// Check if the argument is vulnerable using data flow analysis
		if isVulnerableWithDataFlow(firstArg, pass.FileContext.Source, dataFlowGraph) {
			pass.Report(pass, node, "Potential SQL injection vulnerability detected, use parameterized queries instead")
		}
	})

	return nil, nil
}

// Check if node represents a user input source
func isUserInput(node *sitter.Node, sourceCode []byte, userInputSources map[string]bool) bool {
	if node == nil {
		return false
	}

	if node.Type() == "member_expression" {
		fullExpression := node.Content(sourceCode)
		if userInputSources[fullExpression] {
			return true
		}

		// Handle nested expressions like req.body.username
		for source := range userInputSources {
			if len(fullExpression) >= len(source) && fullExpression[:len(source)] == source {
				return true
			}
		}
	} else if node.Type() == "call_expression" {
		funcNode := node.ChildByFieldName("function")
		if funcNode != nil {
			funcName := funcNode.Content(sourceCode)
			if userInputSources[funcName] {
				return true
			}
		}
	}

	return false
}

// Check if a call is to a sanitizer function
func isSanitizerCall(node *sitter.Node, sourceCode []byte) (bool, string) {
	if node == nil || node.Type() != "call_expression" {
		return false, ""
	}

	funcNode := node.ChildByFieldName("function")
	if funcNode == nil {
		return false, ""
	}

	var funcName string
	var path string

	if funcNode.Type() == "member_expression" {
		propertyNode := funcNode.ChildByFieldName("property")
		if propertyNode != nil {
			funcName = propertyNode.Content(sourceCode)
			path = funcName
		}
	} else if funcNode.Type() == "identifier" {
		funcName = funcNode.Content(sourceCode)
		path = funcName
	}

	return sanitizerFunctions[funcName], path
}

// Handle data flow through binary expressions (like string concatenation)
func handleBinaryExpressionDataFlow(node *sitter.Node, dfNode *DataFlowNode, dataFlowGraph map[string]*DataFlowNode, sourceCode []byte) {
	if node == nil || node.Type() != "binary_expression" {
		return
	}

	left := node.ChildByFieldName("left")
	right := node.ChildByFieldName("right")
	operator := node.ChildByFieldName("operator")

	// Only process string concatenation
	if operator != nil && operator.Content(sourceCode) == "+" {
		// Check left side
		if left != nil && left.Type() == "identifier" {
			leftVarName := left.Content(sourceCode)
			if sourceNode, exists := dataFlowGraph[leftVarName]; exists {
				dfNode.Sources = append(dfNode.Sources, sourceNode)
				if sourceNode.Tainted {
					dfNode.Tainted = true
				}
			}
		}

		// Check right side
		if right != nil && right.Type() == "identifier" {
			rightVarName := right.Content(sourceCode)
			if sourceNode, exists := dataFlowGraph[rightVarName]; exists {
				dfNode.Sources = append(dfNode.Sources, sourceNode)
				if sourceNode.Tainted {
					dfNode.Tainted = true
				}
			}
		}

		// Recursively process nested expressions
		if left != nil && left.Type() == "binary_expression" {
			handleBinaryExpressionDataFlow(left, dfNode, dataFlowGraph, sourceCode)
		}

		if right != nil && right.Type() == "binary_expression" {
			handleBinaryExpressionDataFlow(right, dfNode, dataFlowGraph, sourceCode)
		}
	}
}

// Handle data flow through template strings
func handleTemplateStringDataFlow(node *sitter.Node, dfNode *DataFlowNode, dataFlowGraph map[string]*DataFlowNode, sourceCode []byte) {
	if node == nil || node.Type() != "template_string" {
		return
	}

	// Look for template substitutions
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type() == "template_substitution" {
			// Get the expression inside ${}
			exprNode := child.NamedChild(0)
			if exprNode != nil && exprNode.Type() == "identifier" {
				varName := exprNode.Content(sourceCode)
				if sourceNode, exists := dataFlowGraph[varName]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)
					if sourceNode.Tainted {
						dfNode.Tainted = true
					}
				}
			}
		}
	}
}

// Handle data flow through function arguments
func handleCallExpressionDataFlow(node *sitter.Node, dfNode *DataFlowNode, dataFlowGraph map[string]*DataFlowNode, sourceCode []byte) {
	if node == nil || node.Type() != "call_expression" {
		return
	}

	args := node.ChildByFieldName("arguments")
	if args == nil {
		return
	}

	// Check each argument for taint
	for i := 0; i < int(args.NamedChildCount()); i++ {
		arg := args.NamedChild(i)
		if arg == nil {
			continue
		}

		if arg.Type() == "identifier" {
			argName := arg.Content(sourceCode)
			if sourceNode, exists := dataFlowGraph[argName]; exists {
				dfNode.Sources = append(dfNode.Sources, sourceNode)
				if sourceNode.Tainted {
					dfNode.Tainted = true
				}
			}
		}
	}
}

// Propagate taint through the data flow graph
func propagateTaint(dataFlowGraph map[string]*DataFlowNode) {
	// Repeat until no changes are made
	changed := true
	for changed {
		changed = false

		for _, dfNode := range dataFlowGraph {
			// If node is already tainted, no need to check further
			if dfNode.Tainted {
				continue
			}

			// Check if any source is tainted
			for _, source := range dfNode.Sources {
				if source.Tainted && !dfNode.Tainted {
					dfNode.Tainted = true
					changed = true
					break
				}
			}
		}
	}
}

// Check if a node is vulnerable using data flow analysis
func isVulnerableWithDataFlow(node *sitter.Node, sourceCode []byte, dataFlowGraph map[string]*DataFlowNode) bool {
	if node == nil {
		return false
	}

	switch node.Type() {
	case "identifier":
		// Look up in the data flow graph
		varName := node.Content(sourceCode)
		if dfNode, exists := dataFlowGraph[varName]; exists {
			// Node is vulnerable if it's tainted and not sanitized
			return dfNode.Tainted && !dfNode.Sanitized
		}

	case "binary_expression":
		// Check if either side is vulnerable
		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")

		return isVulnerableWithDataFlow(left, sourceCode, dataFlowGraph) ||
			isVulnerableWithDataFlow(right, sourceCode, dataFlowGraph)

	case "template_string":
		// Check each substitution
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child != nil && child.Type() == "template_substitution" {
				exprNode := child.NamedChild(0)
				if isVulnerableWithDataFlow(exprNode, sourceCode, dataFlowGraph) {
					return true
				}
			}
		}

	case "call_expression":
		// Check if this is a sanitizer function
		isSanitized, _ := isSanitizerCall(node, sourceCode)
		if isSanitized {
			return false
		}

		// Check arguments
		args := node.ChildByFieldName("arguments")
		if args != nil {
			for i := 0; i < int(args.NamedChildCount()); i++ {
				arg := args.NamedChild(i)
				if isVulnerableWithDataFlow(arg, sourceCode, dataFlowGraph) {
					return true
				}
			}
		}
	}

	return false
}
