// TSX (TypeScript + JSX) support
//
//globstar:registry-exclude
package typescript

import (
	"reflect"

	"globstar.dev/analysis"
)

var TsxAnalyzer = &analysis.Analyzer{
	Name:       "tsx",
	ResultType: reflect.TypeOf(&analysis.ScopeTree{}),
	Run:        buildTsxTree,
	Language:   analysis.LangTsx,
}

func buildTsxTree(pass *analysis.Pass) (any, error) {
	// TSX uses same scope resolution as TypeScript
	scope := analysis.MakeScopeTree(pass.Analyzer.Language, pass.FileContext.Ast, pass.FileContext.Source)
	return scope, nil
}
