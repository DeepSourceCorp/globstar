package python

import (

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"

	"strings"
)

var AwsLambdaMySqlInjection *analysis.Analyzer = &analysis.Analyzer{
	Name:        "aws-lambda-mysqli",
	Language:    analysis.LangPy,
	Description: "An SQL statement in the code is using data from the event object, which could lead to SQL injection if the input is user-controlled and not properly sanitized. To prevent this vulnerability, it's recommended to use parameterized queries or prepared statements. For example, instead of directly embedding variables in the query, use cursor.execute('SELECT * FROM projects WHERE status = %s', ('active',)) to ensure safe query execution.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkAwsLambdaMySqlInjection,
}

func checkAwsLambdaMySqlInjection(pass *analysis.Pass) (interface{}, error) {
	sqlCursorMap := make(map[string]bool)
	sqlStringMap := make(map[string]bool)
	eventVarMap := make(map[string]bool)

	// first pass: get cursor variable and event variable names
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if isSqlCursor(rightNode, pass.FileContext.Source) {
			sqlCursorMap[leftNode.Content(pass.FileContext.Source)] = true
		}

		if isEventVar(rightNode, pass.FileContext.Source) {
			eventVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// second pass: get tainted sql string variable name
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if isTaintedSqlString(rightNode, pass.FileContext.Source, eventVarMap) {
			sqlStringMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// check for insecure sql query calls
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		funcCall := funcNode.Content(pass.FileContext.Source)
		if !strings.Contains(funcCall, "execute") {
			return
		}

		isSqlVarPresent := false
		for key := range sqlCursorMap {
			if strings.Contains(funcCall, key) {
				isSqlVarPresent = true
			}
		}

		if !isSqlVarPresent {
			return
		}

		argNode := node.ChildByFieldName("arguments")

		// Tainted sql strings will only contain a single variable
		// or a single formatted string. If argument count is > 1 then it must be
		// a parametrized sql query which is safe
		if argNode != nil && argNode.NamedChildCount() > 1 {
			return
		}

		if isDangerousArgument(argNode, pass.FileContext.Source, sqlStringMap, eventVarMap) {
			pass.Report(pass, node, "Detected SQL injection risk. Query contains tainted data.")
		}
	})

	return nil, nil
}

func isDangerousArgument(node *sitter.Node, source []byte, sqlVarMap, eventVarMap map[string]bool) bool {
	if node.Type() != "argument_list" {
		return false
	}

	arg := node.NamedChild(0)
	switch arg.Type() {
	case "identifier":
		argContent := arg.Content(source)
		for key := range sqlVarMap {
			if strings.Contains(argContent, key) {
				return true
			}
		}
	case "binary_operator":
		rightNode := arg.ChildByFieldName("right")
		rightNodeContent := rightNode.Content(source)

		// check if variable containing tainted is present in the query
		for key := range eventVarMap {
			if strings.Contains(rightNodeContent, key) {
				return true
			}
		}

		// check if the `event` variable is used in the sql query
		if strings.Contains(rightNodeContent, "event") {
			return true
		}

	case "string":
		strContent := arg.Content(source)
		// check if f-string
		if strContent[0] != 'f' {
			return false
		}

		allChildren := getAllChildren(arg, 0)

		// check if tainted data is present in f-string interpolation
		for _, child := range allChildren {
			if child.Type() == "interpolation" {
				interpolationContent := child.Content(source)
				if strings.Contains(interpolationContent, "event") {
					return true
				}

				for key := range eventVarMap {
					if strings.Contains(interpolationContent, key) {
						return true
					}
				}
			}
		}
	case "call":
		funcNode := arg.ChildByFieldName("function")
		funcAttr := funcNode.Content(source)
		if !strings.HasSuffix(funcAttr, ".format") {
			return false
		}
		funcArgNode := arg.ChildByFieldName("arguments")
		funcArgContent := funcArgNode.Content(source)

		// check if tainted
		for key := range eventVarMap {
			if strings.Contains(funcArgContent, key) {
				return true
			}
		}

		if strings.Contains(funcArgContent, "event") {
			return true
		}
	}
	return false
}

func isTaintedSqlString(node *sitter.Node, source []byte, eventVarMap map[string]bool) bool {
	switch node.Type() {
	case "binary_operator":
		rightNode := node.ChildByFieldName("right")
		rightNodeContent := rightNode.Content(source)

		// check if variable containing tainted `event` is passed to sql query
		for key := range eventVarMap {
			if strings.Contains(rightNodeContent, key) {
				return true
			}
		}

		// check if the `event` variable is directly used in the sql query
		if strings.Contains(rightNodeContent, "event") {
			return true
		}
	case "string":
		strContent := node.Content(source)

		// check if f-string
		if strContent[0] != 'f' {
			return false
		}

		allChildren := getAllChildren(node, 0)

		// check if tainted data is present in f-string interpolation
		for _, child := range allChildren {
			if child.Type() == "interpolation" {
				interpolationContent := child.Content(source)
				if strings.Contains(interpolationContent, "event") {
					return true
				}

				for key := range eventVarMap {
					if strings.Contains(interpolationContent, key) {
						return true
					}
				}
			}
		}

	case "call":
		funcNode := node.ChildByFieldName("function")
		funcAttr := funcNode.Content(source)
		if !strings.HasSuffix(funcAttr, ".format") {
			return false
		}
		argNode := node.ChildByFieldName("arguments")
		argContent := argNode.Content(source)

		// check if tainted
		for key := range eventVarMap {
			if strings.Contains(argContent, key) {
				return true
			}
		}

		if strings.Contains(argContent, "event") {
			return true
		}
	}
	return false
}

func isEventVar(node *sitter.Node, source []byte) bool {
	return node.Type() == "subscript" && strings.HasPrefix(node.Content(source), "event")
}

func isSqlCursor(node *sitter.Node, source []byte) bool {
	return node.Type() == "call" && strings.Contains(node.Content(source), "cursor")
}

func getAllChildren(node *sitter.Node, startIdx int) []*sitter.Node {
	childrenCount := node.NamedChildCount()
	childrenNodes := make([]*sitter.Node, 0)
	for i := startIdx; i < int(childrenCount); i++ {
		childrenNodes = append(childrenNodes, node.NamedChild(i))
	}
	return childrenNodes
}
