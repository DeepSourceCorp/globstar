package javascript

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var NoDoubleEq *analysis.Analyzer = &analysis.Analyzer{
	Name:        "no-double-eq",
	Language:    analysis.LangJs,
	Description: "This checker checks for the usage of '==' in JavaScript code. It reports an issue if '==' is used for comparison. It suggests using '===' instead.",
	Category:    analysis.CategoryBugRisk,
	Severity:    analysis.SeverityWarning,
	Run:         noDoubleEq,
}

func noDoubleEq(pass *analysis.Pass) (interface{}, error) {
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "binary_expression" {
			operator := node.ChildByFieldName("operator")
			if operator == nil {
				return
			}

			if operator.Content(pass.FileContext.Source) == "==" {
				pass.Report(pass, operator, "Do not use '==' for comparison. Prefer '===' instead.")
			}
		}
	})
	return nil, nil
}
