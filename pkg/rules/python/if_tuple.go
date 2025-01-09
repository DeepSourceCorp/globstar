package python_rules

import (
	sitter "github.com/smacker/go-tree-sitter"
	one "github.com/srijan-paul/deepgrep/pkg/one"
)

func checkIfTuple(r one.Rule, ana *one.Analyzer, node *sitter.Node) {
	cond := node.ChildByFieldName("condition")
	if cond.Type() == "tuple" {
		ana.Report(&one.Issue{
			Message: "Do not use 'if' with a tuple argument, as tuples always evaluate to true",
			Range:   node.Range(),
		})
	}
}

func IfTuple() one.Rule {
	var entry one.VisitFn = checkIfTuple
	return one.CreateRule("if_statement", one.LangPy, &entry, nil)
}

