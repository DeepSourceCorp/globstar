package javascript

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var NoExec = &analysis.Analyzer{
	Name:        "no_exec",
	Language:    analysis.LangJs,
	Description: "Use of exec on user input can lead to code injection vulnerabilities",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Requires:    []*analysis.Analyzer{DataFlowAnalyzer},
	Run:         detectExecOnUserInput,
}

func detectExecOnUserInput(pass *analysis.Pass) (interface{}, error) {
	// Map of vulnerable function names to watch for
	vulnerableSinkFunctions := map[string]struct{}{
		"exec": struct{}{},
	}

	vulnerableSourceFuncs := map[string]struct{}{
		"prompt": struct{}{},
	}

	dfg := pass.ResultOf[DataFlowAnalyzer].(*DataFlowGraph)
	if dfg == nil {
		return nil, nil
	}

	scopeTree := dfg.ScopeTree

	if scopeTree == nil {
		return nil, nil
	}

	// Map to maintain any tainted nodes
	taintedNodes := make(map[*sitter.Node]struct{})
	analyzedFuncs := make(map[string]bool) // Prevent infinite recursion

	// stores all the possible tainted inputs
	possibleTaintedInputs := make(map[*sitter.Node]string)

	var analyzeFunctionBody func(node *sitter.Node) bool

	analyzeFunctionBody = func(node *sitter.Node) bool {
		if node == nil {
			return false
		}

		// Check if this node is a sink
		if node.Type() == "call_expression" {
			funcNode := node.ChildByFieldName("function")
			if funcNode != nil && funcNode.Type() == "member_expression" {
				prop := funcNode.ChildByFieldName("property")
				if prop != nil {
					funcName := prop.Content(pass.FileContext.Source)
					if _, ok := vulnerableSinkFunctions[funcName]; ok {
						// Check if any argument is tainted
						args := node.ChildByFieldName("arguments")
						if int(args.NamedChildCount()) > 0 {
							return true
						}
					}
				}
			}
		}
		// Recursively analyze children
		vulnFound := false
		for i := 0; i < int(node.NamedChildCount()); i++ {
			if analyzeFunctionBody(node.NamedChild(i)) {
				vulnFound = true
			}
		}
		return vulnFound
	}

	// First pass: collect all suspicious source functions
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil || node.Type() != "variable_declarator" {
			return
		}
		nameNode := node.ChildByFieldName("name")
		valueNode := node.ChildByFieldName("value")
		if nameNode == nil || nameNode.Type() != "identifier" {
			return
		}

		if valueNode == nil || valueNode.Type() != "call_expression" {
			return
		}

		funcNode := valueNode.ChildByFieldName("function")
		if funcNode == nil || funcNode.Type() != "member_expression" {
			return
		}

		propertyNode := funcNode.ChildByFieldName("property")
		if propertyNode == nil {
			return
		}

		// Extract the function name
		funcName := propertyNode.Content(pass.FileContext.Source)
		if _, ok := vulnerableSourceFuncs[funcName]; !ok {
			return
		}
		possibleTaintedInputs[nameNode] = funcName
	})

	functionDefs := dfg.FuncDefs
	if functionDefs == nil {
		return nil, nil
	}

	// Second pass: analyze function bodies
	for funcName, funcDef := range functionDefs {
		if analyzedFuncs[funcName] {
			continue // Skip if already analyzed
		}
		analyzedFuncs[funcName] = true

		// Track which parameters might be tainted

		// Check function parameters against known tainted inputs

		// Analyze function body
		if analyzeFunctionBody(funcDef.Node) {
			// If vulnerability found, mark the function as tainted
			if funcDef.Node != nil {
				taintedNodes[funcDef.Node] = struct{}{}
			}
		}
	}

	if len(taintedNodes) == 0 {
		return nil, nil
	}

	// analyze and mark functions as tainted
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node == nil || node.Type() != "call_expression" {
			return
		}

		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			return
		}

		switch funcNode.Type() {
		case "member_expression":
			return
		case "identifier":
			// Check if the function is in the list of known tainted functions
			funcName := funcNode.Content(pass.FileContext.Source)
			if functionDefs[funcName] != nil {
				taintNode := functionDefs[funcName].Node
				if _, ok := taintedNodes[taintNode]; ok {
					// Report the issue
					pass.Report(pass, node, "Potentially unsafe function call with tainted input")
				}
			}

		}

	})

	return nil, nil
}
