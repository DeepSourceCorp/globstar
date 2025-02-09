package analysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func DiscoverGoRules(dir string) ([]string, error) {
	goRules := []string{}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.AllErrors&parser.SkipObjectResolution)
	if err != nil {
		return goRules, err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				switch fn := n.(type) {
				case *ast.FuncDecl:
					if !fn.Name.IsExported() {
						return true
					}
					// match the signature of the function to be analysis.Rule
					if fn.Type.Params.NumFields() != 0 {
						return true
					}
					if fn.Type.Results.NumFields() != 1 {
						return true
					}

					selExpr, ok := fn.Type.Results.List[0].Type.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					if selExpr.Sel.Name != "Rule" || selExpr.X.(*ast.Ident).Name != "analysis" {
						return true
					}
					goRules = append(goRules, fmt.Sprintf("%s.%s", pkg.Name, fn.Name.Name))
				default:
					return true
				}
				return true
			})
		}
	}

	return goRules, nil
}

func GenerateMainFunction(goRules []string) string {
	mainFunc := `package main

import (
	"github.com/DeepSourceCorp/globstar/pkg/analysis"
)

func main() {
	rules := []analysis.Rule{
	`
	for _, rule := range goRules {
		mainFunc += fmt.Sprintf("%s(),\n", rule)
	}
	mainFunc += `}
	// run the analysis
}
`
	return mainFunc
}
