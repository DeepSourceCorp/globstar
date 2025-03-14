package python

import (
	"slices"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var AwsLambdaDangerousOsSystemCall *analysis.Analyzer = &analysis.Analyzer{
	Name:        "aws-lambda-dangerous-os-system-call",
	Language:    analysis.LangPy,
	Description: "Detected a potentially dangerous use of the `os` function with an argument influenced by the `event` object. This could allow command injection if external data reaches this call. Use the `subprocess` module instead to prevent security risks.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkAwsLambdaDangerousOsSystemCall,
}

func checkAwsLambdaDangerousOsSystemCall(pass *analysis.Pass) (interface{}, error) {
	systemCall := []string{"system", "popen", "popen2", "popen3", "popen4"}
	eventVarMap := make(map[string]bool)  // var for `event['url']` etc.
	cmdVarMap := make(map[string]bool)    // var for command strings
	intermVarMap := make(map[string]bool) // when intermediate variable is used for formatting the command string

	// first pass: track the variable names for commands and event subscripts (if called)
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		// f-strings will be added to intermVarMap
		if rightNode.Type() == "string" && rightNode.Content(pass.FileContext.Source)[0] != 'f' {
			cmdVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}

		if isEventSubscript(rightNode, pass.FileContext.Source) {
			eventVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// second pass: track the variable names of intermediate command string formatting
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isTaintedCmdString(rightNode, pass.FileContext.Source, eventVarMap, cmdVarMap, intermVarMap) {
			intermVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}

		funcObjNode := funcNode.ChildByFieldName("object")
		funcAttrNode := funcNode.ChildByFieldName("attribute")

		if funcAttrNode.Type() != "identifier" || funcObjNode.Type() != "identifier" {
			return
		}

		if funcObjNode.Content(pass.FileContext.Source) != "os" {
			return
		}

		if !slices.Contains(systemCall, funcAttrNode.Content(pass.FileContext.Source)) {
			return
		}
		
		argsListNode := node.ChildByFieldName("arguments")
		if argsListNode.Type() != "argument_list" {
			return
		}

		systemCallArgListNode := getNamedChildren(argsListNode, 0)

		for _, argNode := range systemCallArgListNode {
			if isTaintedCmdString(argNode, pass.FileContext.Source, eventVarMap, cmdVarMap, intermVarMap) {
				pass.Report(pass, node, "Detected `os` system call with tainted `event` data - potential command injection")
			} else if isDangerousVar(argNode, pass.FileContext.Source, eventVarMap, intermVarMap) {
				pass.Report(pass, node, "Detected `os` system call with tainted `event` data - potential command injection")
			}
		}
	})

	return nil, nil
}

func isDangerousVar(node *sitter.Node, source []byte, eventVarMap, intermVarMap map[string]bool) bool {
	if node.Type() != "identifier" {
		return false
	}

	return eventVarMap[node.Content(source)] || intermVarMap[node.Content(source)]
}

func isTaintedCmdString(node *sitter.Node, source []byte, eventVarMap, cmdVarMap, intermVarMap map[string]bool) bool {
	switch node.Type() {
	case "call":
		funcNode := node.ChildByFieldName("function")

		if funcNode.Type() != "attribute" {
			return false
		}

		funcObjNode := funcNode.ChildByFieldName("object")
		funcAttrNode := funcNode.ChildByFieldName("attribute")

		if funcAttrNode.Type() != "identifier" {
			return false
		}

		if funcAttrNode.Content(source) != "format" {
			return false
		}

		argsNode := node.ChildByFieldName("arguments")
		argsList := getNamedChildren(argsNode, 0)

		if funcObjNode.Type() == "identifier" {
			if !cmdVarMap[funcObjNode.Content(source)] {
				return false
			}

			for _, argVal := range argsList {
				if argVal.Type() == "identifier" {
					if eventVarMap[argVal.Content(source)] {
						return true
					}
				} else if argVal.Type() == "subscript" {
					return isEventSubscript(argVal, source)
				}
			}
		} else if funcObjNode.Type() == "string" {
			for _, argVal := range argsList {
				if argVal.Type() == "identifier" {
					if eventVarMap[argVal.Content(source)] {
						return true
					} else if argVal.Type() == "subscript" {
						return isEventSubscript(argVal, source)
					}
				}
			}
		}

	case "binary_operator":
		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if leftNode.Type() == "string" {
			if rightNode.Type() == "tuple" {
				tupleArgNodes := getNamedChildren(rightNode, 0)
				for _, tupArg := range tupleArgNodes {
					if tupArg.Type() == "identifier" {
						if eventVarMap[tupArg.Content(source)] {
							return true
						}
					} else if tupArg.Type() == "subscript" {
						return isEventSubscript(tupArg, source)
					}
				}
			} else if rightNode.Type() == "identifier" {
				return eventVarMap[rightNode.Content(source)]
			}
		} else if leftNode.Type() == "identifier" {
			if !cmdVarMap[leftNode.Content(source)] {
				return false
			}

			if rightNode.Type() == "tuple" {
				tupleArgNodes := getNamedChildren(rightNode, 0)

				for _, tupArg := range tupleArgNodes {
					if tupArg.Type() == "identifier" {
						if eventVarMap[tupArg.Content(source)] {
							return true
						} else if tupArg.Type() == "subscript" {
							return isEventSubscript(tupArg, source)
						}
					}
				}
			}
		}

	case "string":
		if node.Content(source)[0] != 'f' {
			return false
		}
		allStringChildren := getNamedChildren(node, 0)
		for _, strnode := range allStringChildren {
			if strnode.Type() == "interpolation" {
				exprNode := strnode.ChildByFieldName("expression")
				if isEventSubscript(exprNode, source) {
					return true
				}

				if eventVarMap[exprNode.Content(source)] {
					return true
				}
			}
		}

	}

	return false
}

func isEventSubscript(node *sitter.Node, source []byte) bool {
	if node.Type() != "subscript" {
		return false
	}

	valIdNode := node

	// when there are more than 1 subscript accesses, we need to go down the tree
	// to get to the identifier
	for valIdNode.Type() != "identifier" {
		valIdNode = valIdNode.ChildByFieldName("value")
	}

	eventIdentifier := valIdNode.Content(source)
	return strings.Contains(eventIdentifier, "event")
}
