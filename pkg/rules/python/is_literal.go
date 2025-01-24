package python_rules

import (
	"github.com/DeepSourceCorp/globstar/pkg/analysis"
	sitter "github.com/smacker/go-tree-sitter"
)

func checkComparisonOp(r analysis.Rule, ana *analysis.Analyzer, node *sitter.Node) {
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
		ana.Report(&analysis.Issue{
			Message: "Do not use 'is' to compare literals. Use '==' instead",
			Range:   node.Range(),
		})
	}
}

func IsLiteral() analysis.Rule {
	var entry analysis.VisitFn = checkComparisonOp
	return analysis.CreateRule("comparison_operator", analysis.LangPy, &entry, nil)
}
