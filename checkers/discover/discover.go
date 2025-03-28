package discover

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func DiscoverGoCheckers(dir string) ([]string, error) {
	goCheckers := []string{}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.AllErrors&parser.SkipObjectResolution|parser.ParseComments)
	if err != nil {
		return goCheckers, err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			isExcluded := false
			if len(file.Comments) > 0 {
				firstCommentGroup := file.Comments[0]
				for _, comment := range firstCommentGroup.List {
					if strings.TrimSpace(comment.Text) == "//globstar:registry-exclude" {
						isExcluded = true
						break
					}
				}
			}

			// skip this checker since exclude directive comment exists
			if isExcluded {
				continue
			}
			globstarPkgName := ""
			for _, imp := range file.Imports {
				// check for registry exclude comment in checker file
				if imp.Path.Value == `"globstar.dev/analysis"` {
					if imp.Name == nil {
						globstarPkgName = "analysis"
					} else {
						globstarPkgName = imp.Name.Name
					}
					break
				}
			}

			if globstarPkgName == "" {
				continue
			}

			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					if decl.Tok != token.VAR {
						continue
					}

					for _, spec := range decl.Specs {
						valueSpec, ok := spec.(*ast.ValueSpec)
						if !ok {
							continue
						}

						if len(valueSpec.Names) != 1 || len(valueSpec.Values) != 1 {
							continue
						}

						expr, ok := valueSpec.Values[0].(*ast.UnaryExpr)
						if !ok {
							continue
						}

						lit, ok := expr.X.(*ast.CompositeLit)
						if !ok {
							continue
						}

						selectorExpr, ok := lit.Type.(*ast.SelectorExpr)
						if !ok {
							continue
						}

						ident, ok := selectorExpr.X.(*ast.Ident)
						if !ok {
							continue
						}

						if selectorExpr.Sel.Name != "Analyzer" || ident.Name != globstarPkgName {
							continue
						}

						goCheckers = append(goCheckers, fmt.Sprintf("%s.%s", pkg.Name, valueSpec.Names[0].Name))
					}
				default:
					continue
				}
			}
		}
	}

	return goCheckers, nil
}
