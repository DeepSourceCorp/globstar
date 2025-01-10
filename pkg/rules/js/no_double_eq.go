package js_rules

import (
	sitter "github.com/smacker/go-tree-sitter"
	one "github.com/srijan-paul/deepgrep/pkg/one"
)

func noDoubleEq(r one.Rule, ana *one.Analyzer, node *sitter.Node) {
	op := node.ChildByFieldName("operator")
	if op.Content(ana.ParseResult.Source) == "==" {
		ana.Report(&one.Issue{
			Message: "Do not use '==' for comparison. Prefer '===' instead.",
			Range:   op.Range(),
		})
	}
}

func NoDoubleEq() one.Rule {
	var entry one.VisitFn = noDoubleEq
	return one.CreateRule("binary_expression", one.LangJs, &entry, nil)
}
