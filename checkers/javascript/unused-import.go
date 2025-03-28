package javascript

import (
	"fmt"

	"globstar.dev/analysis"
)

var UnusedImport = &analysis.Analyzer{
	Name:        "unused-import",
	Requires:    []*analysis.Analyzer{ScopeAnalyzer},
	Run:         checkUnusedImports,
	Language:    analysis.LangJs,
	Description: "This checker checks for unused imports in JavaScript code. Unused imports can be removed to reduce the size of the bundle. Unused imports are also a code smell and can indicate that the code is not well-organized.",
	Category:    analysis.CategoryAntipattern,
	Severity:    analysis.SeverityInfo,
}

func checkUnusedImports(pass *analysis.Pass) (interface{}, error) {
	// Get scope tree from previous analysis
	scopeResult := pass.ResultOf[ScopeAnalyzer]
	if scopeResult == nil {
		return nil, nil
	}

	scope := scopeResult.(*analysis.ScopeTree)

	for _, scope := range scope.ScopeOfNode {
		for _, variable := range scope.Variables {
			if variable.Kind == analysis.VarKindImport {
				if len(variable.Refs) == 0 {
					pass.Report(pass, variable.DeclNode, fmt.Sprintf("unused import \"%s\"", variable.Name))
				}
			}
		}
	}

	return nil, nil
}
