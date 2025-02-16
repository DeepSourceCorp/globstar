package js_rules

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/globstar/pkg/analysis"
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
