package python

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var AwsLambdaTaintedHtmlResponse *analysis.Analyzer = &analysis.Analyzer{
	Name: "aws-lambda-tainted-html-resp",
	Language: analysis.LangPy,
	Description: "Directly inserting user input into HTML can create cross-site scripting (XSS) vulnerabilities. This potentially allows attackers to steal sensitive user data. Always use secure HTML rendering methods instead of manual HTML construction.",
	Category: analysis.CategorySecurity,
	Severity: analysis.SeverityWarning,
	Run: checkAwsLambdaTaintedHtmlResponse,
}

func checkAwsLambdaTaintedHtmlResponse(pass *analysis.Pass) (interface{}, error) {
	htmlVarMap := make(map[string]bool)
	eventVarMap := make(map[string]bool)
	intermVarMap := make(map[string]bool)

	// first pass: check for variable storing html string
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		// f-strings will not be detected here
		if rightNode.Type() == "string" && rightNode.Content(pass.FileContext.Source)[0] == 'f' {
			return
		}

		if rightNode.Type() == "string" {
			htmlVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}

		if isEventSubscript(rightNode, pass.FileContext.Source) {
			eventVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// second pass: check for intermediate variables with string formatting
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isTaintedArg(rightNode, pass.FileContext.Source, htmlVarMap, eventVarMap, intermVarMap) {
			intermVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})
	
	// detect insecure html response in an assignment
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		// leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if rightNode.Type() != "dictionary" {
			return
		}

		allDictPairNodes := getNamedChildren(rightNode, 0)

		for _, pairNode := range allDictPairNodes {
			if isDangerousPair(pairNode, pass.FileContext.Source, htmlVarMap, eventVarMap, intermVarMap) {
				pass.Report(pass, pairNode, "Detected potential XSS vulnerability - avoid direct user input in HTML response")
			}
		}
	})

	// detected insecure html response in return statement
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "return_statement" {
			return
		}

		dictionaryNode := node.NamedChild(0)
		if dictionaryNode.Type() != "dictionary" {
			return
		}

		allDictPairNodes := getNamedChildren(dictionaryNode, 0)

		for _, pairNode := range allDictPairNodes {
			if isDangerousPair(pairNode, pass.FileContext.Source, htmlVarMap, eventVarMap, intermVarMap) {
				pass.Report(pass, pairNode, "Detected potential XSS vulnerability - avoid direct user input in HTML response")
			}
		}
	})

	return nil, nil
}

func isDangerousPair(node *sitter.Node, source []byte, htmlVarMap, eventVarMap, intermVarMap map[string]bool) bool {
	if node.Type() != "pair" {
		return false
	}
	keyNode := node.ChildByFieldName("key")
	if keyNode.Type() != "string" || trimQuotes(keyNode.Content(source)) != "body" {
		return false
	}

	valueNode := node.ChildByFieldName("value")
	if isEventSubscript(valueNode, source) {
		return true
	}

	return isEventSubscript(valueNode, source) || isTaintedArg(valueNode, source, htmlVarMap, eventVarMap, intermVarMap)

}

func isTaintedArg(arg *sitter.Node, source []byte, htmlVarMap, eventVarMap, intermVarMap map[string]bool) bool {
	switch arg.Type() {
	case "call":
		funcAttrNode := arg.ChildByFieldName("function")
		if funcAttrNode.Type() != "attribute" {
			return false
		}

		funcNameObjectNode := funcAttrNode.ChildByFieldName("object")
		if funcNameObjectNode.Type() != "identifier" && funcNameObjectNode.Type() != "string" {
			return false
		}

		ishtml := false
		for codevar := range htmlVarMap {
			if codevar == funcNameObjectNode.Content(source) {
				ishtml = true
			}
		}

		// this check only applies when the function object is not a string
		if !ishtml && funcNameObjectNode.Type() == "identifier" {
			return false
		}

		funcNameAttrVar := funcAttrNode.ChildByFieldName("attribute")
		if funcNameAttrVar.Type() != "identifier" && funcNameAttrVar.Content(source) != "format" {
			return false
		}

		argsListNode := arg.ChildByFieldName("arguments")
		if argsListNode.Type() != "argument_list" {
			return false
		}

		argsNode := getNamedChildren(argsListNode, 0)
		for _, callArg := range argsNode {
			if isEventSubscript(callArg, source) {
				return true
			}

			if callArg.Type() == "identifier" {
				for eventvar := range eventVarMap {
					if eventvar == callArg.Content(source) {
						return true
					}
				}
			}
		}

	case "binary_operator":
		leftNode := arg.ChildByFieldName("left")
		rightNode := arg.ChildByFieldName("right")

		if leftNode.Type() != "identifier" {
			return false
		}

		iscode := false
		for codevar := range htmlVarMap {
			if codevar == leftNode.Content(source) {
				iscode = true
			}
		}

		if !iscode {
			return false
		}

		if isEventSubscript(rightNode, source) {
			return true
		}

		return eventVarMap[rightNode.Content(source)]

	case "string":
		if arg.Content(source)[0] != 'f' {
			return false
		}
		stringChildrenNodes := getNamedChildren(arg, 0)
		for _, strchild := range stringChildrenNodes {
			if strchild.Type() == "interpolation" {
				exprNode := strchild.ChildByFieldName("expression")

				if exprNode.Type() == "subscript" {
					return isEventSubscript(exprNode, source)
				} else if exprNode.Type() == "identifier" {
					return intermVarMap[exprNode.Content(source)]
				} else {
					return false
				}
			}
		}

	case "identifier":
		return intermVarMap[arg.Content(source)] || eventVarMap[arg.Content(source)]

	}

	return false
}


func isEventSubscript(node *sitter.Node, source []byte) bool {
	if node.Type() != "subscript" {
		return false
	}

	valueNode := node.ChildByFieldName("value")
	if valueNode.Type() != "identifier" {
		return false
	}

	eventIdentifier := valueNode.Content(source)
	return strings.Contains(eventIdentifier, "event")
}