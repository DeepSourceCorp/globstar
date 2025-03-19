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

		if rightNode.Type() != "string" && rightNode.Type() != "call" {
			return
		}

		if rightNode.Content(pass.FileContext.Source) == "\"\"" || rightNode.Content(pass.FileContext.Source) == "''" || isEmptyRequestGet(rightNode, pass.FileContext.Source) {
			emptyPasswordVarMap[leftNode.Content(pass.FileContext.Source)] = true
		}
	})

	// get function param names containing empty strings
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() != "function_definition" {
			return
		}

		parameterNode := node.ChildByFieldName("parameters")
		if parameterNode.Type() != "parameters" {
			return
		}

		paramNodesList := getNamedChildren(parameterNode, 0)
		for _, paramNode := range paramNodesList {
			if paramNode.Type() == "default_parameter" {
				name := paramNode.ChildByFieldName("name")
				valNode := paramNode.ChildByFieldName("value")
				if name.Type() == "identifier" && (valNode.Type() == "string" || valNode.Type() == "call") {
					val := valNode.Content(pass.FileContext.Source)
					if val == "\"\"" || val == "''" || isEmptyRequestGet(valNode, pass.FileContext.Source) {
						emptyPasswordVarMap[name.Content(pass.FileContext.Source)] = true
					}
				}
			}
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

func isEmptyRequestGet(node *sitter.Node, source []byte) bool {
	if node.Type() != "call" {
		return false
	}

	funcNode := node.ChildByFieldName("function")
	if funcNode.Type() != "attribute" {
		return false
	}

	// checking for the pattern request.<...>.get()
	if !strings.HasPrefix(funcNode.Content(source), "request.") && !strings.HasSuffix(funcNode.Content(source), ".get") {
		return false
	}

	argListNode := node.ChildByFieldName("arguments")
	if argListNode.Type() != "argument_list" {
		return false
	}
	if argListNode.NamedChildCount() < 2 {
		return false
	}
	emptyArg := argListNode.NamedChild(1)
	if emptyArg.Type() != "string" {
		return false
	}

	if emptyArg.Content(source) == "\"\"" || emptyArg.Content(source) == "''" {
		return true
	}
	return false
}
