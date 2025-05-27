package python

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
	"strings"
	// "sort" // Uncomment if sorting of types in the message is desired for deterministic test output
)

var MultipleReturnTypes = &analysis.Analyzer{
	Name:        "multiple-return-types",
	Language:    analysis.LangPy,
	Description: "Detects functions with multiple return types.",
	Category:    analysis.CategoryBugRisk,
	Severity:    analysis.SeverityInfo,
	Run:         checkMultipleReturnTypes,
}

func findReturnStatementsRecursive(
	currentNode *sitter.Node,
	pass *analysis.Pass,
	returnTypes map[string]bool,
	hasExplicitReturn *bool, 
	source []byte,
) {
	if currentNode == nil {
		return
	}

	if currentNode.Type() == "function_definition" || currentNode.Type() == "class_definition" {
		return
	}

	if currentNode.Type() == "return_statement" {
		*hasExplicitReturn = true 
		returnValueNode := currentNode.NamedChild(0)
		var currentType string

		if returnValueNode == nil {
			currentType = "NoneType"
		} else {
			switch returnValueNode.Type() {
			case "string":
				currentType = "str"
			case "integer":
				currentType = "int"
			case "float":
				currentType = "float"
			case "true", "false":
				currentType = "bool"
			case "none":
				currentType = "NoneType"
			case "identifier":
				// currentType = "identifier:" + returnValueNode.Content(source) // Strict version
				currentType = "identifier" // Generalized version
			case "call":
				// callNameNode := returnValueNode.ChildByFieldName("function") // Needed for strict version
				// if callNameNode != nil {
				// 	currentType = "call_result:" + callNameNode.Content(source) // Strict version
				// } else {
				// 	currentType = "call_result:unknown" // Strict version
				// }
				currentType = "call_result" // Generalized version
			default:
				currentType = "complex_type"
			}
		}
		returnTypes[currentType] = true
		return 
	}

	for i := 0; i < int(currentNode.ChildCount()); i++ {
		child := currentNode.Child(i)
		findReturnStatementsRecursive(child, pass, returnTypes, hasExplicitReturn, source)
	}
}

func checkMultipleReturnTypes(pass *analysis.Pass) (interface{}, error) {
	source := pass.FileContext.Source 

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "function_definition" {
			return
		}

		functionNameNode := node.ChildByFieldName("name")
		functionName := "unknown_function"
		if functionNameNode != nil {
			functionName = functionNameNode.Content(source)
		}

		returnTypes := make(map[string]bool)
		hasExplicitReturnStatement := false

		functionBodyNode := node.ChildByFieldName("body")
		if functionBodyNode == nil {
			return
		}

		for i := 0; i < int(functionBodyNode.ChildCount()); i++ {
			childOfBody := functionBodyNode.Child(i)
			findReturnStatementsRecursive(childOfBody, pass, returnTypes, &hasExplicitReturnStatement, source)
		}
		
		// Simplified canFallOffEnd logic from previous step:
		// Only add NoneType from fall-off if NO explicit returns were found at all.
		canFallOffEnd := false
		if !hasExplicitReturnStatement { 
			if functionBodyNode.NamedChildCount() > 0 {
				lastStatement := functionBodyNode.NamedChild(int(functionBodyNode.NamedChildCount() - 1))
				if lastStatement != nil && lastStatement.Type() != "return_statement" && lastStatement.Type() != "raise_statement" {
					canFallOffEnd = true
				}
			} else { 
				canFallOffEnd = true
			}
		}
		
		if canFallOffEnd { 
			if !hasExplicitReturnStatement && len(returnTypes) == 0 { 
				returnTypes["NoneType"] = true
			}
		}

		if len(returnTypes) > 1 {
			var foundTypes []string
			for t := range returnTypes {
				foundTypes = append(foundTypes, t)
			}
			message := "Function '" + functionName + "' has multiple return types: " + strings.Join(foundTypes, ", ")
			pass.Report(pass, node, message)
		}
	})
	return nil, nil
}
