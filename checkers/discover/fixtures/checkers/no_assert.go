//registry:exclude
package checkers

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var NoAssert *analysis.Analyzer = &analysis.Analyzer{
	Name:        "no-assert",
	Language:    analysis.LangPy,
	Description: "This checker checks for the usage of `assert` statement in Python code. It is risky as they are removed when Python is run optimized mode",
	Category:    analysis.CategoryBugRisk,
	Severity:    analysis.SeverityWarning,
	Run:         checkNoAssert,
}

func checkNoAssert(pass *analysis.Pass) (interface{}, error) {
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "assert_statement" {
			pass.Report(pass, node, "Do not use assert statement to enforce constraints. They are removed in optimized bytecode")
		}
	})

	return nil, nil
}
