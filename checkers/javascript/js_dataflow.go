package javascript

import (
	"fmt"
	"reflect"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DataFlowAnalyzer = &analysis.Analyzer{
	Name:        "DataFlowAnalyzer",
	Language:    analysis.LangJs,
	Description: "Create a Data Flow Graph for a javascript file",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Run:         createDataFlowGraph,
	ResultType:  reflect.TypeOf(&DataFlowGraph{}),
	Requires:    []*analysis.Analyzer{ScopeAnalyzer},
}

// DataFlowNode represents a node in our data flow graph
type DataFlowNode struct {
	Node     *sitter.Node
	Sources  []*DataFlowNode
	Scope    *analysis.Scope
	Variable *analysis.Variable // Tracking variable instead of the varName helps in handling scope
	FuncDef  *FunctionDefinition
}

type FunctionDefinition struct {
	Node       *sitter.Node
	Parameters []*analysis.Variable
	Body       *sitter.Node
	Scope      *analysis.Scope
}

type DataFlowGraph struct {
	Graph     map[*analysis.Variable]*DataFlowNode
	ScopeTree *analysis.ScopeTree
}

var functionDefinitions = make(map[string]*FunctionDefinition)

// var DataFlowGraph = make(map[*analysis.Variable]*DataFlowNode)

func createDataFlowGraph(pass *analysis.Pass) (interface{}, error) {

	scopeResult, err := buildScopeTree(pass)
	if err != nil {
		return nil, fmt.Errorf("failed to build the scope tree \n")
	}

	scopeTree := scopeResult.(*analysis.ScopeTree)

	// Map to track variable definitions and their data flow nodes
	dataFlowGraph := &DataFlowGraph{
		Graph:     make(map[*analysis.Variable]*DataFlowNode),
		ScopeTree: scopeTree,
	}

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
			}

			if nameNode != nil && nameNode.Type() == "identifier" && valueNode != nil {
				varName := nameNode.Content(pass.FileContext.Source)

				variable := currentScope.Lookup(varName)
				// Create data flow node
				dfNode := &DataFlowNode{
					Node:     valueNode,
					Sources:  []*DataFlowNode{},
					Scope:    currentScope,
					Variable: variable,
				}

				switch valueNode.Type() {
				case "identifier":
					// If value is another variable, link to its data flow node
					sourceVarName := valueNode.Content(pass.FileContext.Source)
					currVar := currentScope.Lookup(sourceVarName)
					if sourceNode, exists := dataFlowGraph.Graph[currVar]; exists {
						dfNode.Sources = append(dfNode.Sources, sourceNode)
					}

				case "call_expression":
					// Track data flow through function arguments
					handleCallExpressionDataFlow(valueNode, dfNode, dataFlowGraph.Graph, pass.FileContext.Source, currentScope)

				case "binary_expression":
					// Handle string concatenation data flow
					handleBinaryExpressionDataFlow(valueNode, dfNode, dataFlowGraph.Graph, pass.FileContext.Source, currentScope)

				case "template_string":
					// Handle template literals data flow
					handleTemplateStringDataFlow(valueNode, dfNode, dataFlowGraph.Graph, pass.FileContext.Source, currentScope)

				case "function_expression":
					funcScope := scopeTree.GetScope(valueNode)
					funcBody := valueNode.ChildByFieldName("body")
					if funcBody == nil {
						return
					}
					funcDef := &FunctionDefinition{
						Node:  valueNode,
						Body:  funcBody,
						Scope: funcScope,
					}

					for _, param := range funcScope.Variables {
						funcDef.Parameters = append(funcDef.Parameters, param)
					}

					// Since this anon func doesn't have any name, it'll be related to the variable that is declaring it
					functionDefinitions[varName] = funcDef
					dfNode.FuncDef = funcDef

				case "arrow_function":
					funcScope := scopeTree.GetScope(valueNode)

					funcBody := valueNode.ChildByFieldName("body")
					if funcBody == nil {
						return
					}
					funcDef := &FunctionDefinition{
						Node:  valueNode,
						Body:  funcBody,
						Scope: funcScope,
					}

					for _, param := range funcScope.Variables {
						funcDef.Parameters = append(funcDef.Parameters, param)
					}

					functionDefinitions[varName] = funcDef
					dfNode.FuncDef = funcDef

				}
				dataFlowGraph.Graph[variable] = dfNode

			}
		}

		// Handle function declarations like `function foo(){...}`
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

			for _, param := range currentScope.Variables {
				funcDef.Parameters = append(funcDef.Parameters, param)
			}

			functionDefinitions[funcName] = funcDef
		}

		// Handle IIFE `(function(){...})()`
		if node.Type() == "call_expression" {
			immidiateFunc := node.ChildByFieldName("function")
			// Used to verify that the call_expression is actually pointing to an IIFE(immidiately invoked function expression)
			// also filters out false positives of regular call expressions like console.log(), foo(x) etc.
			if immidiateFunc == nil || immidiateFunc.Type() != "parenthesized_expression" {
				return
			}

			funcExpr := immidiateFunc.NamedChild(0)
			if funcExpr == nil {
				return
			}

			funcDef := &FunctionDefinition{
				Node:  funcExpr,
				Body:  funcExpr.ChildByFieldName("body"),
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

			// TODO: Figure out a way to store anon functions in the function definition map.

		}

		if node.Type() == "class_declaration" {
			var dfNode *DataFlowNode
			className := node.ChildByFieldName("name")
			if className != nil {
				varClassName := className.Content(pass.FileContext.Source)
				classNameVar := currentScope.Lookup(varClassName)
				classScope := scopeTree.GetScope(classNameVar.DeclNode)
				if classScope == nil {
					return
				}
				dfNode = &DataFlowNode{
					Node:     classNameVar.DeclNode,
					Scope:    classScope,
					Variable: classNameVar,
				}
			}
			dataFlowGraph.Graph[dfNode.Variable] = dfNode
		}

	})

	return dataFlowGraph, nil
}

