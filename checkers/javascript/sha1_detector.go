package javascript

import (
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var Sha1 = &analysis.Analyzer{
	Name:        "sha1_detector",
	Language:    analysis.LangJs,
	Description: "Avoid using SHA1 for cryptographic purposes",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityWarning,
	Requires:    []*analysis.Analyzer{DataFlowAnalyzer},
	Run:         detectSha1Usage,
}

func detectSha1Usage(pass *analysis.Pass) (interface{}, error) {
	pkgs := []string{"jssha", "jssha/sha1", "jssha/dist/sha1"}

	// Will be used to track the Encrypting library being used
	var pkgDeclaratorVar []*analysis.Variable

	dfg := pass.ResultOf[DataFlowAnalyzer].(*DataFlowGraph)

	if dfg == nil {
		return nil, nil
	}

	scopeTree := dfg.ScopeTree

	if scopeTree == nil {
		return nil, nil
	}

	// gather the variable linking to the js-sha package.
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}

		if node.Type() == "variable_declarator" {
			nameNode := node.ChildByFieldName("name")
			valueNode := node.ChildByFieldName("value")
			if valueNode != nil {
				if valueNode.Type() == "call_expression" {
					funcNode := valueNode.ChildByFieldName("function")
					if funcNode != nil && funcNode.Type() == "member_expression" {
						if funcNode.Content(pass.FileContext.Source) == "require" {
							// check if the value inside the `require` is a valid package
							packageName := valueNode.ChildByFieldName("arguments").Child(0)
							if packageName != nil && packageName.Type() == "string" {
								packageNameContent := packageName.Content(pass.FileContext.Source)
								if !slices.Contains(pkgs, packageNameContent) {
									// If the package is not in the list, we can ignore it
									return
								}
							}
						}
					}
				}
			}

			if nameNode != nil && nameNode.Type() == "identifier" {
				varName := nameNode.Content(pass.FileContext.Source)
				if varName != "" {
					nameVar := scopeTree.GetScope(node).Lookup(varName)
					if nameVar != nil {
						pkgDeclaratorVar = append(pkgDeclaratorVar, nameVar)
					}
				}
			}
		}
		if node.Type() == "import_statement" {
			// Handle the case for import declaration eg. import jssha from "jssha"

			packageName := node.ChildByFieldName("source")
			if packageName != nil && packageName.Type() == "string" {
				packageNameContent := packageName.NamedChild(0).Content(pass.FileContext.Source)
				if packageNameContent != "jssha" {
					return
				}

			}

			importField := node.NamedChild(0)

			if importField != nil && importField.Type() == "import_clause" {
				importIdentifier := importField.NamedChild(0)
				if importIdentifier != nil && importIdentifier.Type() == "identifier" {
					idName := importIdentifier.Content(pass.FileContext.Source)
					if idName != "" {
						nameVar := scopeTree.GetScope(node).Lookup(idName)
						if nameVar != nil {
							pkgDeclaratorVar = append(pkgDeclaratorVar, nameVar)
						}
					}
				}
			}

		}
	})

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}

		if node.Type() == "new_expression" {
			ctor := node.ChildByFieldName("constructor")
			arg := node.ChildByFieldName("arguments")
			if ctor != nil && arg != nil {
				ctorVar := scopeTree.GetScope(ctor).Lookup(ctor.Content(pass.FileContext.Source))
				if ctorVar != nil && slices.Contains(pkgDeclaratorVar, ctorVar) {
					hashAlgo := arg.NamedChild(0)
					if hashAlgo == nil {
						return
					}

					hashAlgoStr := hashAlgo.NamedChild(0)
					hashAlgoName := hashAlgoStr.Content(pass.FileContext.Source)
					if hashAlgoName == "SHA-1" {
						pass.Report(pass, node, "SHA-1 is not recommended for cryptographic purposes")
					}

				}
			}

		}
	})

	return nil, nil

}
