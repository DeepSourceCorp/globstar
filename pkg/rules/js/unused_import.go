package js_rules

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	one "github.com/srijan-paul/deepgrep/pkg/one"
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

func isUnused(scopeTree *one.ScopeTree, name string) bool {
	variable, exists := scopeTree.Root.Variables[name]
	return !exists || len(variable.Refs) == 0
}

func checkUnusedImport(r one.Rule, ana *one.Analyzer, importClause *sitter.Node) {
	scopeTree := ana.ParseResult.ScopeTree
	if scopeTree == nil {
		return
	}

	namedImports := one.FirstChildOfType(importClause, "named_imports")
	if namedImports != nil {
		nChildren := int(namedImports.NamedChildCount())
		for i := 0; i < nChildren; i++ {
			importSpecifier := namedImports.NamedChild(i)
			if importSpecifier.Type() != "import_specifier" {
				continue
			}

			name := importSpecifierName(importSpecifier, ana.ParseResult.Source)
			if isUnused(scopeTree, name) {
				ana.Report(&one.Issue{
					Message: fmt.Sprintf("'%s' is imported but never used", name),
					Range:   importSpecifier.Range(),
				})
			}
		}
	}


	defaultImport := one.FirstChildOfType(importClause, "identifier")
	if defaultImport != nil {
		name := defaultImport.Content(ana.ParseResult.Source)
		if isUnused(scopeTree, name) {
			ana.Report(&one.Issue{
				Message: fmt.Sprintf("'%s' is imported but never used", name),
				Range:   defaultImport.Range(),
			})
		}
	}
}

func UnusedImport() one.Rule {
	var exit one.VisitFn = checkUnusedImport
	return one.CreateRule("import_clause", one.LangJs, nil, &exit)
}

