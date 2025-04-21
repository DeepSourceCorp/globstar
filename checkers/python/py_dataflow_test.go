package python

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"globstar.dev/analysis"
)

func parsePyCode(t *testing.T, source []byte) *analysis.ParseResult {
	pass, err := analysis.Parse("file.py", source, analysis.LangPy, analysis.LangPy.Grammar())
	require.NoError(t, err)

	return pass
}

func TestDataFlowAnalysis(t *testing.T) {
	t.Run("variable assignnment data flow", func(t *testing.T) {
		source := `
x = 10
def f(x):
	pass
f(x)
		`
		parseResult := parsePyCode(t, []byte(source))
		pass := &analysis.Pass{
			Analyzer: 
			DataFlowAnalyzer,
			FileContext: parseResult,
		}

		dfgStruct, err := createDataFlowGraph(pass)
		assert.NoError(t, err)
		dfg := dfgStruct.(*DataFlowGraph)
		flowGraph := dfg.Graph
		assert.NotNil(t, flowGraph)
		scopeTree := dfg.ScopeTree
		assert.NotNil(t, scopeTree)
	})
}

func TestClassDataFlow(t *testing.T) {
	source := `
class A:
	a = 10
	def __init__(self, a):
		self.a = a
	def f(self):
		return self.a
	`
	parseResult := parsePyCode(t, []byte(source))
	pass := &analysis.Pass{
		Analyzer: DataFlowAnalyzer,
		FileContext: parseResult,
	}
	dfgStruct, err := createDataFlowGraph(pass)
	assert.NoError(t, err)
	dfg := dfgStruct.(*DataFlowGraph)
	scopeTree := dfg.ScopeTree
	graph := dfg.Graph
	assert.NotNil(t, scopeTree)
	classVar := scopeTree.Root.Children[0].Lookup("A")
	assert.NotNil(t, classVar)
	dfgClassNode := graph[classVar]
	assert.NotNil(t, dfgClassNode)

	classDef := dfg.ClassDefs
	assert.NotNil(t, classDef)
	assert.NotNil(t, classDef[classVar])
	assert.Greater(t, len(classDef[classVar].Methods), 0)
	assert.Greater(t, len(classDef[classVar].Properties), 0)
}