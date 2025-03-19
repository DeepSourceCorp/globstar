package python

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DjangoSQLInjection *analysis.Analyzer = &analysis.Analyzer{
	Name:        "django-sql-injection",
	Language:    analysis.LangPy,
	Description: "dnwe2e",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkDjangoSQLInjection,
}

func checkDjangoSQLInjection(pass *analysis.Pass) (interface{}, error) {
	requestVarMap := make(map[string]bool)
	intermVarMap := make(map[string]bool)
	cursorVarMap := make(map[string]bool)

	// get request data variable names
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
			requestVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// get django.db.connection cursor var name
	analysis.Preorder(pass, func(node *sitter.Node) {
		switch node.Type() {
		case "assignment":
			leftNode := node.ChildByFieldName("left")
			rightNode := node.ChildByFieldName("right")

			if rightNode == nil {
				return
			}

			if rightNode.Type() != "call" {
				return
			}

			funcNode := rightNode.ChildByFieldName("function")
			if funcNode.Type() != "attribute" {
				return
			}

			funcObj := funcNode.ChildByFieldName("object")
			funcAttr := funcNode.ChildByFieldName("attribute")

			if funcAttr.Type() != "identifier" && funcObj.Type() != "identifier" {
				return
			}

			if funcObj.Content(pass.FileContext.Source) == "connection" && funcAttr.Content(pass.FileContext.Source) == "cursor" {
				cursorVarMap[leftNode.Content(pass.FileContext.Source)] = true
			}

		case "with_statement":
			withClauseNode := node.NamedChild(0)
			withItemNode := withClauseNode.NamedChild(0)
			valueNode := withItemNode.ChildByFieldName("value")

			if valueNode.Type() != "as_pattern" {
				return
			}

			callNode := valueNode.NamedChild(0)
			callFuncNode := callNode.ChildByFieldName("function")

			if callFuncNode.Type() != "attribute" && callFuncNode.Content(pass.FileContext.Source) != "connection.cursor" {
				return
			}

			aliasNode := valueNode.ChildByFieldName("alias")
			cursorVarMap[aliasNode.Content(pass.FileContext.Source)] = true
		}
	})

	// get user tainted data variables
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if isUserTainted(rightNode, pass.FileContext.Source, intermVarMap, requestVarMap) {
			intermVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// detect sql injections
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		if !isRawSqlMethod(node, pass.FileContext.Source) && !isCursorExecuteMethod(node, pass.FileContext.Source, cursorVarMap) && !isObjectRawMethod(node, pass.FileContext.Source) {
			return
		}

		funcArgslist := node.ChildByFieldName("arguments")
		if funcArgslist.Type() != "argument_list" {
			return
		}

		funcArgs := getNamedChildren(funcArgslist, 0)
		for _, arg := range funcArgs {
			if isUserTainted(arg, pass.FileContext.Source, intermVarMap, requestVarMap) {
				pass.Report(pass, node, "diwhiweew")
			}
		}
	})

	return nil, nil
}

func isRawSqlMethod(node *sitter.Node, source []byte) bool {
	funcNode := node.ChildByFieldName("function")
	funcName := funcNode.Content(source)

	rawsqlpat := `\bRawSQL\b`
	re := regexp.MustCompile(rawsqlpat)

	return re.MatchString(funcName)
}

func isCursorExecuteMethod(node *sitter.Node, source []byte, cursorVarMap map[string]bool) bool {
	funcNode := node.ChildByFieldName("function")
	if funcNode.Type() != "attribute" {
		return false
	}
	funcObj := funcNode.ChildByFieldName("object")
	funcAttr := funcNode.ChildByFieldName("attribute")

	return cursorVarMap[funcObj.Content(source)] && funcAttr.Content(source) == "execute"
}

func isObjectRawMethod(node *sitter.Node, source []byte) bool {
	funcNode := node.ChildByFieldName("function")
	if funcNode.Type() != "attribute" {
		return false
	}

	return strings.HasSuffix(funcNode.Content(source), ".objects.raw")
}