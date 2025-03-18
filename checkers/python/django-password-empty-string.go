package python

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DjangoPasswordEmptyString *analysis.Analyzer = &analysis.Analyzer{
	Name:        "django-password-empty-string",
	Language:    analysis.LangPy,
	Description: "Using empty string as password in `set_password()` may cause unintended behavior. To set an unusable password please set the password to None or call `set_unusable_password()`",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Run:         checkDjangoPasswordEmptyString,
}

func checkDjangoPasswordEmptyString(pass *analysis.Pass) (interface{}, error) {
	emptyPasswordVarMap := make(map[string]bool)

	// get variables containing empty strings
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "assignment" {
			return
		}

		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if rightNode == nil {
			return
		}

		if rightNode.Type() != "string" {
			return
		}

		if rightNode.Content(pass.FileContext.Source) == "\"\"" || rightNode.Content(pass.FileContext.Source) == "''" {
			emptyPasswordVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// catch empty string used as password
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode.Type() != "attribute" {
			return
		}

		if !strings.HasSuffix(funcNode.Content(pass.FileContext.Source), ".set_password") {
			return
		}

		argList := node.ChildByFieldName("arguments")
		if argList.Type() != "argument_list" {
			return
		}

		argNodes := getNamedChildren(argList, 0)
		for _, arg := range argNodes {
			if emptyPasswordVarMap[arg.Content(pass.FileContext.Source)] {
				pass.Report(pass, node, "Detected empty string being used as password - may case unintended behavior")
			} else if arg.Type() == "string" {
				passwd := arg.Content(pass.FileContext.Source)
				if passwd == "\"\"" || passwd == "''" {
					pass.Report(pass, node, "Detected empty string being used as password - may case unintended behavior")
				}
			}
		}
	})

	return nil, nil
}
