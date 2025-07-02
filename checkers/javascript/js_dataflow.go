//globstar:registry-exclude

package javascript

import (
	"fmt"
	"reflect"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DataFlowAnalyzer = &analysis.Analyzer{
	Name:        "data_flow_analyzer",
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

type FunctionCall struct {
	Node    *sitter.Node
	Sources []*DataFlowNode
	DfNode  *DataFlowNode
}

type ClassDefinition struct {
	Node       *sitter.Node
	Properties []*analysis.Variable
	Methods    []*FunctionDefinition
	Scope      *analysis.Scope
}

type DataFlowGraph struct {
	Graph     map[*analysis.Variable]*DataFlowNode
	ScopeTree *analysis.ScopeTree
	FuncDefs  map[string]*FunctionDefinition
	ClassDefs map[*analysis.Variable]*ClassDefinition
	FuncCalls map[*sitter.Node]*FunctionCall
}

var functionDefinitions = make(map[string]*FunctionDefinition)
var classDefinitions = make(map[*analysis.Variable]*ClassDefinition)
var functionCalls = make(map[*sitter.Node]*FunctionCall)

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
		FuncDefs:  make(map[string]*FunctionDefinition),
	}

	// First pass: build initial data flow graph
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}

		currentScope := scopeTree.GetScope(node)
		// fmt.Println(currentScope.Variables)
		if currentScope == nil {
			return
		}

		// Track variable declarations and assignments
		// TODO: Add sources for the nameNode.
		if node.Type() == "variable_declarator" || node.Type() == "assignment_expression" {
			var nameNode, valueNode *sitter.Node

			if node.Type() == "variable_declarator" {
				nameNode = node.ChildByFieldName("name")
				valueNode = node.ChildByFieldName("value")
			} else { // assignment_expression
				nameNode = node.ChildByFieldName("left")
				valueNode = node.ChildByFieldName("right")
			}

			// fmt.Println(node)

			if nameNode != nil && nameNode.Type() == "identifier" && valueNode != nil {
				var dfNode *DataFlowNode
				varName := nameNode.Content(pass.FileContext.Source)
				variable := currentScope.Lookup(varName)
				// fmt.Println(variable)
				if variable != nil {
					dfNode = &DataFlowNode{
						Node:     nameNode,
						Sources:  []*DataFlowNode{},
						Scope:    currentScope,
						Variable: variable,
					}
				}
				// Create data flow node

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

			funcVar := currentScope.Lookup(funcName)
			if funcVar == nil {
				return
			}

			for _, param := range currentScope.Variables {
				funcDef.Parameters = append(funcDef.Parameters, param)
			}
			functionDefinitions[funcName] = funcDef
			dataFlowGraph.Graph[funcVar] = &DataFlowNode{
				Node:     funcNameNode,
				Sources:  []*DataFlowNode{},
				Scope:    currentScope,
				Variable: funcVar,
				FuncDef:  funcDef,
			}
		}

		// Handle IIFE `(function(){...})()`
		if node.Type() == "call_expression" {
			immidiateFunc := node.ChildByFieldName("function")
			// Used to verify that the call_expression is actually pointing to an IIFE(immidiately invoked function expression)
			// also filters out false positives of regular call expressions like console.log(), foo(x) etc.
			if immidiateFunc != nil && immidiateFunc.Type() == "parenthesized_expression" {
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
				// Create a data flow node for the IIFE
			}
			if immidiateFunc != nil && immidiateFunc.Type() == "identifier" {

				funcname := immidiateFunc.Content(pass.FileContext.Source)

				_, exists := functionDefinitions[funcname]
				if !exists {
					funcVar := currentScope.Lookup(funcname)
					if funcVar == nil {
						return
					}
					dfNode := &DataFlowNode{
						Node:     immidiateFunc,
						Sources:  []*DataFlowNode{},
						Scope:    currentScope,
						Variable: funcVar,
					}

					handleCallExpressionDataFlow(node, dfNode, dataFlowGraph.Graph, pass.FileContext.Source, currentScope)

					functionCalls[node] = &FunctionCall{
						Node:    immidiateFunc,
						Sources: dfNode.Sources,
						DfNode:  dfNode,
					}

				}
			}

			// Need a way to track function calls like setTimeout etc.
			// Can do something like this:
			// check if the function exits in the function definitions map. If it doesn't exist, then it must be a function call specific to the language.
			// Then handle it accordingly, for sources etc.
		}

		if node.Type() == "class_declaration" {
			var dfNode *DataFlowNode
			className := node.ChildByFieldName("name")
			if className == nil {
				return
			}

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
			dataFlowGraph.Graph[dfNode.Variable] = dfNode

			classBody := node.ChildByFieldName("body")
			if classBody == nil {
				return
			}

			var classMethods []*FunctionDefinition
			var classProperties []*analysis.Variable

			// Iterate over class body children
			for i := range int(classBody.NamedChildCount()) {
				classChild := classBody.NamedChild(i)
				if classChild == nil {
					continue
				}

				// Check for method definitions
				if classChild.Type() == "method_definition" {
					methodNameNode := classChild.ChildByFieldName("name")
					if methodNameNode != nil && methodNameNode.Type() == "property_identifier" {
						// methodName := methodNameNode.Content(pass.FileContext.Source)
						methodDef := &FunctionDefinition{
							Node:       classChild,
							Body:       classChild.ChildByFieldName("body"),
							Parameters: []*analysis.Variable{},
							Scope:      classScope,
						}

						params := node.ChildByFieldName("parameters")
						if params != nil {
							for i := 0; i < int(params.NamedChildCount()); i++ {
								param := params.NamedChild(i)
								if param.Type() == "identifier" {
									paramName := param.Content(pass.FileContext.Source)
									paramVar := currentScope.Lookup(paramName)
									if paramVar != nil {
										methodDef.Parameters = append(methodDef.Parameters, paramVar)
									}

								}
							}
						}
						classMethods = append(classMethods, methodDef)
					}
				}

				// Check for class properties
				if classChild.Type() == "public_field_definition" {
					// fmt.Println(propNameNode.Content(pass.FileContext.Source))
					propNameNode := classChild.ChildByFieldName("name")
					if propNameNode != nil && propNameNode.Type() == "property_identifier" {
						propName := propNameNode.Content(pass.FileContext.Source)
						propVar := classScope.Children[0].Lookup(propName)
						if propVar != nil {
							classProperties = append(classProperties, propVar)
						}
					}

				}
			}

			classDef := &ClassDefinition{
				Node:       node,
				Properties: classProperties,
				Methods:    classMethods,
				Scope:      classScope,
			}

			classDefinitions[classNameVar] = classDef
			// Add logic to collect properties and methods inside this block itself.
			// Iterate over name children -> Find method definitions -> Add them to an instance of function definition
			// Iterate over class properties -> Add them to an instance of class definition
			// Add the class definition to the data flow graph
		}

	})
	dataFlowGraph.FuncDefs = functionDefinitions
	dataFlowGraph.ClassDefs = classDefinitions
	dataFlowGraph.FuncCalls = functionCalls

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

		// Add handling of template strings inside a function call
		if arg.Type() == "template_string" {
			// fmt.Println(arg.Content(sourceCode))
			handleTemplateStringDataFlow(arg, dfNode, DataFlowGraph, sourceCode, scope)
		}

		if arg.Type() == "binary_expression" {
			handleBinaryExpressionDataFlow(arg, dfNode, DataFlowGraph, sourceCode, scope)
		}
	}

}
