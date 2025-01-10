package python_rules

import (
	sitter "github.com/smacker/go-tree-sitter"
	one "github.com/srijan-paul/deepgrep/pkg/one"
)

func checkComparisonOp(r one.Rule, ana *one.Analyzer, node *sitter.Node) {
	lhs := node.Child(0)
	operator := node.Child(1)
	rhs := node.Child(2)

	if operator == nil {
		return
	}

	operatorStr := operator.Content(ana.ParseResult.Source)
	if operatorStr != "is" {
		return
	}

	lhsIsLiteral := false
	if lhs != nil {
		lhsType := lhs.Type()
		lhsIsLiteral = lhsType == "integer" || lhsType == "float" || lhsType == "string" 
	}

	rhsIsLiteral := false
	if rhs != nil {
		rhsType := rhs.Type()
		rhsIsLiteral = rhsType == "integer" || rhsType == "float" || rhsType == "string"
	}

	if lhsIsLiteral || rhsIsLiteral {
		ana.Report(&one.Issue{
			Message: "Do not use 'is' to compare literals. Use '==' instead",
			Range:   node.Range(),
		})
	}
}

func IsLiteral() one.Rule {
	var entry one.VisitFn = checkComparisonOp 
	return one.CreateRule("comparison_operator", one.LangPy, &entry, nil)
}

