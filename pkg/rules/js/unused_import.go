package js_rules

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/globstar/pkg/analysis"
)

func importSpecifierName(importSpecifier *sitter.Node, source []byte) string {
	name := importSpecifier.ChildByFieldName("name")
	if importSpecifier.Child(2) != nil {
		// alias (<imported> as <local>)
		local := importSpecifier.Child(2)
		return local.Content(source)
	}

	// no alias
	return name.Content(source)
}

func isUnused(scopeTree *analysis.ScopeTree, name string) bool {
	variable, exists := scopeTree.Root.Variables[name]
	return !exists || len(variable.Refs) == 0
}

func checkUnusedImport(r analysis.Rule, ana *analysis.Analyzer, importClause *sitter.Node) {
	scopeTree := ana.ParseResult.ScopeTree
	if scopeTree == nil {
		return
	}

	namedImports := analysis.FirstChildOfType(importClause, "named_imports")
	if namedImports != nil {
		nChildren := int(namedImports.NamedChildCount())
		for i := 0; i < nChildren; i++ {
			importSpecifier := namedImports.NamedChild(i)
			if importSpecifier.Type() != "import_specifier" {
				continue
			}

			name := importSpecifierName(importSpecifier, ana.ParseResult.Source)
			if isUnused(scopeTree, name) {
				ana.Report(&analysis.Issue{
					Message: fmt.Sprintf("'%s' is imported but never used", name),
					Range:   importSpecifier.Range(),
				})
			}
		}
	}

	defaultImport := analysis.FirstChildOfType(importClause, "identifier")
	if defaultImport != nil {
		name := defaultImport.Content(ana.ParseResult.Source)
		if isUnused(scopeTree, name) {
			ana.Report(&analysis.Issue{
				Message: fmt.Sprintf("'%s' is imported but never used", name),
				Range:   defaultImport.Range(),
			})
		}
	}
}

func UnusedImport() analysis.Rule {
	var exit analysis.VisitFn = checkUnusedImport
	return analysis.CreateRule("import_clause", analysis.LangJs, nil, &exit)
}
