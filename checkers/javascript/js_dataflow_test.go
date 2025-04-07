//globstar:registry-exclude
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

	t.Run("variable_data_flow", func(t *testing.T) {
		source := `
			var a = 10
			function f(x){
				a = x * 2	
				return a;
			}

			f(a)
			console.log(a)
			`
		parseResult := parseJsCode(t, []byte(source))

		pass := &ana.Pass{
			Analyzer:    DataFlowAnalyzer,
			FileContext: parseResult,
		}

		_, err := createDataFlowGraph(pass)
		assert.NoError(t, err)

	})
}
