package python_rules

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/globstar/pkg/analysis"
)

func checkIfTuple(r analysis.Rule, ana *analysis.Analyzer, node *sitter.Node) {
	cond := node.ChildByFieldName("condition")
	if cond.Type() == "tuple" {
		ana.Report(&analysis.Issue{
			Message: "Do not use 'if' with a tuple argument, as tuples always evaluate to true",
			Range:   node.Range(),
		})
	}
}

func IfTuple() analysis.Rule {
	var entry analysis.VisitFn = checkIfTuple
	return analysis.CreateRule("if_statement", analysis.LangPy, &entry, nil)
}
