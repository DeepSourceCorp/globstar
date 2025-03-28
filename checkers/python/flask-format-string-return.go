package python

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var FlaskFormatStringReturn *analysis.Analyzer = &analysis.Analyzer{
	Name:        "flask-format-string-return",
	Language:    analysis.LangPy,
	Description: "Returning formatted strings directly from Flask routes creates cross-site scripting vulnerabilities when user input is incorporated without proper escaping. Attackers can inject malicious JavaScript that executes in users' browsers. Flask's template engine with `render_template()` automatically handles proper escaping to prevent these attacks.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkFlaskFormatStringReturn,
}

func checkFlaskFormatStringReturn(pass *analysis.Pass) (interface{}, error) {
	taintDataMap := make(map[string]bool)
	intermVarMap := make(map[string]bool)

	// tainted variable from decorated function
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "decorated_definition" {
			return
		}

		decoratorNode := node.NamedChild(0)
		if decoratorNode.Type() != "decorator" {
			return
		}

		callNode := decoratorNode.NamedChild(0)
		if callNode.Type() != "call" {
			return
		}

		funcNode := callNode.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}

		funcAttr := funcNode.ChildByFieldName("attribute")
		if funcAttr.Type() != "identifier" && funcAttr.Content(pass.FileContext.Source) != "route" {
			return
		}

		defNode := node.ChildByFieldName("definition")
		if defNode.Type() != "function_definition" {
			return
		}

		paramNode := defNode.ChildByFieldName("parameters")
		if paramNode.Type() != "parameters" {
			return
		}

		params := getNamedChildren(paramNode, 0)

		for _, p := range params {
			taintDataMap[p.Content(pass.FileContext.Source)] = true
		}
	})

	// variable tainted by request
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isRequestCall(rightNode, pass.FileContext.Source) {
			taintDataMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// variable of data formatted by tainted variable
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isTaintFormatted(rightNode, pass.FileContext.Source, intermVarMap, taintDataMap) {
			intermVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// detection
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "return_statement" {
			return
		}
		returnValNode := node.NamedChild(0)
		if isTaintFormatted(returnValNode, pass.FileContext.Source, intermVarMap, taintDataMap) {
			pass.Report(pass, node, "Flask route returns formatted string allowing XSS - use render_template() instead")
		}
	})

	return nil, nil
}

func isTaintFormatted(node *sitter.Node, source []byte, intermVarMap, taintDataMap map[string]bool) bool {
	switch node.Type() {
	case "call":
		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return false
		}
		if !strings.HasSuffix(funcNode.Content(source), ".format") {
			return false
		}

		funcArgsListNode := node.ChildByFieldName("arguments")
		if funcArgsListNode.Type() != "argument_list" {
			return false
		}

		argsNode := getNamedChildren(funcArgsListNode, 0)
		for _, arg := range argsNode {
			if isRequestCall(arg, source) || taintDataMap[arg.Content(source)] || intermVarMap[arg.Content(source)] {
				return true
			}
		}

	case "binary_operator":
		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if leftNode.Type() != "string" {
			return false
		}

		if rightNode.Type() == "identifier" {
			return isRequestCall(rightNode, source) || taintDataMap[rightNode.Content(source)] || intermVarMap[rightNode.Content(source)]
		} else if rightNode.Type() == "tuple" {
			tupleArgNodes := getNamedChildren(rightNode, 0)
			for _, targ := range tupleArgNodes {
				if isRequestCall(targ, source) || taintDataMap[targ.Content(source)] || intermVarMap[targ.Content(source)] {
					return true
				}
			}
		}

	case "string":
		stringChildNodes := getNamedChildren(node, 0)
		for _, child := range stringChildNodes {
			if child.Type() == "interpolation" {
				exprNode := child.NamedChild(0)
				if isRequestCall(exprNode, source) || taintDataMap[exprNode.Content(source)] || intermVarMap[exprNode.Content(source)] {
					return true
				}
			}
		}

	case "identifier":
		return intermVarMap[node.Content(source)] || taintDataMap[node.Content(source)]
	}

	return false
}
