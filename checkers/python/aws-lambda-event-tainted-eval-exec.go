package python

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var AwsLambdaEventTaintedEvalExec *analysis.Analyzer = &analysis.Analyzer{
	Name:        "aws-lambda-event-tainted-eval-exec",
	Language:    analysis.LangPy,
	Description: "The use of `exec/eval` functions poses a critical security risk by executing dynamic code at runtime. This creates potential code injection vulnerabilities when processing external input, allowing attackers to run arbitrary code. Always ensure evaluated content comes from trusted sources and undergoes thorough sanitization, or use safer alternatives like parsing and validation.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkAwsLambdaEventTaintedEvalExec,
}

func checkAwsLambdaEventTaintedEvalExec(pass *analysis.Pass) (interface{}, error) {
	codeVarMap := make(map[string]bool)
	eventVarMap := make(map[string]bool)
	intermVarMap := make(map[string]bool)

	// first pass: get variables storing code & variables storing event subscripted data
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		// string might be code. we don't know
		if rightNode.Type() == "string" {
			codeVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}

		// check if variable contains data from event parameter
		if isEventSubscript(rightNode, pass.FileContext.Source) {
			eventVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// second pass: get intermediate variables storing code formatted with event subscript
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isTaintedArg(rightNode, pass.FileContext.Source, codeVarMap, eventVarMap, intermVarMap) {
			intermVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "identifier" {
			return
		}

		funcName := funcNode.Content(pass.FileContext.Source)
		if funcName != "exec" && funcName != "eval" {
			return
		}

		argsNode := node.ChildByFieldName("arguments")
		if argsNode.Type() != "argument_list" {
			return
		}

		codeArg := argsNode.NamedChild(0)

		if isTaintedArg(codeArg, pass.FileContext.Source, codeVarMap, eventVarMap, intermVarMap) {
			pass.Report(pass, node, "Detected AWS Lambda function using `exec/eval` with external input - potential code injection vulnerability")
		}

	})

	return nil, nil
}

func isTaintedArg(arg *sitter.Node, source []byte, codeVarMap, eventVarMap, intermVarMap map[string]bool) bool {
	switch arg.Type() {
	case "call":
		funcAttrNode := arg.ChildByFieldName("function")
		if funcAttrNode.Type() != "attribute" {
			return false
		}

		funcNameObjectNode := funcAttrNode.ChildByFieldName("object")
		if funcNameObjectNode.Type() != "identifier" {
			return false
		}

		iscode := false
		for codevar := range codeVarMap {
			if codevar == funcNameObjectNode.Content(source) {
				iscode = true
			}
		}

		if !iscode {
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
		for codevar := range codeVarMap {
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
		return intermVarMap[arg.Content(source)]
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
