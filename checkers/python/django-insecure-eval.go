package python

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DjangoInsecureEval *analysis.Analyzer = &analysis.Analyzer{
	Name:        "django-insecure-eval",
	Language:    analysis.LangPy,
	Description: "Using `eval` with user data creates a severe security vulnerability that allows attackers to execute arbitrary code on your system. This dangerous practice can lead to complete system compromise, data theft, or service disruption. Instead, replace `eval` with dedicated libraries or methods specifically designed for your required functionality.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkDjangoInsecureEval,
}

func checkDjangoInsecureEval(pass *analysis.Pass) (interface{}, error) {
	requestVarMap := make(map[string]bool)

	// check for assignment of `request` data stored in variables
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode.Type() != "call" && rightNode.Type() != "subscript" {
			return
		}

		if isRequestCall(rightNode, pass.FileContext.Source) {
			varName := leftNode.Content(pass.FileContext.Source)
			requestVarMap[varName] = true
		}

	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if !strings.Contains(funcNode.Content(pass.FileContext.Source), "eval") {
			return
		}

		argNode := node.ChildByFieldName("arguments")
		args := argNode.Content(pass.FileContext.Source)
		argumentList := getNamedChildren(argNode, 0)

		for _, arg := range argumentList {
			if arg.Type() == "identifier" {
				for key := range requestVarMap {
					if strings.Contains(args, key) {
						pass.Report(pass, node, "Detected user data in `eval` call which can cause remote code execution")
					}
				}
			} else if isRequestCall(arg, pass.FileContext.Source) {
				pass.Report(pass, node, "Detected user data in `eval` call which can cause remote code execution")
			}
		}

	})
	return nil, nil
}

func isRequestCall(node *sitter.Node, source []byte) bool {
	switch node.Type() {
	case "call":
		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return false
		}
		objectNode := funcNode.ChildByFieldName("object")
		if !strings.Contains(objectNode.Content(source), "request") {
			return false
		}

		attributeNode := funcNode.ChildByFieldName("attribute")
		if attributeNode.Type() != "identifier" {
			return false
		}

		if !strings.Contains(attributeNode.Content(source), "get") {
			return false
		}

		return true

	case "subscript":
		valueNode := node.ChildByFieldName("value")
		if valueNode.Type() != "attribute" {
			return false
		}

		objNode := valueNode.ChildByFieldName("object")
		if objNode.Type() != "identifier" && objNode.Content(source) != "request" {
			return false
		}

		return true
	}

	return false
}
