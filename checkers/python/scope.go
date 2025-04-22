//globstar:registry-exclude
// scope resolution for Python files

package python

import (
	"globstar.dev/analysis"
	"reflect"
)

var ScopeAnalyzer = &analysis.Analyzer{
	Name:       "py-scope",
	ResultType: reflect.TypeOf(&analysis.ScopeTree{}),
	Run:        buildScopeTree,
	Language:   analysis.LangPy,
}

func buildScopeTree(pass *analysis.Pass) (any, error) {
	// creates scope builder for python
	scope := analysis.MakeScopeTree(pass.Analyzer.Language, pass.FileContext.Ast, pass.FileContext.Source)
	return scope, nil
}
