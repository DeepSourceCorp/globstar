//globstar:registry-exclude

package python

import (
	"fmt"
	"reflect"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DataFlowAnalyzer = &analysis.Analyzer{
	Name:        "py-dataflow-analyzer",
	Language:    analysis.LangPy,
	Description: "Create a data flow graph for Python",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         createPyDFG,
	ResultType:  reflect.TypeOf(&DataFlowGraph{}),
	Requires:    []*analysis.Analyzer{ScopeAnalyzer},
}

type DataFlowNode struct {
	Node     *sitter.Node
	Sources  []*DataFlowNode
	Scope    *analysis.Scope
	Variable *analysis.Variable
	FuncDef  *FunctionDefinition
}

type FunctionDefinition struct {
	Node       *sitter.Node
	Parameters []*analysis.Variable
	Body       *sitter.Node
	Scope      *analysis.Scope
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
	FunDefs   map[string]*FunctionDefinition
	ClassDefs map[*analysis.Variable]*ClassDefinition
}

var functionDefinitions = make(map[string]*FunctionDefinition)
var classDefinitions = make(map[*analysis.Variable]*ClassDefinition)

func createPyDFG(pass *analysis.Pass) (interface{}, error) {
	scopeResult, err := buildScopeTree(pass)
	if err != nil {
		return nil, fmt.Errorf("failed to build the source tree")
	}

	scopeTree := scopeResult.(*analysis.ScopeTree)

	dfg := &DataFlowGraph{
		Graph:     make(map[*analysis.Variable]*DataFlowNode),
		ScopeTree: scopeTree,
		FunDefs:   make(map[string]*FunctionDefinition),
	}

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}

		currentScope := scopeTree.GetScope(node)
		if currentScope == nil {
			return
		}

		// track variable declarations and assignments
		if node.Type() == "assignment" {
			var nameNode, valueNode *sitter.Node

			nameNode = node.ChildByFieldName("left")
			valueNode = node.ChildByFieldName("right")

			if nameNode != nil && nameNode.Type() == "identifier" && valueNode != nil {
				var dfNode *DataFlowNode
				varName := nameNode.Content(pass.FileContext.Source)
				variable := currentScope.Lookup(varName)

				if variable == nil {
					dfNode = &DataFlowNode{
						Node:     nameNode,
						Sources:  []*DataFlowNode{},
						Scope:    currentScope,
						Variable: variable,
					}
				}

				switch valueNode.Type() {
				case "identifier":
					// if value is another variable, link to its data flow node
					sourceVarName := valueNode.Content(pass.FileContext.Source)
					currVar := currentScope.Lookup(sourceVarName)
					if sourceNode, exists := dfg.Graph[currVar]; exists {
						dfNode.Sources = append(dfNode.Sources, sourceNode)
					}

				case "call":
					handleFunctionCallDataFlow(valueNode, dfNode, dfg.Graph, pass.FileContext.Source, currentScope)

				case "binary_operator":
					handleBinaryExprDataFlow(valueNode, dfNode, dfg.Graph, pass.FileContext.Source, currentScope)

				// analyze the variables in an f-string
				case "string":
					if valueNode.Content(pass.FileContext.Source)[0] == 'f' {
						handleFStringDataFlow(valueNode, dfNode, dfg.Graph, pass.FileContext.Source, currentScope)
					}

				// lambda expressions are also functions
				case "lambda":
					lambdaScope := scopeTree.GetScope(valueNode)
					lambdaBody := valueNode.ChildByFieldName("body")
					if lambdaBody == nil {
						return
					}

					funcDef := &FunctionDefinition{
						Node:  valueNode,
						Body:  lambdaBody,
						Scope: lambdaScope,
					}

					for _, param := range lambdaScope.Variables {
						funcDef.Parameters = append(funcDef.Parameters, param)
					}

					functionDefinitions[varName] = funcDef
					dfNode.FuncDef = funcDef
				}
				dfg.Graph[variable] = dfNode

			}
		}

		if node.Type() == "function_definition" {
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
			dfg.Graph[funcVar] = &DataFlowNode{
				Node:     funcNameNode,
				Sources:  []*DataFlowNode{},
				Scope:    currentScope,
				Variable: funcVar,
				FuncDef:  funcDef,
			}
		}

		if node.Type() == "class_definition" {
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

			classBody := node.ChildByFieldName("body")
			if classBody == nil {
				return
			}

			var classMethods []*FunctionDefinition
			var classProperties []*analysis.Variable

			dfNode = &DataFlowNode{
				Node: classNameVar.DeclNode,
				Scope: classScope,
				Variable: classNameVar,
			}

			dfg.Graph[dfNode.Variable] = dfNode

			for i := range int(classBody.NamedChildCount()) {
				classChild := classBody.NamedChild(i)
				if classChild == nil {
					return
				}

				if classChild.Type() == "function_definition" {
					classMethodNameNode := classChild.ChildByFieldName("name")
					if classMethodNameNode != nil && classMethodNameNode.Type() == "identifier" {
						methodDef := &FunctionDefinition{
							Node: classChild,
							Body: classChild.ChildByFieldName("body"),
							Parameters: []*analysis.Variable{},
							Scope: classScope,
						}

						params := node.ChildByFieldName("parameters")
						if params != nil {
							for i := range int(params.NamedChildCount()) {
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
				} else if classChild.Type() == "assignment" {
					classVarNameNode := classChild.ChildByFieldName("left")
					if classVarNameNode != nil && classVarNameNode.Type() == "identifier" {
						classVarName := classVarNameNode.Content(pass.FileContext.Source)
						classVar := classScope.Children[0].Lookup(classVarName)
						if classVar != nil {
							classProperties = append(classProperties, classVar)
						}
					}
				}
			}

			classDef := &ClassDefinition{
				Node: node,
				Properties: classProperties,
				Methods: classMethods,
				Scope: classScope,
			}
			
			classDefinitions[classNameVar] = classDef
		}
	})

	dfg.FunDefs = functionDefinitions
	dfg.ClassDefs = classDefinitions

	return dfg, nil
}

func handleFStringDataFlow(node *sitter.Node, dfNode *DataFlowNode, dfg map[*analysis.Variable]*DataFlowNode, source []byte, scope *analysis.Scope) {
	if node == nil || node.Type() != "string" {
		return
	}

	interpolations := analysis.ChildrenWithFieldName(node, "interpolation")
	for _, interpNode := range interpolations {
		exprNode := interpNode.ChildByFieldName("expression")
		if exprNode != nil && exprNode.Type() == "identifier" {
			varName := exprNode.Content(source)
			if variable := scope.Lookup(varName); variable != nil {
				if sourceNode, exists := dfg[variable]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)
				}
			}
		}
	}
}

