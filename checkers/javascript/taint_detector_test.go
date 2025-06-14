package javascript

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"globstar.dev/analysis"
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

	var TaintDetector = &analysis.Analyzer{
		Name:        "taint_detector",
		Language:    analysis.LangJs,
		Description: "Taint detector",
		Category:    analysis.CategorySecurity,
		Severity:    analysis.SeverityCritical,
		Requires:    []*analysis.Analyzer{DataFlowAnalyzer},
		Run: detectTaint([]string{`
	(expression_statement
		(assignment_expression
		  	right: (call_expression
             			function: (identifier) @sourceName 
               			  ))(#eq? @sourceName "getUserInput"))`}, []string{`
		(call_expression
            function: (identifier) @sinkName
              (#eq? @sinkName "perform_db_operation"))`}),
	}
	parseRes := parseJsCode(t, []byte(source))
	pass := &analysis.Pass{
		FileContext: parseRes,
		Analyzer:    TaintDetector,
	}
	patterns, err := TaintDetector.Run(pass)
	assert.NoError(t, err)

	dfg, err := createDataFlowGraph(pass)
	assert.NoError(t, err)

	scopeTree := dfg.(*DataFlowGraph).ScopeTree
	assert.NotNil(t, scopeTree)

	assert.Len(t, patterns.(map[string]interface{})["sinkNodes"], 1)
	assert.Len(t, patterns.(map[string]interface{})["sourceNodes"], 1)
	assert.Len(t, patterns.(map[string]interface{})["sinkPatterns"], 1)
	assert.Len(t, patterns.(map[string]interface{})["sourcePatterns"], 1)

	var sourceVar *analysis.Variable

	for _, node := range patterns.(map[string]interface{})["sourceNodes"].([]*sitter.Node) {
		parentNode := node.Parent().Parent().Parent()

		idScope := scopeTree.GetScope(parentNode.ChildByFieldName("left"))
		assert.NotNil(t, idScope)

		idVar := idScope.Lookup(parentNode.ChildByFieldName("left").Content(parseRes.Source))
		assert.NotNil(t, idVar)

		sourceVar = idVar
	}

	for _, node := range patterns.(map[string]interface{})["sinkNodes"].([]*sitter.Node) {
		parentNode, err := analysis.GetRootNode(node)
		assert.NoError(t, err)
		arg := parentNode.ChildByFieldName("arguments").NamedChild(0)

		scope := scopeTree.GetScope(parentNode)
		scopeVar := scope.Lookup(arg.Content(parseRes.Source))

		assert.Equal(t, scopeVar, sourceVar)
	}

}
