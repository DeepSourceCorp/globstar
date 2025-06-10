package javascript

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	ana "globstar.dev/analysis"
)

func TestPatternDetection(t *testing.T) {
	source := `
function getUserInput(key) {

    return document.getElementById(key).value;

}

userInput = getUserInput('username')

// A sink method, which performs some raw databse operation on the userInput
perform_db_operation(userInput)

}
`

	parseRes := parseJsCode(t, []byte(source))
	var TaintDetectorMock = &ana.Analyzer{
		Name:        "taint_detector",
		Language:    ana.LangJs,
		Description: "Taint detector",
		Category:    ana.CategorySecurity,
		Severity:    ana.SeverityCritical,
		Requires:    []*ana.Analyzer{DataFlowAnalyzer},
		Run: detectTaint([]string{`
		(call_expression
            function: (identifier) @sourceName 
              (#eq? @sourceName "getUserInput"))`}, []string{`
		(call_expression
            function: (identifier) @sinkName
              (#eq? @sinkName "perform_db_operation"))`}),
	}
	pass := &ana.Pass{
		FileContext: parseRes,
		Analyzer:    TaintDetectorMock,
	}
	patterns, err := TaintDetectorMock.Run(pass)
	assert.NoError(t, err)

	assert.Len(t, patterns.(map[string]interface{})["sinkNodes"], 1)
	assert.Len(t, patterns.(map[string]interface{})["sourceNodes"], 1)
	assert.Len(t, patterns.(map[string]interface{})["sinkPatterns"], 1)
	assert.Len(t, patterns.(map[string]interface{})["sourcePatterns"], 1)

	for _, node := range patterns.(map[string]interface{})["sinkNodes"].([]*sitter.Node) {
		t.Log(node.Content(parseRes.Source))
	}
}
