package golang

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var HiddenGoRoutineAnalyzer = &analysis.Analyzer{
	Name:        "hidden_goroutine",
	Language:    analysis.LangGo,
	Description: "A function that solely consists of launching a goroutine creates a `fire and forget` background process. Such functions provide no means of tracking the goroutine's execution, handling any errors that occur within it, or gracefully terminating it when needed.",
	Category:    analysis.CategoryBugRisk,
	Severity:    analysis.SeverityWarning,
	Run:         detectHiddenGoroutine,
}

// Logic: Get the function declaration -> Check if it contains *ONLY* goroutines -> Body can also contain comments.
func detectHiddenGoroutine(pass *analysis.Pass) (interface{}, error) {
	hidden := true
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}
		if node.Type() == "function_declaration" {
			funcBody := node.ChildByFieldName("body")
			if funcBody == nil {
				return
			}
			// Check the childs of the function body:
			for i := 0; i < int(funcBody.NamedChildCount()); i++ {
				child := funcBody.NamedChild(i)
				if child == nil {
					return
				}

				if child.Type() != "go_statement" && child.Type() != "comment" {
					hidden = false
				}

			}

			if hidden == false {
				return
			}

			pass.Report(pass, node, "")
		}

	})

	return nil, nil
}
