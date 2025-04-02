package javascript

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

// Identify sources of user input (entry points)
var userInputSources = map[string]bool{
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
	"req.users":               true,
	"req.user.id":             true,
}
var SQLInjectionDFG = &analysis.Analyzer{
	Name:        "sql_injection",
	Language:    analysis.LangJs,
	Description: "Using raw SQL queries with unvalidated input can lead to SQL injection vulnerabilities",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Run:         detectSQLInjectionDFG,
	Requires:    []*analysis.Analyzer{ScopeAnalyzer},
}

// DataFlowNode represents a node in our data flow graph
type DataFlowNode struct {
	Node           *sitter.Node
	Tainted        bool
	Sources        []*DataFlowNode
	Sanitized      bool
	SanitizedPaths map[string]bool // Track which paths have been sanitized
	Scope          *analysis.Scope
	Variable       *analysis.Variable // Tracking variable instead of the varName helps in handling scope
	FuncDef        *FunctionDefinition
}

type FunctionDefinition struct {
	Node       *sitter.Node
	Parameters []*analysis.Variable
	Body       *sitter.Node
	Scope      *analysis.Scope
}

// Map to track function calls that sanitize inputs
var sanitizerFunctions = map[string]bool{
	"escape":                   true,
	"sanitize":                 true,
	"parameterize":             true,
	"mysql_real_escape_string": true,
}

// Map of vulnerable function names to watch for
var vulnerableFunctions = map[string]bool{
	"query":             true,
	"raw":               true,
	"$queryRawUnsafe":   true,
	"$executeRawUnsafe": true,
	"execute":           true,
	"db.query":          true,
	"connection.query":  true,
}

var functionDefinitions = make(map[string]*FunctionDefinition)

func detectSQLInjectionDFG(pass *analysis.Pass) (interface{}, error) {

	// Map to track variable definitions and their data flow nodes
	dataFlowGraph := make(map[*analysis.Variable]*DataFlowNode)

	scopeResult, err := buildScopeTree(pass)
	if err != nil {
		return nil, fmt.Errorf("failed to build the scope tree \n")
	}

	scopeTree := scopeResult.(*analysis.ScopeTree)

	// First pass: build initial data flow graph
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}

		currentScope := scopeTree.GetScope(node)
		if currentScope == nil {
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
				// fmt.Printf("%v = %v\n", nameNode.Content(pass.FileContext.Source), valueNode.Content(pass.FileContext.Source))
			}

			if nameNode != nil && nameNode.Type() == "identifier" && valueNode != nil {
				varName := nameNode.Content(pass.FileContext.Source)

				variable := currentScope.Lookup(varName)
				// Create or update data flow node
				dfNode := &DataFlowNode{
					Node:           valueNode,
					Tainted:        false,
					Sources:        []*DataFlowNode{},
					Sanitized:      false,
					SanitizedPaths: make(map[string]bool),
					Scope:          currentScope,
					Variable:       variable,
				}

				// Check if the value is a source of user input
				if isUserInput(valueNode, pass.FileContext.Source, userInputSources) {
					dfNode.Tainted = true
				} else {
					switch valueNode.Type() {
					case "identifier":
						// If value is another variable, link to its data flow node
						sourceVarName := valueNode.Content(pass.FileContext.Source)
						currVar := currentScope.Lookup(sourceVarName)
						if sourceNode, exists := dataFlowGraph[currVar]; exists {
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
						handleCallExpressionDataFlow(valueNode, dfNode, dataFlowGraph, pass.FileContext.Source, currentScope)

					case "binary_expression":
						// Handle string concatenation data flow
						handleBinaryExpressionDataFlow(valueNode, dfNode, dataFlowGraph, pass.FileContext.Source, currentScope)

					case "template_string":
						// Handle template literals data flow
						handleTemplateStringDataFlow(valueNode, dfNode, dataFlowGraph, pass.FileContext.Source, currentScope)
					}
				}

				dataFlowGraph[variable] = dfNode

			}
		}

		if node.Type() == "function_declaration" {
			funcNameNode := node.ChildByFieldName("name")
			if funcNameNode == nil {
				return
			}
			funcName := funcNameNode.Content(pass.FileContext.Source)

			funcDef := &FunctionDefinition{
				Node:  node,
				Body:  node.ChildByFieldName("body"),
				Scope: currentScope,
			}

			params := node.ChildByFieldName("parameters")
			if params != nil {
				for i := 0; i < int(params.NamedChildCount()); i++ {
					param := params.NamedChild(i)
					if param.Type() == "identifier" {
						paramName := param.Content(pass.FileContext.Source)
						paramVar := currentScope.Lookup(paramName)
						if paramVar != nil {
							funcDef.Parameters = append(funcDef.Parameters, paramVar)
						}

					}
				}

			}

			functionDefinitions[funcName] = funcDef
		}
	})

	// Propagate taint through the data flow graph
	propagateTaint(dataFlowGraph)

	// Second pass: detect SQL injection vulnerabilities using the data flow graph
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}
		currentScope := scopeTree.GetScope(node)

		switch node.Type() {
		case "call_expression":
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
				if IsVulnerableWithDataFlow(node, pass.FileContext.Source, dataFlowGraph, currentScope) {
					pass.Report(pass, node, "Found SQL-Injection attempt")
					return
				}

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
			if IsVulnerableWithDataFlow(firstArg, pass.FileContext.Source, dataFlowGraph, currentScope) {
				pass.Report(pass, node, "Found SQL-Injection attempt")
				return
			}

		case "if_statement":
			// fmt.Println(node.Content(pass.FileContext.Source))
			if IsVulnerableWithDataFlow(node, pass.FileContext.Source, dataFlowGraph, currentScope) {
				pass.Report(pass, node, "Found SQL-Injection attempt")
				return
			}
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
func handleBinaryExpressionDataFlow(node *sitter.Node, dfNode *DataFlowNode, dataFlowGraph map[*analysis.Variable]*DataFlowNode, sourceCode []byte, scope *analysis.Scope) {
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
			if variable := scope.Lookup(leftVarName); variable != nil {
				if sourceNode, exists := dataFlowGraph[variable]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)
					if sourceNode.Tainted {
						dfNode.Tainted = true
					}
				}
			}

		}

		// Check right side
		if right != nil && right.Type() == "identifier" {
			rightVarName := right.Content(sourceCode)
			if variable := scope.Lookup(rightVarName); variable != nil {
				if sourceNode, exists := dataFlowGraph[variable]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)
					if sourceNode.Tainted {
						dfNode.Tainted = true
					}
				}
			}

		}

		// Recursively process nested expressions
		if left != nil && left.Type() == "binary_expression" {
			handleBinaryExpressionDataFlow(left, dfNode, dataFlowGraph, sourceCode, scope)
		}

		if right != nil && right.Type() == "binary_expression" {
			handleBinaryExpressionDataFlow(right, dfNode, dataFlowGraph, sourceCode, scope)
		}
	}
}

