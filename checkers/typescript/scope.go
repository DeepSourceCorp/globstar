// scope resolution implementation for TypeScript files
//
//globstar:registry-exclude
package typescript

import (
	"reflect"

	"globstar.dev/analysis"
)

var ScopeAnalyzer = &analysis.Analyzer{
	Name:       "ts-scope",
	ResultType: reflect.TypeOf(&analysis.ScopeTree{}),
	Run:        buildScopeTree,
	Language:   analysis.LangTs,
}

func buildScopeTree(pass *analysis.Pass) (any, error) {
	scope := analysis.MakeScopeTree(pass.Analyzer.Language, pass.FileContext.Ast, pass.FileContext.Source)
	return scope, nil
}
