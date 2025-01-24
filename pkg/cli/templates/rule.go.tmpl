package js_rules

import (
	"github.com/DeepSourceCorp/globstar/pkg/analysis"
	sitter "github.com/smacker/go-tree-sitter"
)

func noDoubleEq(r analysis.Rule, ana *analysis.Analyzer, node *sitter.Node) {
	op := node.ChildByFieldName("operator")
	if op.Content(ana.ParseResult.Source) == "==" {
		ana.Report(&analysis.Issue{
			Message: "Do not use '==' for comparison. Prefer '===' instead.",
			Range:   op.Range(),
		})
	}
}

func NoDoubleEq() analysis.Rule {
	var entry analysis.VisitFn = noDoubleEq
	return analysis.CreateRule("binary_expression", analysis.LangJs, &entry, nil)
}
