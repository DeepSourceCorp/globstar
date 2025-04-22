package javascript

import (
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var EvalExpress = &analysis.Analyzer{
	Name:        "eval_express",
	Language:    analysis.LangJs,
	Description: "Avoid dynamically evaluating untrusted input, which can lead to a code injection vulnerability.",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Requires:    []*analysis.Analyzer{DataFlowAnalyzer},
	Run:         detectEvalExpress,
}

func detectEvalExpress(pass *analysis.Pass) (interface{}, error) {
	dfg := pass.ResultOf[DataFlowAnalyzer].(*DataFlowGraph)
	if dfg == nil {
		return nil, nil
	}

	scopeTree := dfg.ScopeTree
	if scopeTree == nil {
		return nil, nil
	}

	flowGraph := dfg.Graph
	if flowGraph == nil {
		return nil, nil
	}

	// Common user-input sources for JavaScript applications.
	userInputSrc := []string{"req.query.input", "req.params.id", "req.body", "req.cookies.sessionId", "localStorage.getItem"}

	// These methods call eval directly, or under the hood. This leads to vulnerability.
	vulnMethods := []string{"eval", "setTimeout", "Function"}
	taintedNodes := []*DataFlowNode{}

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil {
			return
		}
		if len(vulnMethods) < 0 && len(userInputSrc) < 0 {
			return
		}

		if node.Type() == "variable_declarator" {
			currScope := scopeTree.GetScope(node)

			nameNode := node.ChildByFieldName("name")
			valueNode := node.ChildByFieldName("value")

			var nameVar *analysis.Variable
			if nameNode != nil {
				nameVar = currScope.Lookup(nameNode.Content(pass.FileContext.Source))
			}

			var nameDfNode *DataFlowNode

			if dfnode, ok := flowGraph[nameVar]; ok {
				nameDfNode = dfnode
			}

			if valueNode != nil {
				switch valueNode.Type() {
				// Track and mark any user-input sources, on the data-flow node level.
				case "member_expression", "identifier":
					valueContent := valueNode.Content(pass.FileContext.Source)
					if slices.Contains(userInputSrc, valueContent) {
						taintedNodes = append(taintedNodes, nameDfNode)
					}

				// Check for any possibility of vulnerability
				case "call_expression":
					funcNode := valueNode.ChildByFieldName("function")
					var funcName string
					if funcNode != nil {
						funcName = funcNode.Content(pass.FileContext.Source)
					}
					if slices.Contains(vulnMethods, funcName) {
						if ContainsAny(nameDfNode.Sources, taintedNodes) {
							pass.Report(pass, node, "Eval attempt on user input, code injection vulnerability.")
						}
					}

				case "new_expression":
					ctor := valueNode.ChildByFieldName("constructor")
					var ctorName string
					if ctor != nil {
						ctorName = ctor.Content(pass.FileContext.Source)
					}
					if ctorName == "Function" {
						args := valueNode.ChildByFieldName("arguments")
						if args != nil {
							for i := 0; i < int(args.NamedChildCount()); i++ {
								child := args.NamedChild(i)
								if child != nil && child.Type() == "template_string" {
									for j := 0; j < int(child.NamedChildCount()); j++ {
										exprNode := child.NamedChild(j)
										if exprNode != nil && exprNode.Type() == "template_substitution" {
											idNode := exprNode.NamedChild(0)
											if idNode != nil && idNode.Type() == "identifier" {
												varName := idNode.Content(pass.FileContext.Source)
												if variable := currScope.Lookup(varName); variable != nil {
													if slices.Contains(taintedNodes, flowGraph[variable]) {
														pass.Report(pass, node, "Eval attempt on user input, code injection vulnerability.")
													}
												}
											}
										}

									}
								}
							}
						}
					}
				}

			}
		}

	})

	return nil, nil
}

func ContainsAny[T comparable](a, b []T) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}

// TODO:
// - [] Add vuln detection for call Expressions.
// - [] Updated DFG to handle New Function() statements, instead of brute-forcing here
