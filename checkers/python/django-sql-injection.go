package python

import (
	"regexp"

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

	// detect RawSQL injections
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}


		funcNode := node.ChildByFieldName("function")
		funcName := funcNode.Content(pass.FileContext.Source)

		rawsqlpat := `\bRawSQL\b`
		re := regexp.MustCompile(rawsqlpat)

		if !re.MatchString(funcName) {
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