// Handle data flow through binary expressions (like string concatenation)
func handleBinaryExpressionDataFlow(node *sitter.Node, dfNode *DataFlowNode, DataFlowGraph map[*analysis.Variable]*DataFlowNode, sourceCode []byte, scope *analysis.Scope) {
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
				if sourceNode, exists := DataFlowGraph[variable]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)

				}
			}

		}

		// Check right side
		if right != nil && right.Type() == "identifier" {
			rightVarName := right.Content(sourceCode)
			if variable := scope.Lookup(rightVarName); variable != nil {
				if sourceNode, exists := DataFlowGraph[variable]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)

				}
			}

		}

		// Recursively process nested expressions
		if left != nil && left.Type() == "binary_expression" {
			handleBinaryExpressionDataFlow(left, dfNode, DataFlowGraph, sourceCode, scope)
		}

		if right != nil && right.Type() == "binary_expression" {
			handleBinaryExpressionDataFlow(right, dfNode, DataFlowGraph, sourceCode, scope)
		}
	}
}

// Handle data flow through template strings
func handleTemplateStringDataFlow(node *sitter.Node, dfNode *DataFlowNode, DataFlowGraph map[*analysis.Variable]*DataFlowNode, sourceCode []byte, scope *analysis.Scope) {
	if node == nil || node.Type() != "template_string" {
		return
	}

	// Look for template substitutions
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child != nil && child.Type() == "template_substitution" {
			// Get the ression inside ${}
			exprNode := child.NamedChild(0)
			if exprNode != nil && exprNode.Type() == "identifier" {
				varName := exprNode.Content(sourceCode)
				if variable := scope.Lookup(varName); variable != nil {
					if sourceNode, exists := DataFlowGraph[variable]; exists {
						dfNode.Sources = append(dfNode.Sources, sourceNode)

					}
				}
			}
		}
	}
}

// Handle data flow through function arguments
func handleCallExpressionDataFlow(node *sitter.Node, dfNode *DataFlowNode, DataFlowGraph map[*analysis.Variable]*DataFlowNode, sourceCode []byte, scope *analysis.Scope) {
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
				if sourceNode, exists := DataFlowGraph[variable]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)
				}
			}

		}
	}
}
