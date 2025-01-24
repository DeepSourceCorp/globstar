package python_rules

import (
	"github.com/DeepSourceCorp/globstar/pkg/analysis"
	sitter "github.com/smacker/go-tree-sitter"
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