// Handle data flow through template strings
func handleTemplateStringDataFlow(node *sitter.Node, dfNode *DataFlowNode, dataFlowGraph map[*analysis.Variable]*DataFlowNode, sourceCode []byte, scope *analysis.Scope) {
	if node == nil || node.Type() != "template_string" {
		return
	}

	// Look for template substitutions
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type() == "template_substitution" {
			// Get the ression inside ${}
			exprNode := child.NamedChild(0)
			// fmt.Printf("Expression inside template: %v\n", exprNode)
			if exprNode != nil && exprNode.Type() == "identifier" {
				varName := exprNode.Content(sourceCode)
				if variable := scope.Lookup(varName); variable != nil {
					// fmt.Printf("Variable name: %v\n", varName)

					if sourceNode, exists := dataFlowGraph[variable]; exists {
						dfNode.Sources = append(dfNode.Sources, sourceNode)
						if sourceNode.Tainted {
							dfNode.Tainted = true
						}
					}
				}

			}
		}
	}
}

// Handle data flow through function arguments
func handleCallExpressionDataFlow(node *sitter.Node, dfNode *DataFlowNode, dataFlowGraph map[*analysis.Variable]*DataFlowNode, sourceCode []byte, scope *analysis.Scope) {
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
			if variable := scope.Lookup(argName); variable != nil {
				if sourceNode, exists := dataFlowGraph[variable]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)
					if sourceNode.Tainted {
						dfNode.Tainted = true
					}
				}
			}

		}
	}
}

// Propagate taint through the data flow graph
func propagateTaint(dataFlowGraph map[*analysis.Variable]*DataFlowNode) {
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
func IsVulnerableWithDataFlow(node *sitter.Node, sourceCode []byte, dataFlowGraph map[*analysis.Variable]*DataFlowNode, scope *analysis.Scope) bool {
	if node == nil {
		return false
	}
	switch node.Type() {
	case "identifier":
		// Look up in the data flow graph
		varName := node.Content(sourceCode)
		currentScope := scope
		for currentScope != nil {
			if variable := currentScope.Lookup(varName); variable != nil {
				if dfNode, exists := dataFlowGraph[variable]; exists {
					// Node is vulnerable if it's tainted and not sanitized
					return dfNode.Tainted && !dfNode.Sanitized
				}
			}
			currentScope = currentScope.Upper
		}

	case "binary_expression":
		// Check if either side is vulnerable
		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")

		return IsVulnerableWithDataFlow(left, sourceCode, dataFlowGraph, scope) ||
			IsVulnerableWithDataFlow(right, sourceCode, dataFlowGraph, scope)

	case "template_string":
		// Check each substitution
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child != nil && child.Type() == "template_substitution" {
				exprNode := child.NamedChild(0)
				if IsVulnerableWithDataFlow(exprNode, sourceCode, dataFlowGraph, scope) {
					return true
				}
			}
		}

	case "call_expression":
		funcNode := node.ChildByFieldName("function")
		if funcNode != nil && funcNode.Type() == "identifier" {
			funcName := funcNode.Content(sourceCode)
			args := node.ChildByFieldName("arguments")
			if args != nil {
				if analyzeUserDefinedFunction(funcName, args, sourceCode, dataFlowGraph, scope) {
					return true
				}
			}
		}

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
				if IsVulnerableWithDataFlow(arg, sourceCode, dataFlowGraph, scope) {
					return true
				}
			}
		}
	case "if_statement":
		condition := node.ChildByFieldName("condition")
		for i := 0; i < int(node.NamedChildCount()); i++ {
			if IsVulnerableWithDataFlow(condition, sourceCode, dataFlowGraph, scope) {
				return true
			}
		}

		consequence := node.ChildByFieldName("consequence")
		if analyzeBlockForVulnerabilities(consequence, sourceCode, dataFlowGraph, scope) {
			return true
		}

		alternative := node.ChildByFieldName("alternative")
		if alternative != nil && analyzeBlockForVulnerabilities(alternative, sourceCode, dataFlowGraph, scope) {
			return true
		}
	}

	return false
}

