package javascript

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

// var TaintDetector = &analysis.Analyzer{
// 	Name:        "taint_detector",
// 	Language:    analysis.LangJs,
// 	Description: "Taint detector",
// 	Category:    analysis.CategorySecurity,
// 	Severity:    analysis.SeverityCritical,
// 	Requires:    []*analysis.Analyzer{DataFlowAnalyzer},
// 	Run:         detectTaint([]string{"sink"}, []string{"source"}),
// }

func detectTaint(sink []string, source []string) func(pass *analysis.Pass) (any, error) {

	return func(pass *analysis.Pass) (interface{}, error) {
		var sinkPatterns []*sitter.Query
		for _, sink := range sink {
			sinkPattern, err := sitter.NewQuery([]byte(sink), analysis.LangJs.Grammar())
			if err != nil {
				return nil, fmt.Errorf("failed to create sink pattern: %w", err)
			}
			sinkPatterns = append(sinkPatterns, sinkPattern)
		}

		var sourcePatterns []*sitter.Query
		for _, source := range source {
			sourcePattern, err := sitter.NewQuery([]byte(source), analysis.LangJs.Grammar())
			if err != nil {
				return nil, fmt.Errorf("failed to create source pattern: %w", err)
			}
			sourcePatterns = append(sourcePatterns, sourcePattern)
		}

		if len(sinkPatterns) == 0 || len(sourcePatterns) == 0 {
			return nil, fmt.Errorf("no patterns found")
		}

		var sourceNodes []*sitter.Node
		var sinkNodes []*sitter.Node
		for _, query := range sourcePatterns {
			qc := sitter.NewQueryCursor()
			defer qc.Close()
			qc.Exec(query, pass.FileContext.Ast)
			for {
				m, ok := qc.NextMatch()
				if !ok {
					break
				}
				m = qc.FilterPredicates(m, pass.FileContext.Source)
				for _, capture := range m.Captures {
					captureNode := capture.Node
					sourceNodes = append(sourceNodes, captureNode)
				}

			}
		}

		for _, query := range sinkPatterns {
			qc := sitter.NewQueryCursor()
			defer qc.Close()
			qc.Exec(query, pass.FileContext.Ast)
			for {
				m, ok := qc.NextMatch()
				if !ok {
					break
				}
				m = qc.FilterPredicates(m, pass.FileContext.Source)
				for _, capture := range m.Captures {
					captureNode := capture.Node
					sinkNodes = append(sinkNodes, captureNode)
				}
			}
		}

		if len(sinkNodes) == 0 || len(sourceNodes) == 0 {
			return nil, fmt.Errorf("no sink or source pattern matched")
		}

		return map[string]interface{}{
			"sinkNodes":      sinkNodes,
			"sourceNodes":    sourceNodes,
			"sinkPatterns":   sinkPatterns,
			"sourcePatterns": sourcePatterns,
		}, nil
	}
}