func handleBinaryExprDataFlow(node *sitter.Node, dfNode *DataFlowNode, dfg map[*analysis.Variable]*DataFlowNode, source []byte, scope *analysis.Scope) {
	if node == nil || node.Type() != "binary_operator" {
		return
	}

	left := node.ChildByFieldName("left")
	right := node.ChildByFieldName("right")

	if left != nil && left.Type() == "identifier" {
		leftVar := left.Content(source)
		if variable := scope.Lookup(leftVar); variable != nil {
			if sourceNode, exists := dfg[variable]; exists {
				dfNode.Sources = append(dfNode.Sources, sourceNode)
			}
		}
	}

	if right != nil && right.Type() == "identifier" {
		rightVar := right.Content(source)
		if variable := scope.Lookup(rightVar); variable != nil {
			if sourceNode, exists := dfg[variable]; exists {
				dfNode.Sources = append(dfNode.Sources, sourceNode)
			}
		}
	}

	// process nested binary expression
	if left != nil && left.Type() == "binary_operator" {
		handleBinaryExprDataFlow(left, dfNode, dfg, source, scope)
	}

	if right != nil && right.Type() == "binary_operator" {
		handleBinaryExprDataFlow(right, dfNode, dfg, source, scope)
	}
}

func handleFunctionCallDataFlow(node *sitter.Node, dfNode *DataFlowNode, dfg map[*analysis.Variable]*DataFlowNode, source []byte, scope *analysis.Scope) {
	if node == nil || node.Type() != "call" {
		return
	}

	args := node.ChildByFieldName("arguments")
	if args == nil || args.Type() != "argument_list" {
		return
	}

	for i := range int(args.NamedChildCount()) {
		arg := args.NamedChild(i)
		if arg == nil {
			continue
		}

		if arg.Type() == "identifier" {
			argName := arg.Content(source)
			if variable := scope.Lookup(argName); variable != nil {
				if sourceNode, exists := dfg[variable]; exists {
					dfNode.Sources = append(dfNode.Sources, sourceNode)
				}
			}
		}
	}
}
