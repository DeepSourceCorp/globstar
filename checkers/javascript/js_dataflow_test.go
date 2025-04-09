package javascript

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ana "globstar.dev/analysis"
)

func parseJsCode(t *testing.T, source []byte) *ana.ParseResult {
	pass, err := ana.Parse("", source, ana.LangJs, ana.LangJs.Grammar())
	require.NoError(t, err)

	return pass
}

func TestDataFlowAnalysis(t *testing.T) {

	// TODO:
	// - Update test-cases to handle dataflow implementation instead of sql-injection cases.
	// - Add other edge-cases including function blocks, Class Blocks etc.

	// t.Run("variable_data_flow", func(t *testing.T) {
	// 	source := `
	// 		var a = 10
	// 		function f(x){
	// 			a = x * 2
	// 			return a;
	// 		}

	// 		f(a)
	// 		console.log(a)
	// 		`
	// 	parseResult := parseJsCode(t, []byte(source))

	// 	pass := &ana.Pass{
	// 		Analyzer:    DataFlowAnalyzer,
	// 		FileContext: parseResult,
	// 	}

	// 	dfgStruct, err := createDataFlowGraph(pass)
	// 	assert.NoError(t, err)

	// 	dfg := dfgStruct.(*DataFlowGraph)

	// 	flowGraph := dfg.Graph
	// 	assert.NotNil(t, flowGraph)

	// 	scopeTree := dfg.ScopeTree
	// 	assert.NotNil(t, scopeTree)

	// 	aVar := scopeTree.Root.Children[0].Lookup("a")
	// 	assert.NotNil(t, aVar)

	// 	funcVar := scopeTree.Root.Children[0].Lookup("f")
	// 	assert.NotNil(t, funcVar)

	// 	funcScope := scopeTree.GetScope(funcVar.DeclNode.ChildByFieldName("body"))
	// 	assert.NotNil(t, funcScope)

	// 	aFuncVar := funcScope.Lookup("a")
	// 	assert.NotNil(t, aFuncVar)

	// 	// no re-assignment of a inside the function body.
	// 	assert.Equal(t, aVar, aFuncVar)

	// })

	// t.Run("variable_data_flow_through_multiple_functions", func(t *testing.T) {
	// 	source := `
	// 		var a = 10
	// 		function f(x){
	// 			a = x * 2
	// 			return a;
	// 		}

	// 		f(a)
	// 		console.log(a)

	// 		function g(x) {
	// 			var a = x * 2;
	// 			return a;
	// 		}

	// 		g(a)
	// 		console.log(a)
	// 		`
	// 	parseResult := parseJsCode(t, []byte(source))

	// 	pass := &ana.Pass{
	// 		Analyzer:    DataFlowAnalyzer,
	// 		FileContext: parseResult,
	// 	}

	// 	dfgStruct, err := createDataFlowGraph(pass)
	// 	assert.NoError(t, err)

	// 	dfg := dfgStruct.(*DataFlowGraph)

	// 	flowGraph := dfg.Graph
	// 	assert.NotNil(t, flowGraph)

	// 	scopeTree := dfg.ScopeTree
	// 	assert.NotNil(t, scopeTree)

	// 	aVar := scopeTree.Root.Children[0].Lookup("a")
	// 	assert.NotNil(t, aVar)

	// 	f1Var := scopeTree.Root.Children[0].Lookup("f")
	// 	assert.NotNil(t, f1Var)

	// 	f1Scope := scopeTree.GetScope(f1Var.DeclNode.ChildByFieldName("body"))
	// 	assert.NotNil(t, f1Scope)

	// 	aF1Var := f1Scope.Lookup("a")
	// 	assert.NotNil(t, aF1Var)

	// 	// no re-assignment of a inside the function body.
	// 	assert.Equal(t, aVar, aF1Var)

	// 	f2Var := scopeTree.Root.Children[0].Lookup("g")
	// 	assert.NotNil(t, f2Var)

	// 	f2Scope := scopeTree.GetScope(f2Var.DeclNode.ChildByFieldName("body"))
	// 	assert.NotNil(t, f2Scope)

	// 	aF2Var := f2Scope.Lookup("a")
	// 	assert.NotNil(t, aF2Var)

	// 	// reassignment of `a` inside of g(x) causes it to be a different variable inside aF2Var
	// 	assert.NotEqual(t, aVar, aF2Var)

	// })

	t.Run("variable_assignment_data_flow", func(t *testing.T) {
		// Taint Logic not implemented after refactoring the data_flow_analyzer.
		source := `
		x = 10
		function f(x){
			// do something	
		}
			
		f(x)
		`

		parseResult := parseJsCode(t, []byte(source))
		pass := &ana.Pass{
			Analyzer:    DataFlowAnalyzer,
			FileContext: parseResult,
		}
		dfgStruct, err := createDataFlowGraph(pass)
		assert.NoError(t, err)
		dfg := dfgStruct.(*DataFlowGraph)
		flowGraph := dfg.Graph
		assert.NotNil(t, flowGraph)
		scopeTree := dfg.ScopeTree
		assert.NotNil(t, scopeTree)

		t.Log(scopeTree.Root.Children[0].Variables)
		t.Logf("+++++++%v++++++++\n", dfg.Graph)

		// for variable, node := range dfg.Graph {
		// 	t.Logf("Variable in graph: %s, Kind: %v\n", variable.Name, variable.Kind)
		// 	t.Logf("Node details: %+v\n", node)
		// }
		// functionScope := scopeTree.GetScope(functionVar.DeclNode)
		// assert.NotNil(t, functionScope)
		// functionDef := flowGraph[functionVar].FuncDef
		// assert.NotNil(t, functionDef)
		// assert.NotNil(t, functionDef.Body)
	})

}
