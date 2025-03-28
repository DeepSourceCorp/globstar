// scope resolution implementation for JS and TS files
//
//globstar:registry-exclude
package javascript

import (
	"reflect"

	"globstar.dev/analysis"
)

var ScopeAnalyzer = &analysis.Analyzer{
	Name:       "js-scope",
	ResultType: reflect.TypeOf(&analysis.ScopeTree{}),
	Run:        buildScopeTree,
	Language:   analysis.LangJs,
}

func buildScopeTree(pass *analysis.Pass) (any, error) {
	// Create scope builder for JavaScript
	scope := analysis.MakeScopeTree(pass.Analyzer.Language, pass.FileContext.Ast, pass.FileContext.Source)
	return scope, nil
}