func analyzeBlockForVulnerabilities(node *sitter.Node, sourceCode []byte, dataFlowGraph map[*analysis.Variable]*DataFlowNode, scope *analysis.Scope) bool {
	if node == nil {
		return false
	}
	switch node.Type() {
	case "statement_block":
		// Analyze each statement in the block
		for i := 0; i < int(node.NamedChildCount()); i++ {
			stmt := node.NamedChild(i)
			switch stmt.Type() {
			case "lexical_declaration":
				// for i := 0; i < int(stmt.NamedChildCount()); i++ {
				// 	fmt.Println(stmt.Child(i).Content(sourceCode))
				// }
				declarators := analysis.ChildrenOfType(stmt, "variable_declarator")
				for _, declarator := range declarators {
					// Get the name node of the declarator
					nameNode := declarator.ChildByFieldName("name")
					nameVar := scope.Lookup(nameNode.Content(sourceCode))
					if nameVar != nil {
						// fmt.Println(dataFlowGraph[nameVar].Tained)
						// if IsVulnerableWithDataFlow(nameNode, sourceCode, dataFlowGraph, scope) {
						// 	fmt.Println("Tainted.")
						// 	return true
						// }
					}
				}
			}

		}
	case "expression_statement":
		expr := node.NamedChild(0)
		return IsVulnerableWithDataFlow(expr, sourceCode, dataFlowGraph, scope)

	default:
		return IsVulnerableWithDataFlow(node, sourceCode, dataFlowGraph, scope)
	}

	return false
}

func analyzeUserDefinedFunction(funcName string, args *sitter.Node, sourceCode []byte, dataFlowGraph map[*analysis.Variable]*DataFlowNode, scope *analysis.Scope) bool {
	funcDef, exists := functionDefinitions[funcName]
	if !exists {
		return false
	}

	// Create a map of parameter to argument taint status
	paramTaint := make(map[*analysis.Variable]bool)

	// Match arguments to parameters
	for i := 0; i < int(args.NamedChildCount()) && i < len(funcDef.Parameters); i++ {
		arg := args.NamedChild(i)
		param := funcDef.Parameters[i]

		// Check if argument is tainted
		if arg.Type() == "identifier" {
			argName := arg.Content(sourceCode)
			if variable := scope.Lookup(argName); variable != nil {
				if sourceNode, exists := dataFlowGraph[variable]; exists {
					paramTaint[param] = sourceNode.Tainted
				}
			}
		}
	}

	// Analyze function body with parameter taint information
	return analyzeNodeForVulnerabilities(funcDef.Body, sourceCode, dataFlowGraph, funcDef.Scope, paramTaint)
}

func analyzeNodeForVulnerabilities(node *sitter.Node, sourceCode []byte, dataFlowGraph map[*analysis.Variable]*DataFlowNode, scope *analysis.Scope, paramTaint map[*analysis.Variable]bool) bool {
	if node == nil {
		return false
	}

	switch node.Type() {
	case "call_expression":
		funcNode := node.ChildByFieldName("function")
		if funcNode != nil {
			funcName := funcNode.Content(sourceCode)
			if vulnerableFunctions[funcName] {
				// Check if any tainted parameter is used in the arguments
				args := node.ChildByFieldName("arguments")
				if args != nil {
					for i := 0; i < int(args.NamedChildCount()); i++ {
						arg := args.NamedChild(i)
						if arg.Type() == "identifier" {
							argName := arg.Content(sourceCode)
							if variable := scope.Lookup(argName); variable != nil {
								if tainted, exists := paramTaint[variable]; exists && tainted {
									return true
								}
							}
						}
					}
				}
			}
		}
	}

	// Recursively analyze children
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if analyzeNodeForVulnerabilities(child, sourceCode, dataFlowGraph, scope, paramTaint) {
			return true
		}
	}

	return false
}
