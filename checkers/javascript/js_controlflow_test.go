package javascript

import (
	"testing"

	"github.com/stretchr/testify/require"
	ana "globstar.dev/analysis"
)

func TestControlFlowGraph(t *testing.T) {
	source := `
		let a = 21;
		let b = 22;
		f(a);
		function f(x) {
			let y = x + 1;
			return y;
		}

		f(a);
	`
	parseResult := ParseJsCode(t, []byte(source))
	pass := &ana.Pass{
		Analyzer:    ControlFlowAnalyzer,
		FileContext: parseResult,
	}

	_, err := createControlFlowGraph(pass)
	require.NoError(t, err)
}
