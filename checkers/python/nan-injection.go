package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var NanInjection *analysis.Analyzer = &analysis.Analyzer{
	Name:        "nan-injection",
	Language:    analysis.LangPy,
	Description: "User input is directly cast to `bool()`, `float()`, or `complex()`, allowing an attacker to inject Python's `NaN`. This can lead to undefined behavior, especially in comparisons. To mitigate this, either use a different type for casting or explicitly check for all capitalizations of 'nan' before conversion.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkNanInjection,
}

func checkNanInjection(pass *analysis.Pass) (interface{}, error) {
	reqVarMap := make(map[string]bool)

	// get the variable name from the Flask decorated route function
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "decorated_definition" {
			return
		}

		decNode := node.NamedChild(0)
		if decNode.Type() != "decorator" {
			return
		}
		callNode := decNode.NamedChild(0)
		if callNode.Type() != "call" {
			return
		}
		funcNode := callNode.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}
		if !strings.HasSuffix(funcNode.Content(pass.FileContext.Source), ".route") {
			return
		}

		defNode := node.ChildByFieldName("definition")
		if defNode.Type() != "function_definition" {
			return
		}

		paramsNode := defNode.ChildByFieldName("parameters")
		if paramsNode.NamedChildCount() == 0 {
			return
		}
		if paramsNode.Type() != "parameters" {
			return
		}
		allparamNodes := getNamedChildren(paramsNode, 0)
		for _, p := range allparamNodes {
			reqVarMap[p.Content(pass.FileContext.Source)] = true
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		rightNode := node.ChildByFieldName("right")
		leftNode := node.ChildByFieldName("left")

		if rightNode == nil {
			return
		}

		if isRequestCall(rightNode, pass.FileContext.Source) {
			reqVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}

	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "function_definition" {
			return
		}

		funcBody := node.ChildByFieldName("body")
		if funcBody.Type() != "block" {
			return
		}

		ifConditionPresent := false

		funcBodyNodes := getNamedChildren(funcBody, 0)

		for _, bodyNode := range funcBodyNodes {
			if bodyNode.Type() == "if_statement" {
				conditionNode := bodyNode.ChildByFieldName("condition")
				for req := range reqVarMap {
					re := regexp.MustCompile(`\b` + req + `\b`)
					if re.MatchString(conditionNode.Content(pass.FileContext.Source)) {
						ifConditionPresent = true
					}
				}
			} else if bodyNode.Type() == "expression_statement" {
				exprNode := bodyNode.NamedChild(0)
				if isInsecureTypeCast(exprNode, pass.FileContext.Source, reqVarMap) && !ifConditionPresent {
					pass.Report(pass, bodyNode, "Directly type-casting user data could cause potential NaN injection vulnerabilities that can cause undefined behavior")
				}
			}
		}
	})

	return nil, nil
}

func isInsecureTypeCast(node *sitter.Node, source []byte, reqvarMap map[string]bool) bool {
	switch node.Type() {
	case "assignment":
		rightNode := node.ChildByFieldName("right")
		if rightNode == nil {
			return false
		}

		if rightNode.Type() != "call" {
			return false
		}

		funcNode := rightNode.ChildByFieldName("function")
		if funcNode.Type() != "identifier" {
			return false
		}

		funcName := funcNode.Content(source)
		if funcName != "float" && funcName != "complex" && funcName != "bool" {
			return false
		}

		argListNode := rightNode.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return false
		}

		argNode := argListNode.NamedChild(0)
		if isRequestCall(argNode, source) {
			return true
		} else if argNode.Type() == "identifier" {
			return reqvarMap[argNode.Content(source)]
		}

	case "call":
		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "identifier" {
			return false
		}

		funcName := funcNode.Content(source)
		if funcName != "float" && funcName != "complex" && funcName != "bool" {
			return false
		}

		argListNode := node.ChildByFieldName("arguments")
		if argListNode.Type() != "argument_list" {
			return false
		}

		argNode := argListNode.NamedChild(0)
		if isRequestCall(argNode, source) {
			return true
		} else if argNode.Type() == "identifier" {
			return reqvarMap[argNode.Content(source)]
		}
	}

	return false
}
