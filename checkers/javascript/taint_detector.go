package javascript

// var TaintAnalyzer = &analysis.Analyzer{
// 	Name:        "taint_analysis_function",
// 	Language:    analysis.LangJs,
// 	Description: "Taint detector",
// 	Category:    analysis.CategorySecurity,
// 	Severity:    analysis.SeverityCritical,
// 	Requires:    []*analysis.Analyzer{DataFlowAnalyzer},
// }

// func detectTaint(source []string, sink []string) func(pass *analysis.Pass) (any, error) {
// 	return func(pass *analysis.Pass) (interface{}, error) {
// 		dfg := pass.ResultOf[DataFlowAnalyzer].(*DataFlowGraph)
// 		if dfg == nil {
// 			return nil, fmt.Errorf("no data flow graph found")
// 		}
// 		scopeTree := dfg.ScopeTree
// 		if scopeTree == nil {
// 			fmt.Println("no scope tree found")
// 			return nil, fmt.Errorf("no scope tree found")
// 		}

// 		var sinkPatterns []*sitter.Query
// 		for _, sink := range sink {
// 			sinkPattern, err := sitter.NewQuery([]byte(sink), analysis.LangJs.Grammar())
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to create sink pattern: %w", err)
// 			}
// 			sinkPatterns = append(sinkPatterns, sinkPattern)
// 		}

// 		var sourcePatterns []*sitter.Query
// 		for _, source := range source {
// 			sourcePattern, err := sitter.NewQuery([]byte(source), analysis.LangJs.Grammar())
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to create source pattern: %w", err)
// 			}
// 			sourcePatterns = append(sourcePatterns, sourcePattern)
// 		}

// 		if len(sinkPatterns) == 0 || len(sourcePatterns) == 0 {
// 			return nil, fmt.Errorf("no patterns found")
// 		}

// 		var sourceNodes []*sitter.Node
// 		var sinkNodes []*sitter.Node
// 		for _, query := range sourcePatterns {
// 			qc := sitter.NewQueryCursor()
// 			defer qc.Close()
// 			qc.Exec(query, pass.FileContext.Ast)
// 			for {
// 				m, ok := qc.NextMatch()
// 				if !ok {
// 					break
// 				}
// 				m = qc.FilterPredicates(m, pass.FileContext.Source)
// 				for _, capture := range m.Captures {
// 					captureNode := capture.Node
// 					sourceNodes = append(sourceNodes, captureNode)
// 				}

// 			}
// 		}

// 		for _, query := range sinkPatterns {
// 			qc := sitter.NewQueryCursor()
// 			defer qc.Close()
// 			qc.Exec(query, pass.FileContext.Ast)
// 			for {
// 				m, ok := qc.NextMatch()
// 				if !ok {
// 					break
// 				}
// 				m = qc.FilterPredicates(m, pass.FileContext.Source)
// 				for _, capture := range m.Captures {
// 					captureNode := capture.Node
// 					sinkNodes = append(sinkNodes, captureNode)
// 				}
// 			}
// 		}

// 		if len(sinkNodes) == 0 || len(sourceNodes) == 0 {
// 			return nil, fmt.Errorf("no sink or source pattern matched")
// 		}

// 		pass.Report(pass, sinkNodes[0], "sink node found")
// 		// // Get the data flow graph to track variable relationships

// 		// // Track source variables that flow into sinks
// 		// var taintedFlows []struct {
// 		// 	source *sitter.Node
// 		// 	sink   *sitter.Node
// 		// }

// 		// // For each source node, get its variable
// 		// for _, sourceNode := range sourceNodes {
// 		// 	// Get the assignment node (parent.parent.parent of source capture)
// 		// 	assignNode := sourceNode.Parent().Parent().Parent()
// 		// 	if assignNode == nil {
// 		// 		continue
// 		// 	}

// 		// 	// Get the identifier node and its scope
// 		// 	idNode := assignNode.ChildByFieldName("left")
// 		// 	if idNode == nil {
// 		// 		continue
// 		// 	}

// 		// 	idScope := scopeTree.GetScope(idNode)
// 		// 	if idScope == nil {
// 		// 		continue
// 		// 	}

// 		// 	// Look up the variable for the identifier
// 		// 	sourceVar := idScope.Lookup(idNode.Content(pass.FileContext.Source))
// 		// 	if sourceVar == nil {
// 		// 		continue
// 		// 	}

// 		// 	// For each sink, check if it uses the source variable
// 		// 	for _, sinkNode := range sinkNodes {
// 		// 		// Get the call expression node
// 		// 		callNode, err := analysis.GetRootNode(sinkNode)
// 		// 		if err != nil {
// 		// 			continue
// 		// 		}

// 		// 		// Get the argument node and its variable
// 		// 		argsNode := callNode.ChildByFieldName("arguments")
// 		// 		if argsNode == nil || argsNode.NamedChildCount() == 0 {
// 		// 			continue
// 		// 		}
// 		// 		argNode := argsNode.NamedChild(0)

// 		// 		argScope := scopeTree.GetScope(callNode)
// 		// 		if argScope == nil {
// 		// 			continue
// 		// 		}

// 		// 		argVar := argScope.Lookup(argNode.Content(pass.FileContext.Source))
// 		// 		if argVar == nil {
// 		// 			continue
// 		// 		}

// 		// 		// If the argument variable matches the source variable, we found a tainted flow
// 		// 		if argVar == sourceVar {
// 		// 			taintedFlows = append(taintedFlows, struct {
// 		// 				source *sitter.Node
// 		// 				sink   *sitter.Node
// 		// 			}{sourceNode, sinkNode})
// 		// 		}
// 		// 	}
// 		// }

// 		// for _, tainted := range taintedFlows {
// 		// 	pass.Report(pass, tainted.sink, "")
// 		// }

// 		return map[string]interface{}{
// 			"sinkNodes":      sinkNodes,
// 			"sourceNodes":    sourceNodes,
// 			"sinkPatterns":   sinkPatterns,
// 			"sourcePatterns": sourcePatterns,
// 		}, nil
// 	}
// }
