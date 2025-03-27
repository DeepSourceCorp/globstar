package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var AwsLambdaTaintedSqlString *analysis.Analyzer = &analysis.Analyzer{
	Name:        "aws-lambda-tainted-sql-string",
	Language:    analysis.LangPy,
	Description: "Detected user input directly used to build a SQL string, which can lead to SQL injection vulnerabilities. This could allow an attacker to steal or modify database contents. Use parameterized queries or an ORM like Sequelize to protect against injection attacks.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkAwsLambdaTaintedSqlString,
}

func checkAwsLambdaTaintedSqlString(pass *analysis.Pass) (interface{}, error) {
	eventVarMap := make(map[string]bool)
	sqlStringVarMap := make(map[string]bool)

	// store variables for event subscripts and sql strings
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isEventSubscript(rightNode, pass.FileContext.Source) {
			eventVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}

		if isSqlString(rightNode, pass.FileContext.Source) {
			sqlStringVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// detect tainted sql strings in assignments
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isTaintedSqlString(rightNode, pass.FileContext.Source, sqlStringVarMap, eventVarMap) {
			pass.Report(pass, node, "Detect `event` tainted SQL string - Potential SQL injection vulnerabilities")
		}
	})

	return nil, nil
}

func isTaintedSqlString(node *sitter.Node, source []byte, sqlStringVarMap, eventVarMap map[string]bool) bool {
	switch node.Type() {
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
			if !sqlStringVarMap[leftNode.Content(source)] {
				return false
			}

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
			}
		}

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

		if funcObjNode.Type() == "string" {
			for _, argVal := range argsList {
				if argVal.Type() == "identifier" {
					if eventVarMap[argVal.Content(source)] {
						return true
					}
				} else if argVal.Type() == "subscript" {
					return isEventSubscript(argVal, source)
				}
			}
		} else if funcObjNode.Type() == "identifier" {
			if !sqlStringVarMap[funcObjNode.Content(source)] {
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

func isSqlString(node *sitter.Node, source []byte) bool {
	if node.Type() != "string" {
		return false
	}

	sqlPattern := regexp.MustCompile(`\s*(?i)(select|delete|insert|create|update|alter|drop)\b.*=`)
	return sqlPattern.MatchString(node.Content(source))
}

func isEventSubscript(node *sitter.Node, source []byte) bool {
	if node.Type() != "subscript" {
		return false
	}

	if isRequestCall(node, source) {
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
